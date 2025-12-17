<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue';
import { XMarkIcon, CalendarIcon } from '@heroicons/vue/24/outline';
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/vue/20/solid';
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot, Switch, SwitchGroup, SwitchLabel, Combobox, ComboboxButton, ComboboxInput, ComboboxLabel, ComboboxOption, ComboboxOptions } from '@headlessui/vue';
import type { TimesheetEntry } from '../../types/Timesheet';
import type { BillingCode as BillingCodeAPI } from '../../types/BillingCode';
import { createEmptyTimesheetEntry } from '../../types/Timesheet';
import { getEntries, getTimesheetActiveBillingCodes, getUsers, createEntry, updateEntry, deleteEntry as deleteEntryAPI } from '../../api';
import TimesheetEntryComponent from '../../components/timesheet/TimesheetEntry.vue';
import CalendarEventPicker from '../../components/timesheet/CalendarEventPicker.vue';
import googleCalendarAPI, { type CalendarEvent } from '../../api/googleCalendar';
import { formatDate } from '../../utils/dateUtils';

// DOM refs for calendar positioning
const container = ref<HTMLElement | null>(null);
const containerNav = ref<HTMLElement | null>(null);
const containerOffset = ref<HTMLElement | null>(null);

// State
const weeklyEntries = ref<TimesheetEntry[]>([]);
const isLoading = ref(true);
const error = ref<string | null>(null);
const showModal = ref(false);
const currentDate = ref(new Date());
const currentWeek = ref<Date[]>([]);
const billingCodes = ref<any[]>([]);
const users = ref<any[]>([]);
const showImpersonation = ref(false);
const isSaving = ref(false); // Track saving state

// View as another user (admin only)
const viewingAsUserId = ref<number | null>(null);
const currentUser = ref<any>(null);

// Calendar import state
const showCalendarPicker = ref(false);
const calendarConnected = ref(false);
const isCheckingAuth = ref(false);

// Drag state
const isDragging = ref(false);
const dragDay = ref<Date | null>(null);
const dragStartHour = ref<number | null>(null);
const dragEndHour = ref<number | null>(null);
const dragColumnIndex = ref<number | null>(null);

// Modal form state
const formEntry = ref<TimesheetEntry>(createEmptyTimesheetEntry());

// Hash a string to get a consistent color
const hashCodeToColor = (str: string): string => {
  if (!str || typeof str !== 'string') return '#e2e8f0'; // Default gray for empty strings
  
  // Take only the first part of the code (before underscore)
  const prefix = str.split('_')[0];
  
  // Simple hash function to get a number
  let hash = 0;
  for (let i = 0; i < prefix.length; i++) {
    hash = prefix.charCodeAt(i) + ((hash << 5) - hash);
  }
  
  // Convert to color with good saturation and lightness
  const h = Math.abs(hash) % 360; // Hue (0-360)
  const s = 65 + (Math.abs(hash) % 20); // Saturation (65-85%)
  const l = 50 + (Math.abs(hash) % 10); // Lightness (50-60%)
  
  return `hsl(${h}, ${s}%, ${l}%)`;
};

// Get text color (white/black) based on background color's perceived brightness
const getTextColor = (bgColor: string): string => {
  // For HSL colors, a simple rule: if lightness < 60%, use white, otherwise black
  const match = bgColor.match(/hsl\(\d+,\s*\d+%,\s*(\d+)%\)/);
  
  if (match && match[1]) {
    const lightness = parseInt(match[1], 10);
    return lightness < 60 ? 'text-white' : 'text-gray-900';
  }
  
  // Default to black if can't determine
  return 'text-gray-900';
};

// Additional state for combobox
const billingCodeQuery = ref('');

// Watch for impersonation toggle changes
watch(showImpersonation, (newValue) => {
  if (!newValue) {
    // Reset impersonation data when toggled off
    formEntry.value.impersonate_as_user_id = 0;
    formEntry.value.is_being_impersonated = false;
  }
});

// Filter billing codes based on search query
const filteredBillingCodes = computed(() => {
  if (!billingCodes.value.length) {
    return [];
  }
  
  const query = billingCodeQuery.value.toLowerCase();
  return billingCodes.value.filter((code: any) => {
    return code.name.toLowerCase().includes(query) || 
           (code.code && code.code.toLowerCase().includes(query));
  });
});

// Get the selected billing code
const selectedBillingCode = computed(() => {
  if (!formEntry.value.billing_code_id || !billingCodes.value.length) {
    return null;
  }
  
  return billingCodes.value.find((code: any) => 
    Number(code.id) === Number(formEntry.value.billing_code_id)
  ) || null;
});

// Add a watch for the billing code
watch(selectedBillingCode, (newValue) => {
  if (newValue) {
    // Update the billing_code_id when a billing code is selected
    formEntry.value.billing_code_id = Number(newValue.id);
    console.log('Selected billing code:', newValue.code, 'with ID:', formEntry.value.billing_code_id);
  }
});

// Function to display billing code in the input
const displayBillingCode = (code: any): string => {
  if (!code) return '';
  return code.name;
};

// Initialize component
onMounted(async () => {
  initializeCurrentWeek();
  await fetchWeeklyEntries();
  await fetchBillingCodes();
  await fetchUsers();
  await checkCalendarAuth();
  
  // Set the container scroll position based on the current time
  setTimeout(() => {
    if (container.value && containerNav.value && containerOffset.value) {
      const currentHour = new Date().getHours();
      
      // Only scroll if the current time is within our display range (8am-10pm)
      if (currentHour >= 8 && currentHour <= 22) {
        // Calculate relative position within our time range (8am-10pm)
        // Each hour has 2 rows (for 30-min increments), and we skip the header row
        const targetRow = ((currentHour - 8) * 2) + 1;
        // Each row is 2rem tall (32px)
        const scrollPosition = targetRow * 32; // 2rem = 32px
        
        container.value.scrollTop = scrollPosition;
      }
    }
  }, 100); // Small delay to ensure the DOM is fully rendered
});

// Initialize the current week dates
const initializeCurrentWeek = () => {
  // Create UTC date objects to avoid timezone shifts
  const now = new Date(currentDate.value);
  
  // Find the day of week in local time (0 = Sunday, 6 = Saturday)
  const dayOfWeek = now.getDay();
  
  // Find the start of the week (Sunday) by calculating the date difference
  const startDate = new Date(now);
  startDate.setDate(now.getDate() - dayOfWeek);
  
  // Reset the time component to midnight to avoid time-of-day comparison issues
  startDate.setHours(0, 0, 0, 0);
  
  // Create array of dates for the week
  const weekDates: Date[] = [];
  for (let i = 0; i < 7; i++) {
    const date = new Date(startDate);
    date.setDate(startDate.getDate() + i);
    // Ensure each date starts at midnight
    date.setHours(0, 0, 0, 0);
    weekDates.push(date);
  }
  
  currentWeek.value = weekDates;
};

// Navigate to previous week
const previousWeek = () => {
  const firstDay = new Date(currentWeek.value[0]);
  // Ensure consistent date math by setting to midnight
  firstDay.setHours(0, 0, 0, 0);
  firstDay.setDate(firstDay.getDate() - 7);
  currentDate.value = firstDay;
  initializeCurrentWeek();
  fetchWeeklyEntries();
};

// Navigate to next week
const nextWeek = () => {
  const firstDay = new Date(currentWeek.value[0]);
  // Ensure consistent date math by setting to midnight
  firstDay.setHours(0, 0, 0, 0);
  firstDay.setDate(firstDay.getDate() + 7);
  currentDate.value = firstDay;
  initializeCurrentWeek();
  fetchWeeklyEntries();
};

// Go to today
const goToToday = () => {
  const today = new Date();
  // Ensure consistent date by setting to midnight
  today.setHours(0, 0, 0, 0);
  currentDate.value = today;
  initializeCurrentWeek();
  fetchWeeklyEntries();
};

// Fetch weekly timesheet entries
const fetchWeeklyEntries = async () => {
  isLoading.value = true;
  error.value = null;
  
  try {
    // Get start and end dates of current week
    const weekStartISO = currentWeek.value[0].toISOString().split('T')[0];
    const weekEndISO = currentWeek.value[6].toISOString().split('T')[0];
    
    // Fetch entries with optional user_id parameter for viewing as another user
    const entries = await getEntries(weekStartISO, weekEndISO, viewingAsUserId.value);
    weeklyEntries.value = entries;
    
  } catch (err) {
    console.error('Error fetching timesheet entries:', err);
    error.value = 'Failed to load timesheet entries. Please try again.';
  } finally {
    isLoading.value = false;
  }
};

// Fetch billing codes
const fetchBillingCodes = async () => {
  try {
    const codes = await getTimesheetActiveBillingCodes();
    
    // Convert API billing codes to our required format
    const formattedBillingCodes = codes
      .filter((code: BillingCodeAPI) => {
        // First get ID from one of the possible sources
        const rawId = code.ID;
        const numericId = Number(rawId);
        
        // Verify it's a valid number
        if (isNaN(numericId) || numericId <= 0) {
          console.warn('Skipping billing code with invalid ID:', code);
          return false;
        }
        
        return true;
      })
      .map((code: BillingCodeAPI) => {
        // Convert to timesheet billing code format
        return {
          id: Number(code.ID),
          billing_code_id: Number(code.ID),
          name: code.name || '',
          code: code.code || '',
          project_id: Number(code.project_id) || 0,
          rate: Number(code.rate?.amount) || 0
        };
      });
    
    billingCodes.value = formattedBillingCodes;
  } catch (err) {
    console.error('Error fetching billing codes:', err);
  }
};

// Fetch users for impersonation
const fetchUsers = async () => {
  try {
    // Fetch staff (employees) instead of users
    const staffList = await getUsers(); // This actually calls /api/staff
    // Ensure each staff member has an ID (Employee.ID)
    users.value = staffList.map(staff => ({
      id: Number(staff.ID), // Employee.ID
      first_name: staff.first_name || '',
      last_name: staff.last_name || '',
      email: staff.email || ''
    }));
    
    // Get current user from localStorage
    const userStr = localStorage.getItem('snowpack_user');
    if (userStr) {
      currentUser.value = JSON.parse(userStr);
    }
  } catch (err) {
    console.error('Error fetching staff for impersonation:', err);
  }
};

// Change which user's timesheet to view
const changeViewingUser = async (userId: number | null) => {
  viewingAsUserId.value = userId;
  await fetchWeeklyEntries();
};

// Get display name for a user
const getUserDisplayName = (userId: number | null): string => {
  if (!userId) return 'My Timesheet';
  
  const user = users.value.find(u => u.id === userId);
  if (!user) return 'Unknown User';
  
  return `${user.first_name} ${user.last_name}`.trim() || user.email || 'Unknown User';
};

// Open modal to create a new entry
const createNewEntry = (event: any, day: any) => {
  
  // Reset form values to defaults using the provided factory function
  formEntry.value = createEmptyTimesheetEntry();
  
  // Set the billing code ID to 0 (no selection)
  formEntry.value.billing_code_id = 0;
  
  // Create date strings from the provided day
  const dateStr = day.format ? day.format('YYYY-MM-DD') : new Date(day).toISOString().split('T')[0];
  
  // Set start and end times (one hour apart) on the selected day
  const now = new Date();
  const hours = now.getHours();
  const minutes = now.getMinutes();
  
  const startTime = new Date(dateStr);
  startTime.setHours(hours, minutes, 0, 0);
  
  const endTime = new Date(dateStr);
  endTime.setHours(hours + 1, minutes, 0, 0);
  
  formEntry.value.start = startTime.toISOString();
  formEntry.value.end = endTime.toISOString();
  
  
  showModal.value = true;
  showImpersonation.value = false; // Default to not showing impersonation UI
};

// Drag functionality for calendar
const startDrag = (day: Date, hour: number, event: MouseEvent) => {
  // Get the column index directly from the element with more robust error handling
  const target = event.currentTarget as HTMLElement;
  
  if (!target) {
    console.error('No target element found for drag start');
    return;
  }
  
  // Get and verify both attributes
  const dayAttr = target.getAttribute('data-day');
  const hourAttr = target.getAttribute('data-hour');
  
  // Safe parsing with fallbacks
  const columnIndex = parseInt(dayAttr || '-1', 10);
  const hourValue = parseFloat(hourAttr || '-1');
  
  // Verify we have a valid column index before proceeding
  if (columnIndex === -1) {
    console.error('Invalid column index, cannot start drag');
    return;
  }
  
  // Initialize drag state - use the attribute values directly
  isDragging.value = true;
  dragDay.value = new Date(day);
  dragStartHour.value = hourValue; // Use the attribute value
  dragEndHour.value = hourValue;   // Use the attribute value
  dragColumnIndex.value = columnIndex;
  
  // Prevent text selection during drag
  document.body.style.userSelect = 'none';
  
  // Highlight the starting cell
  highlightDragRange();
  
  // Add mousemove handler for tracking drag within the column
  document.addEventListener('mousemove', handleMouseMove);
  
  // Stop event propagation
  event.stopPropagation();
  event.preventDefault();
};

// Handle mouse movement during drag
const handleMouseMove = (event: MouseEvent) => {
  if (!isDragging.value || dragColumnIndex.value === null) return;
  
  // Get the elements under the cursor
  const elementsUnderCursor = document.elementsFromPoint(event.clientX, event.clientY);
  
  // Find calendar cells - with explicit class check to ensure we get actual cells
  const calendarCells = elementsUnderCursor.filter(el => 
    el.classList.contains('calendar-cell')
  );
  
  
  if (calendarCells.length === 0) return;
  
  // Focus on the first calendar cell (should be the top-most one)
  const cell = calendarCells[0] as HTMLElement;
  
  // Get cell attributes with explicit parsing to avoid any issues
  const cellDayAttr = cell.getAttribute('data-day');
  const cellHourAttr = cell.getAttribute('data-hour');
  
  // Skip if we couldn't get valid attributes
  if (cellDayAttr === null || cellHourAttr === null) {
    console.warn('Cell is missing data attributes');
    return;
  }
  
  // Parse with safety checks
  const cellDay = parseInt(cellDayAttr, 10);
  const cellHour = parseFloat(cellHourAttr);
  
  if (isNaN(cellDay) || isNaN(cellHour)) {
    console.warn('Invalid data attributes:', { cellDayAttr, cellHourAttr });
    return;
  }
  
  // Only process if this cell is in the column where we started the drag
  if (cellDay === dragColumnIndex.value) {
    dragEndHour.value = cellHour;
    highlightDragRange();
  }
};


const endDrag = (event?: MouseEvent) => {
  if (!isDragging.value || !dragDay.value || dragStartHour.value === null || 
      dragEndHour.value === null || dragColumnIndex.value === null) return;
  
  if (event) {
    event.preventDefault();
    event.stopPropagation();
  }
  
  
  // Remove the mousemove handler
  document.removeEventListener('mousemove', handleMouseMove);
  
  // Sort hours to ensure start <= end
  const startHour = Math.min(dragStartHour.value, dragEndHour.value);
  const endHour = Math.max(dragStartHour.value, dragEndHour.value);
  
  // Create entry from drag selection
  formEntry.value = {
    ...createEmptyTimesheetEntry(),
    billing_code_id: 0 // Explicitly set to numeric zero
  };
  
  
  // Use the day from the stored column index for consistency
  const day = currentWeek.value[dragColumnIndex.value];
  if (!day) {
    console.error('Could not find day for column index', dragColumnIndex.value);
    return;
  }
  
  
  // Get the date part in local timezone format
  const dateStr = day.toISOString().split('T')[0];
  
  // Create start time string - format: "2023-01-01T08:00:00"
  // Construct time string directly to avoid timezone conversion issues
  const startHourInt = Math.floor(startHour);
  // Support 15-minute increments: 0, 0.25, 0.5, 0.75
  const startMinInt = Math.round((startHour % 1) * 60);
  
  // Format hours and minutes with leading zeros
  const startHour_str = startHourInt.toString().padStart(2, '0');
  const startMin_str = startMinInt.toString().padStart(2, '0');
  
  // Calculate end time
  const endHourInt = Math.floor(endHour);
  const endMinInt = Math.round((endHour % 1) * 60);
  
  // Add 30 minutes to end time to make it exclusive (default block size)
  let finalEndHourInt = endHourInt;
  let finalEndMinInt = endMinInt + 30;
  
  // Handle minute overflow
  if (finalEndMinInt >= 60) {
    finalEndHourInt += 1;
    finalEndMinInt -= 60;
  }
  
  // Format end hours and minutes
  const endHour_str = finalEndHourInt.toString().padStart(2, '0');
  const endMin_str = finalEndMinInt.toString().padStart(2, '0');
  
  // Create direct ISO strings with the correct time (no timezone conversion)
  formEntry.value.start = `${dateStr}T${startHour_str}:${startMin_str}:00.000Z`;
  formEntry.value.end = `${dateStr}T${endHour_str}:${endMin_str}:00.000Z`;
  
  // Reset drag state
  isDragging.value = false;
  removeHighlighting();
  document.body.style.userSelect = '';
  
  // Open modal
  showModal.value = true;
};

const removeHighlighting = () => {
  // Remove highlight classes from all calendar cells
  const cells = document.querySelectorAll('.calendar-cell');
  cells.forEach(cell => {
    cell.classList.remove('bg-sage-pale');
  });
};

// Highlights all cells in the selected range
const highlightDragRange = () => {
  if (!isDragging.value || dragStartHour.value === null || 
      dragEndHour.value === null || dragColumnIndex.value === null) return;
  
  // First, clear all highlighting
  removeHighlighting();
  
  // Get the min and max hours to ensure we highlight in either direction (up or down)
  const startHour = Math.min(dragStartHour.value, dragEndHour.value);
  const endHour = Math.max(dragStartHour.value, dragEndHour.value);
  
  
  // Get all calendar cells with an explicit selector
  const allCells = document.querySelectorAll('.calendar-cell');
  let highlightedCount = 0;
  
  // Now explicitly check each cell's data attributes
  allCells.forEach(cell => {
    // Skip elements without proper data attributes
    const cellDayAttr = cell.getAttribute('data-day');
    const cellHourAttr = cell.getAttribute('data-hour');
    
    if (cellDayAttr === null || cellHourAttr === null) return;
    
    // Parse with safety checks
    const cellDay = parseInt(cellDayAttr, 10);
    const cellHour = parseFloat(cellHourAttr);
    
    if (isNaN(cellDay) || isNaN(cellHour)) return;
    
    // Check if this cell is in our column and time range
    if (cellDay === dragColumnIndex.value && cellHour >= startHour && cellHour <= endHour) {
      cell.classList.add('bg-sage-pale');
      // Force a reflow to ensure the style is applied
      void (cell as HTMLElement).offsetHeight;
      highlightedCount++;
      console.debug(`Highlighted cell: day=${cellDay}, hour=${cellHour}`);
    }
  });
  
};

// Setup mouse handlers for drag end detection
onMounted(() => {
  document.addEventListener('mouseup', endDrag);
  
  // Add mouseleave handler to end drag if mouse leaves the calendar
  if (container.value) {
    container.value.addEventListener('mouseleave', endDrag);
  }
  
  return () => {
    document.removeEventListener('mouseup', endDrag);
    document.removeEventListener('mousemove', handleMouseMove);
    if (container.value) {
      container.value.removeEventListener('mouseleave', endDrag);
    }
  };
});

// Open modal to edit an existing entry
const editEntry = (entry: TimesheetEntry) => {
  // Create a deep copy to avoid reference issues
  formEntry.value = JSON.parse(JSON.stringify(entry));
  
  // Ensure billing_code_id is a proper number
  formEntry.value.billing_code_id = Number(formEntry.value.billing_code_id || 0);
  
  // If this entry was REJECTED, reset it to DRAFT and show warning
  if (formEntry.value.state === 'ENTRY_STATE_REJECTED') {
    formEntry.value.state = 'ENTRY_STATE_DRAFT';
  }
  
  // Make sure user can interact immediately with the form
  showModal.value = true;
  
  // If the entry has impersonation, show the impersonation UI
  if (entry.impersonate_as_user_id && entry.impersonate_as_user_id > 0) {
    showImpersonation.value = true;
    // Make sure impersonate_as_user_id is explicitly set as a number
    formEntry.value.impersonate_as_user_id = Number(entry.impersonate_as_user_id);
  } else {
    showImpersonation.value = false;
    formEntry.value.impersonate_as_user_id = null;
  }
};

// Close the modal
const closeModal = () => {
  showModal.value = false;
  formEntry.value = createEmptyTimesheetEntry();
  showImpersonation.value = false; // Reset impersonation flag when closing
};

// Save an entry (create or update)
const saveEntry = async () => {
  error.value = null; // Clear previous errors
  isSaving.value = true; // Start saving process
  
  try {
    // Ensure billing_code_id is a number and has a value
    const billingCodeId = Number(formEntry.value.billing_code_id);
    if (isNaN(billingCodeId) || billingCodeId <= 0) {
      error.value = 'Please select a valid billing code';
      isSaving.value = false; // Reset saving state
      return;
    }
    
    // Update the value as a number in the form
    formEntry.value.billing_code_id = billingCodeId;
    
    // Validate required fields
    if (!formEntry.value.start || !formEntry.value.end) {
      error.value = 'Please select both start and end times';
      isSaving.value = false; // Reset saving state
      return;
    }
    
    // Validate times - end must be after start
    const startDate = new Date(formEntry.value.start);
    const endDate = new Date(formEntry.value.end);
    if (endDate <= startDate) {
      error.value = 'End time must be after start time';
      isSaving.value = false; // Reset saving state
      return;
    }
    
    // Validate that notes are provided
    if (!formEntry.value.notes || formEntry.value.notes.trim() === '') {
      error.value = 'Please enter notes for this timesheet entry';
      isSaving.value = false; // Reset saving state
      return;
    }
    
    // Ensure impersonation fields are properly set
    if (showImpersonation.value && formEntry.value.impersonate_as_user_id) {
      formEntry.value.is_being_impersonated = true;
      // Ensure impersonate_as_user_id is a number
      formEntry.value.impersonate_as_user_id = Number(formEntry.value.impersonate_as_user_id);
    } else {
      // When not using impersonation, explicitly set to null
      formEntry.value.impersonate_as_user_id = null;
      formEntry.value.is_being_impersonated = false;
    }
    
    // Make sure notes is always a string
    formEntry.value.notes = formEntry.value.notes.trim();
    
    // Create or update entry
    if (formEntry.value.entry_id === 0) {
      await createEntry(formEntry.value);
      // Successfully created - close modal and refresh
      closeModal();
      await fetchWeeklyEntries();
    } else {
      try {
        // Convert entry_id to a number to ensure proper API call
        formEntry.value.entry_id = Number(formEntry.value.entry_id);
        
        // Validate entry_id
        if (isNaN(formEntry.value.entry_id) || formEntry.value.entry_id <= 0) {
          throw new Error(`Invalid entry ID: ${formEntry.value.entry_id}`);
        }
        
        // If this entry was REJECTED, reset it to DRAFT when saving
        if (formEntry.value.state === 'ENTRY_STATE_REJECTED') {
          formEntry.value.state = 'ENTRY_STATE_DRAFT';
        }
        
        await updateEntry(formEntry.value);
        
        // Successfully updated - close modal and refresh
        closeModal();
        await fetchWeeklyEntries();
      } catch (updateErr: any) {
        console.error('Error updating timesheet entry:', updateErr);
        
        // Extract error message for better user feedback
        let errorMessage = 'Failed to update entry. Please try again.';
        
        if (updateErr.message) {
          errorMessage = updateErr.message;
        } else if (updateErr.response) {
          if (updateErr.response.status === 404) {
            errorMessage = `Failed to update entry #${formEntry.value.entry_id}. The entry may have been deleted or you don't have permission to edit it.`;
          } else if (updateErr.response.status === 409) {
            // Special handling for entries that can't be edited due to their state
            const state = updateErr.response.data?.state || 'non-draft';
            errorMessage = `This entry cannot be modified because it is in ${state} state. Only entries in DRAFT state can be edited.`;
          } else if (updateErr.response.data && updateErr.response.data.error) {
            errorMessage = updateErr.response.data.error;
          }
        }
        
        error.value = errorMessage;
        // Don't close the modal so user can see the error
      }
    }
  } catch (err: any) {
    console.error('Error saving timesheet entry:', err);
    
    // Extract error message for better user feedback
    if (err.message) {
      error.value = err.message;
    } else if (err.response && err.response.data && err.response.data.error) {
      error.value = err.response.data.error;
    } else {
      error.value = 'Failed to save entry. Please try again.';
    }
  } finally {
    isSaving.value = false; // End saving process regardless of outcome
  }
};

// Delete an entry
const deleteEntry = async () => {
  if (!formEntry.value || formEntry.value.entry_id === 0) return;
  
  error.value = null; // Clear previous errors
  isSaving.value = true; // Start saving/deleting process
  
  try {
    await deleteEntryAPI(formEntry.value.entry_id);
    
    // Refresh entries after delete
    await fetchWeeklyEntries();
    closeModal();
  } catch (err: any) {
    console.error('Error deleting timesheet entry:', err);
    
    // Extract error message for better user feedback
    if (err.response && err.response.data && err.response.data.error) {
      error.value = err.response.data.error;
    } else {
      error.value = 'Failed to delete entry. Please try again.';
    }
  } finally {
    isSaving.value = false; // End saving/deleting process regardless of outcome
  }
};

// Check Google Calendar connection status
const checkCalendarAuth = async () => {
  isCheckingAuth.value = true;
  try {
    const status = await googleCalendarAPI.getAuthStatus();
    calendarConnected.value = status.connected;
  } catch (err) {
    console.error('Failed to check calendar auth:', err);
    calendarConnected.value = false;
  } finally {
    isCheckingAuth.value = false;
  }
};

// Open calendar import flow
const openCalendarImport = async () => {
  if (!calendarConnected.value) {
    try {
      await googleCalendarAPI.authorize();
      calendarConnected.value = true;
    } catch (err: any) {
      if (err.message !== 'Authorization window was closed') {
        error.value = 'Failed to connect Google Calendar. Please try again.';
      }
      return;
    }
  }
  
  showCalendarPicker.value = true;
};

// Re-authorize Google Calendar (disconnect and reconnect)
const reauthorizeCalendar = async () => {
  try {
    // Disconnect old tokens
    await googleCalendarAPI.disconnect();
    calendarConnected.value = false;
    
    // Immediately start new authorization
    await googleCalendarAPI.authorize();
    calendarConnected.value = true;
    
    // Show success message
    console.log('Successfully re-authorized Google Calendar');
  } catch (err: any) {
    if (err.message !== 'Authorization window was closed') {
      error.value = 'Failed to re-authorize Google Calendar. Please try again.';
    }
    console.error('Failed to re-authorize calendar:', err);
  }
};

// Handle calendar event selection  
const handleCalendarEventSelect = (event: CalendarEvent) => {
  showCalendarPicker.value = false;
  
  // Extract date and time from RFC3339, keeping the time value as-is
  // RFC3339: "2025-09-30T09:30:00-04:00" -> "2025-09-30T09:30:00.000Z"
  // The timesheet system treats times as "naive" (ignores the Z), so we keep the original time value
  const extractTime = (rfc3339: string) => {
    const dateTimePart = rfc3339.substring(0, 19); // "2025-09-30T09:30:00"
    return dateTimePart + '.000Z';
  };
  
  formEntry.value.start = extractTime(event.start);
  formEntry.value.end = extractTime(event.end);
  formEntry.value.notes = event.summary + (event.description ? '\n\n' + event.description : '');
  formEntry.value.is_meeting = true;
  
  // Ensure the entry modal stays open
  if (!showModal.value) {
    showModal.value = true;
  }
};

// Close calendar picker
const closeCalendarPicker = () => {
  showCalendarPicker.value = false;
};

// Computed week dates for calendar picker
const weekStartDate = computed(() => {
  if (!currentWeek.value.length) return '';
  return currentWeek.value[0].toISOString().split('T')[0];
});

const weekEndDate = computed(() => {
  if (!currentWeek.value.length) return '';
  return currentWeek.value[6].toISOString().split('T')[0];
});

// Computed properties
const monthAndYear = computed(() => {
  if (!currentWeek.value.length) return '';
  const startDate = currentWeek.value[0];
  if (!startDate) return '';
  return startDate.toLocaleDateString('en-US', { month: 'long', year: 'numeric' });
});

const weekRangeText = computed(() => {
  if (!currentWeek.value.length) return '';
  
  const startDate = currentWeek.value[0];
  const endDate = currentWeek.value[6];
  
  if (!startDate || !endDate) return '';
  
  // Format dates using our utility function
  const startDateFormatted = formatDate(startDate, 'short');
  const endDateFormatted = formatDate(endDate, 'short');
  
  return `${startDateFormatted} - ${endDateFormatted}`;
});

const dayHeaders = computed(() => {
  if (!currentWeek.value.length) return [];
  
  return currentWeek.value.map(date => {
    if (!date) return { dayName: '', dayNum: '', date: new Date(), isToday: false };
    
    // Format day name using utility function for consistency
    const dayName = date.toLocaleDateString('en-US', { weekday: 'short', timeZone: 'UTC' });
    const dayNum = date.getDate();
    
    // Compare date strings to determine if this is today
    const isToday = isDateToday(date);
    
    return { dayName, dayNum, date, isToday };
  });
});

// Helper functions
const isDateToday = (date: Date) => {
  // Compare dates as YYYY-MM-DD strings to avoid timezone issues
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  
  // Format both to YYYY-MM-DD format and compare
  const todayStr = today.toISOString().split('T')[0];
  const dateStr = date.toISOString().split('T')[0];
  
  return dateStr === todayStr;
};

// Get entries for a specific day
const getEntriesForDay = (day: Date) => {
  if (!day || !weeklyEntries.value) return [];
  
  // Format the day to YYYY-MM-DD for comparison (in UTC to match how we store dates)
  const dayDateStr = formatDate(day, 'input');
  
  // Filter entries that fall on this day
  return weeklyEntries.value.filter(entry => {
    if (!entry || !entry.start) return false;
    
    try {
      // Extract just the date part from the ISO string to avoid timezone issues
      const entryDatePart = entry.start.split('T')[0];
      
      // Compare the date strings directly - this avoids timezone conversion issues
      return entryDatePart === dayDateStr;
    } catch (e) {
      console.error('Error comparing entry date:', e);
      return false;
    }
  });
};

// Get total hours for the week
const totalWeeklyHours = computed(() => {
  if (!weeklyEntries.value || !currentWeek.value.length) return '0.0';
  
  return weeklyEntries.value.reduce((total, entry) => {
    if (!entry.start || !entry.duration_hours) return total;
    
    // Check if entry is in the current week
    const entryDate = new Date(entry.start);
    const isInCurrentWeek = currentWeek.value.some(date => 
      date && date.toDateString() === entryDate.toDateString()
    );
    
    return isInCurrentWeek ? total + entry.duration_hours : total;
  }, 0).toFixed(1);
});

// Calculate duration between two ISO timestamps
const getDuration = (start?: string, end?: string): string => {
  if (!start || !end) return '0.00';
  
  try {
    const startDate = new Date(start);
    const endDate = new Date(end);
    
    // Calculate difference in milliseconds
    const diffMs = endDate.getTime() - startDate.getTime();
    
    // Convert to hours with 2 decimal places
    const hours = diffMs / (1000 * 60 * 60);
    return hours.toFixed(2);
  } catch (error) {
    console.error('Error calculating duration:', error);
    return '0.00';
  }
};

// Update the updateDateTime function to format times correctly
const updateDateTime = (field: 'date' | 'start-time' | 'end-time', value: string) => {
  // Ensure form entry has start and end dates
  if (!formEntry.value.start) formEntry.value.start = new Date().toISOString();
  if (!formEntry.value.end) formEntry.value.end = new Date().toISOString();
  
  // Get current date and time values
  const startDate = formEntry.value.start.split('T')[0];
  const startTime = formEntry.value.start.split('T')[1].split(':').slice(0, 2).join(':');
  const endTime = formEntry.value.end.split('T')[1].split(':').slice(0, 2).join(':');
  
  // Common shared date for both start and end
  let sharedDate = startDate;
  
  if (field === 'date') {
    // When date changes, update both start and end dates
    sharedDate = value;
    
    // Create new ISO strings with updated date
    formEntry.value.start = `${sharedDate}T${startTime}:00.000Z`;
    formEntry.value.end = `${sharedDate}T${endTime}:00.000Z`;
  } 
  else if (field === 'start-time') {
    // When start time changes, update only start time
    formEntry.value.start = `${sharedDate}T${value}:00.000Z`;
  } 
  else if (field === 'end-time') {
    // When end time changes, update only end time
    formEntry.value.end = `${sharedDate}T${value}:00.000Z`;
  }
  
  // Log the updated values
};
</script>

<template>
  <div class="timesheet-view">
    <div class="flex justify-between items-center p-4 bg-sage-pale border-b border-gray-200 rounded-xl shadow-sm mb-2">
      <div class="flex items-center space-x-3">
        <div class="text-lg font-semibold text-gray-900">
          {{ monthAndYear }}
          <span class="ml-2 text-sm text-gray-500">
            {{ weekRangeText }}
          </span>
          <span class="ml-4 text-sm font-medium text-sage">
            Total: {{ totalWeeklyHours }} hours
          </span>
        </div>
        
        <!-- View As selector -->
        <div class="flex items-center gap-1 ml-1">
          <svg class="h-3.5 w-3.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
          </svg>
          <select 
            v-model="viewingAsUserId"
            @change="changeViewingUser(viewingAsUserId)"
            class="block rounded border-gray-300 bg-white text-gray-900 focus:border-sage focus:outline-none focus:ring-sage"
            style="padding: 0.25rem 2rem 0.25rem 0.5rem; font-size: 0.688rem; line-height: 1.2;"
          >
            <option :value="null">My Timesheet</option>
            <option 
              v-for="user in users" 
              :key="user.id" 
              :value="user.id"
            >
              {{ user.first_name }} {{ user.last_name }}
            </option>
          </select>
        </div>
      </div>
      
      <div class="flex space-x-2">
        <button 
          @click="previousWeek"
          class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
        >
          Previous
        </button>
        <button 
          @click="goToToday"
          class="inline-flex items-center rounded-md bg-sage px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-sage-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sage"
        >
          Today
        </button>
        <button 
          @click="nextWeek"
          class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
        >
          Next
        </button>
      </div>
    </div>
    
    <!-- Viewing As Banner -->
    <div v-if="viewingAsUserId" class="mb-2 bg-blue-50 border-l-4 border-blue-500 px-3 py-2 rounded-md">
      <div class="flex items-center justify-between">
        <div class="flex items-center">
          <svg class="h-4 w-4 text-blue-400 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
          </svg>
          <p class="text-xs text-blue-700">
            <span class="font-medium">Viewing timesheet for:</span> {{ getUserDisplayName(viewingAsUserId) }}
          </p>
        </div>
        <button 
          @click="changeViewingUser(null)"
          class="text-xs text-blue-600 hover:text-blue-800 font-medium"
        >
          Return to My Timesheet
        </button>
      </div>
    </div>
    
    <!-- Global error banner -->
    <div v-if="error" class="mb-4 bg-red-50 border-l-4 border-red-500 p-4 rounded-md">
      <div class="flex">
        <div class="flex-shrink-0">
          <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="ml-3">
          <p class="text-sm text-red-700">{{ error }}</p>
        </div>
        <div class="ml-auto pl-3">
          <div class="-mx-1.5 -my-1.5">
            <button 
              @click="error = null" 
              class="inline-flex rounded-md bg-red-50 p-1.5 text-red-500 hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-red-600 focus:ring-offset-2 focus:ring-offset-red-50"
            >
              <span class="sr-only">Dismiss</span>
              <XMarkIcon class="h-5 w-5" aria-hidden="true" />
            </button>
          </div>
        </div>
      </div>
    </div>
    
    <div v-if="isLoading" class="loading-indicator">
      <div class="animate-spin h-6 w-6 text-primary"></div>
      <span>Loading timesheet...</span>
    </div>
    
    <div v-else-if="error && !weeklyEntries.length" class="error-message">
      {{ error }}
    </div>
    
    <div v-else class="timesheet-container h-[calc(100vh-9rem)] rounded-xl overflow-hidden">
      <div class="flex h-full flex-col">
        <!-- Loading state -->
        <div v-if="isLoading" class="flex flex-col items-center justify-center p-10 bg-white flex-grow">
          <div class="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-indigo-600 mb-4"></div>
          <span class="text-gray-700">Loading timesheet...</span>
        </div>
        
        <!-- Error state -->
        <div v-else-if="error" class="flex flex-col items-center justify-center p-10 bg-white flex-grow">
          <i class="fas fa-exclamation-circle text-4xl text-red-600 mb-4"></i>
          <span class="text-gray-700 mb-2">{{ error }}</span>
          <button @click="fetchWeeklyEntries" class="mt-4 inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            Retry
          </button>
        </div>
        
        <!-- Calendar View -->
        <div v-else ref="container" class="isolate flex flex-auto flex-col overflow-auto bg-white rounded-xl">
          <div style="width: 165%" class="flex max-w-full flex-none flex-col sm:max-w-none md:max-w-full">
            <div ref="containerNav" class="sticky top-0 z-50 flex-none bg-white shadow ring-1 ring-black/5 sm:pr-8 rounded-t-xl">
              <!-- Mobile day headers -->
              <div class="grid grid-cols-7 text-sm/6 text-gray-500 sm:hidden">
                <button 
                  v-for="(header, index) in dayHeaders" 
                  :key="index" 
                  type="button" 
                  class="flex flex-col items-center pb-3 pt-2"
                  @click="createNewEntry(header.date, 12)" 
                >
                  {{ header.dayName.charAt(0) }}
                  <span 
                    :class="[
                      'mt-1 flex size-8 items-center justify-center font-semibold', 
                      header.isToday 
                        ? 'rounded-full bg-sage text-white' 
                        : 'text-gray-900'
                    ]"
                  >
                    {{ header.dayNum }}
                  </span>
                </button>
              </div>

              <!-- Desktop day headers -->
              <div class="-mr-px hidden grid-cols-7 divide-x divide-gray-100 border-r border-gray-100 text-sm/6 text-gray-500 sm:grid">
                <div class="col-end-1 w-14" />
                <div 
                  v-for="(header, index) in dayHeaders" 
                  :key="index" 
                  class="flex items-center justify-center py-3"
                  @click="createNewEntry(header.date, 12)"
                  style="cursor: pointer;"
                >
                  <span :class="{ 'flex items-baseline': header.isToday }">
                    {{ header.dayName }}
                    <span 
                      :class="[
                        header.isToday 
                          ? 'ml-1.5 flex size-8 items-center justify-center rounded-full bg-sage font-semibold text-white' 
                          : 'items-center justify-center font-semibold text-gray-900'
                      ]"
                    >
                      {{ header.dayNum }}
                    </span>
                  </span>
                </div>
              </div>
            </div>
            
            <div class="flex flex-auto">
              <!-- Time gutter -->
              <div class="sticky left-0 z-10 w-14 flex-none bg-white ring-1 ring-gray-100" />
              
              <!-- Calendar grid -->
              <div class="grid flex-auto grid-cols-1 grid-rows-1">
                <!-- Horizontal lines (time slots) -->
                <div 
                  class="col-start-1 col-end-2 row-start-1 grid divide-y divide-gray-100" 
                  style="grid-template-rows: repeat(30, minmax(2rem, 1fr))"
                >
                  <div ref="containerOffset" class="row-end-1 h-7" />
                  
                  <!-- Generate time slots (8am to 10pm, 30-minute increments) -->
                  <template v-for="hour in 15" :key="hour-1">
                    <!-- Hour marker (shows on the hour) -->
                    <div>
                      <div class="sticky left-0 z-20 -ml-14 -mt-2.5 w-14 pr-2 text-right text-xs/5 text-gray-400">
                        {{ (hour+7) < 12 ? `${hour+7}AM` : (hour+7 === 12 ? '12PM' : `${hour+7 - 12}PM`) }}
                      </div>
                    </div>
                    <!-- Half-hour increment (no text) -->
                    <div />
                  </template>
                </div>

                <!-- Vertical lines (days of week) -->
                <div class="col-start-1 col-end-2 row-start-1 hidden grid-cols-7 grid-rows-1 divide-x divide-gray-100 sm:grid sm:grid-cols-7">
                  <div v-for="i in 7" :key="i" :class="`col-start-${i} row-span-full`" />
                  <div class="col-start-8 row-span-full w-8" />
                </div>

                <!-- Events display -->
                <ol 
                  class="col-start-1 col-end-2 row-start-1 grid grid-cols-1 sm:grid-cols-7 sm:pr-8 relative" 
                  style="grid-template-rows: 1.75rem repeat(30, minmax(0, 1fr)) auto"
                >
                  <!-- First, generate complete calendar grid with explicit positioning to prevent indexing issues -->
                  <template v-for="hourIncrement in 30" :key="`row-${hourIncrement}`">
                    <template v-for="(day, dayIndex) in currentWeek" :key="`cell-${hourIncrement}-${dayIndex}`">
                      <div 
                        class="calendar-cell transition-colors duration-100"
                        :style="{
                          gridRow: `${hourIncrement + 1}`,
                          gridColumn: `${dayIndex + 1}`,
                          height: '2rem'
                        }"
                        :data-hour="8 + (hourIncrement - 1) / 2"
                        :data-day="dayIndex"
                        @mousedown="startDrag(day, 8 + (hourIncrement - 1) / 2, $event)"
                        style="cursor: pointer;"
                      >
                        <!-- Debug display showing both the data attributes and coordinates -->
                        <span v-if="false" class="text-xs text-gray-400 select-none opacity-30 hover:opacity-100">
                          {{ (8 + (hourIncrement - 1) / 2).toFixed(1) }}
                        </span>
                      </div>
                    </template>
                  </template>
                  
                  <!-- Timesheet entries with custom positioning layer to avoid affecting the grid -->
                  <template v-for="(day, dayIndex) in currentWeek" :key="`entries-${dayIndex}`">
                    <TimesheetEntryComponent
                      v-for="entry in getEntriesForDay(day)"
                      :key="`${entry.entry_id}-${dayIndex}`"
                      :entry="entry"
                      :day="day"
                      :dayIndex="dayIndex"
                      :timeRangeStart="8"
                      :timeRangeEnd="22"
                      @edit="editEntry"
                    />
                  </template>
                </ol>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Entry modal using Headless UI -->
        <TransitionRoot as="template" :show="showModal">
          <Dialog as="div" class="relative z-50" @close="closeModal">
            <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0" enter-to="opacity-100" leave="ease-in duration-200" leave-from="opacity-100" leave-to="opacity-0">
              <div class="fixed inset-0 bg-gray-500/75 transition-opacity" />
            </TransitionChild>

            <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
              <div class="flex min-h-full items-center justify-center p-4 text-center sm:p-0">
                <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95" enter-to="opacity-100 translate-y-0 sm:scale-100" leave="ease-in duration-200" leave-from="opacity-100 translate-y-0 sm:scale-100" leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95">
                  <DialogPanel class="relative transform overflow-hidden rounded-lg bg-white text-left shadow-xl transition-all sm:my-6 sm:w-full sm:max-w-md">
                    <div class="bg-sage text-white px-4 py-2 sm:px-4 sm:py-2 flex justify-between items-center">
                      <DialogTitle as="h3" class="text-base font-semibold leading-6">
                        {{ formEntry.entry_id ? 'Edit Timesheet Entry' : 'New Timesheet Entry' }}
                      </DialogTitle>
                      <button type="button" class="text-white hover:text-gray-200 focus:outline-none" @click="closeModal">
                        <span class="sr-only">Close</span>
                        <XMarkIcon class="h-5 w-5" aria-hidden="true" />
                      </button>
                    </div>
                    
                    <!-- REJECTED entry warning banner -->
                    <div v-if="formEntry.state === 'ENTRY_STATE_REJECTED'" class="bg-red-50 border-b-2 border-red-300 px-4 py-3">
                      <div class="flex items-center">
                        <svg class="h-5 w-5 text-red-600 mr-2" viewBox="0 0 20 20" fill="currentColor">
                          <path fill-rule="evenodd" d="M8.485 2.495c.673-1.167 2.357-1.167 3.03 0l6.28 10.875c.673 1.167-.17 2.625-1.516 2.625H3.72c-1.347 0-2.189-1.458-1.515-2.625L8.485 2.495zM10 5a.75.75 0 01.75.75v3.5a.75.75 0 01-1.5 0v-3.5A.75.75 0 0110 5zm0 9a1 1 0 100-2 1 1 0 000 2z" clip-rule="evenodd" />
                        </svg>
                        <div class="flex-1">
                          <p class="text-sm font-medium text-red-800">This entry was rejected</p>
                          <p class="text-xs text-red-700 mt-0.5">Make your changes and save to resubmit as DRAFT</p>
                        </div>
                      </div>
                    </div>
                    
                    <div class="bg-white px-3 py-4 sm:p-6">
                      <!-- Import from Calendar button -->
                      <div v-if="!formEntry.entry_id" class="mb-4 space-y-2">
                        <button
                          type="button"
                          @click="openCalendarImport"
                          class="w-full inline-flex items-center justify-center gap-2 rounded-md bg-blue-50 px-3 py-2 text-sm font-medium text-blue-700 hover:bg-blue-100 border border-blue-200"
                        >
                          <CalendarIcon class="h-4 w-4" />
                          Import from Google Calendar
                        </button>
                        <!-- Re-authorize button (shown when already connected) -->
                        <button
                          v-if="calendarConnected"
                          type="button"
                          @click="reauthorizeCalendar"
                          class="w-full inline-flex items-center justify-center gap-2 rounded-md bg-yellow-50 px-3 py-2 text-xs font-medium text-yellow-700 hover:bg-yellow-100 border border-yellow-200"
                        >
                          <svg class="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                          </svg>
                          Re-authorize Calendar
                        </button>
                      </div>
                      
                      <!-- Billing code selector with Combobox -->
                      <div class="py-1 border-b border-gray-100">
                        <Combobox 
                          v-model="selectedBillingCode" 
                          as="div"
                          @update:model-value="code => { 
                            if (code) { 
                              formEntry.billing_code_id = Number(code.id || code.billing_code_id); 
                              console.log('Selected code ID:', formEntry.billing_code_id, 'Code:', code.code);
                            } 
                          }"
                        >
                          <ComboboxLabel class="block text-xs font-medium text-gray-700">Billing Code</ComboboxLabel>
                          <div class="flex items-center gap-2 mt-0.5">
                            <div class="relative flex-1">
                              <ComboboxInput 
                                class="block w-full rounded border-gray-300 py-1 pl-2 pr-8 text-2xs focus:border-sage focus:outline-none focus:ring-sage"
                                @change="billingCodeQuery = $event.target.value"
                                :display-value="displayBillingCode"
                                placeholder="Select or search for a billing code..."
                              />
                              <ComboboxButton class="absolute inset-y-0 right-0 flex items-center pr-2">
                                <ChevronUpDownIcon class="h-4 w-4 text-gray-400" aria-hidden="true" />
                              </ComboboxButton>

                              <ComboboxOptions v-if="filteredBillingCodes.length > 0" class="absolute z-10 mt-1 max-h-48 w-full overflow-auto rounded-md bg-white py-1 text-2xs shadow-lg ring-1 ring-black/5 focus:outline-none">
                                <ComboboxOption v-for="code in filteredBillingCodes" :key="code.id" :value="code" as="template" v-slot="{ active, selected }">
                                  <li :class="['relative cursor-default select-none py-1.5 pl-2 pr-9 text-2xs', active ? 'bg-sage text-white' : 'text-gray-900']">
                                    <div class="flex items-center">
                                      <!-- Color-coded badge for the billing code -->
                                      <span 
                                        v-if="code.code"
                                        :style="{ backgroundColor: hashCodeToColor(code.code) }"
                                        :class="[
                                          'inline-flex items-center px-1.5 py-0.5 rounded text-[0.6rem] font-medium', 
                                          getTextColor(hashCodeToColor(code.code))
                                        ]"
                                      >
                                        {{ code.code }}
                                      </span>
                                      
                                      <span :class="['ml-2 truncate', selected && 'font-semibold']">
                                        {{ code.name }}
                                      </span>
                                    </div>

                                    <span v-if="selected" :class="['absolute inset-y-0 right-0 flex items-center pr-2', active ? 'text-white' : 'text-sage']">
                                      <CheckIcon class="h-4 w-4" aria-hidden="true" />
                                    </span>
                                  </li>
                                </ComboboxOption>
                              </ComboboxOptions>
                            </div>
                            
                            <!-- Show selected badge if a billing code is selected -->
                            <div v-if="selectedBillingCode && selectedBillingCode.code" class="flex-shrink-0">
                              <span 
                                :style="{ backgroundColor: hashCodeToColor(selectedBillingCode.code) }"
                                :class="[
                                  'inline-flex items-center px-1.5 py-0.5 rounded text-[0.65rem] font-medium', 
                                  getTextColor(hashCodeToColor(selectedBillingCode.code))
                                ]"
                              >
                                {{ selectedBillingCode.code }}
                              </span>
                            </div>
                          </div>
                        </Combobox>
                      </div>
                      
                      <!-- Time inputs - more compact -->
                      <div class="bg-sage-pale bg-opacity-30 rounded p-2 border border-sage border-opacity-20 mt-2">
                        <div class="grid grid-cols-3 gap-2 mb-2">
                          <!-- Date for both start and end -->
                          <div class="col-span-1">
                            <label class="block text-2xs font-medium text-gray-700">Date</label>
                            <input 
                              type="date" 
                              id="entry-date"
                              :value="formEntry.start ? formEntry.start.split('T')[0] : ''"
                              @input="e => updateDateTime('date', (e.target as HTMLInputElement).value)"
                              class="mt-0.5 bg-white border border-gray-300 text-gray-900 text-2xs rounded focus:ring-sage focus:border-sage block w-full p-1"
                              required 
                            />
                          </div>
                          
                          <!-- Start and end times -->
                          <div class="col-span-1">
                            <label class="block text-2xs font-medium text-gray-700">Start Time</label>
                            <input 
                              type="time" 
                              id="start-time"
                              :value="formEntry.start ? formEntry.start.split('T')[1].split(':').slice(0, 2).join(':') : ''"
                              @input="e => updateDateTime('start-time', (e.target as HTMLInputElement).value)"
                              class="mt-0.5 bg-white border border-gray-300 text-gray-900 text-2xs rounded focus:ring-sage focus:border-sage block w-full p-1"
                              required 
                            />
                          </div>
                          
                          <div class="col-span-1">
                            <label class="block text-2xs font-medium text-gray-700">End Time</label>
                            <input 
                              type="time" 
                              id="end-time"
                              :value="formEntry.end ? formEntry.end.split('T')[1].split(':').slice(0, 2).join(':') : ''"
                              @input="e => updateDateTime('end-time', (e.target as HTMLInputElement).value)"
                              class="mt-0.5 bg-white border border-gray-300 text-gray-900 text-2xs rounded focus:ring-sage focus:border-sage block w-full p-1"
                              required 
                            />
                          </div>
                        </div>
                        
                        <!-- Duration display -->
                        <div class="flex justify-end">
                          <span class="text-2xs text-gray-600 font-medium bg-white px-2 py-0.5 rounded-full border border-gray-200">
                            Duration: {{ getDuration(formEntry.start, formEntry.end) }} hrs
                          </span>
                        </div>
                      </div>
                      
                      <!-- Notes - Required -->
                      <div class="py-1 mt-2">
                        <div class="flex items-center">
                          <label class="block text-2xs font-medium text-gray-700">Notes</label>
                          <span class="ml-1 text-2xs text-red-500">*</span>
                          <span class="ml-1 text-2xs text-gray-400">(required)</span>
                        </div>
                        <textarea 
                          v-model="formEntry.notes"
                          rows="4"
                          placeholder="Enter notes about this timesheet entry..."
                          class="mt-0.5 block w-full rounded border border-gray-300 py-1.5 px-2 shadow-sm focus:border-sage focus:outline-none focus:ring-sage text-2xs resize-none"
                        ></textarea>
                      </div>
                      
                      <!-- Meeting Checkbox -->
                      <div class="mt-3">
                        <label class="flex items-center cursor-pointer">
                          <input
                            type="checkbox"
                            v-model="formEntry.is_meeting"
                            class="h-3 w-3 rounded border-gray-300 text-sage focus:ring-sage"
                          />
                          <span class="ml-2 text-2xs text-gray-600">
                            This is a meeting
                          </span>
                        </label>
                      </div>
                      
                      <!-- Impersonation section - moved to bottom -->
                      <div class="mt-4 border-t border-gray-100 pt-3">
                        <!-- Toggle for impersonation mode -->
                        <SwitchGroup as="div" class="flex items-center justify-between mb-2">
                          <SwitchLabel as="span" class="text-2xs text-gray-500">
                            Impersonation Mode
                          </SwitchLabel>
                          <Switch
                            v-model="showImpersonation"
                            :class="[
                              showImpersonation ? 'bg-sage' : 'bg-gray-200',
                              'relative inline-flex h-4 w-8 items-center rounded-full'
                            ]"
                          >
                            <span
                              :class="[
                                showImpersonation ? 'translate-x-4' : 'translate-x-0.5',
                                'inline-block h-3 w-3 transform rounded-full bg-white transition'
                              ]"
                            />
                          </Switch>
                        </SwitchGroup>
                        
                        <!-- Impersonation selection - only if toggled on -->
                        <div v-if="showImpersonation" class="bg-yellow-50 p-2 rounded text-xs space-y-1 border border-yellow-200">
                          <div class="flex items-start mb-1">
                            <svg class="h-3 w-3 text-yellow-400 mt-0.5 mr-1 flex-shrink-0" viewBox="0 0 20 20" fill="currentColor">
                              <path fill-rule="evenodd" d="M8.485 2.495c.673-1.167 2.357-1.167 3.03 0l6.28 10.875c.673 1.167-.17 2.625-1.516 2.625H3.72c-1.347 0-2.189-1.458-1.515-2.625L8.485 2.495zM10 5a.75.75 0 01.75.75v3.5a.75.75 0 01-1.5 0v-3.5A.75.75 0 0110 5zm0 9a1 1 0 100-2 1 1 0 000 2z" clip-rule="evenodd" />
                            </svg>
                            <p class="text-2xs text-yellow-700">
                              <strong>Warning:</strong> Creating entry for another user.
                            </p>
                          </div>
                          <select 
                            v-model="formEntry.impersonate_as_user_id" 
                            @change="formEntry.is_being_impersonated = (formEntry.impersonate_as_user_id ?? 0) > 0"
                            class="block w-full rounded border-gray-300 py-1 pl-2 pr-8 text-2xs focus:border-sage focus:outline-none focus:ring-sage"
                          >
                            <option :value="null">Select a user</option>
                            <option v-for="user in users" :key="user.id" :value="Number(user.id)">
                              {{ user.first_name }} {{ user.last_name }}
                            </option>
                          </select>
                        </div>
                      </div>
                    </div>
                    
                    <div class="bg-gray-50 px-3 py-2 sm:flex sm:flex-row-reverse border-t border-gray-200">
                      <button 
                        type="button" 
                        class="inline-flex justify-center rounded bg-sage px-2.5 py-1 text-xs font-semibold text-white shadow-sm hover:bg-sage-dark sm:ml-2 sm:w-auto"
                        @click="saveEntry"
                        :disabled="isSaving"
                      >
                        <template v-if="isSaving">
                          <svg class="animate-spin -ml-1 mr-2 h-3 w-3 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                          </svg>
                          Saving...
                        </template>
                        <template v-else>
                          {{ formEntry.entry_id ? 'Update' : 'Create' }}
                        </template>
                      </button>
                      <button 
                        type="button" 
                        class="mt-2 inline-flex justify-center rounded bg-white px-2.5 py-1 text-xs font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                        @click="closeModal"
                        :disabled="isSaving"
                      >
                        Cancel
                      </button>
                      <div class="sm:flex-grow">
                        <span v-if="isSaving" class="text-2xs text-gray-500 ml-2 sm:ml-0">Processing...</span>
                      </div>
                      <button 
                        v-if="formEntry.entry_id !== 0"
                        type="button" 
                        class="mt-2 inline-flex justify-center rounded bg-red-600 px-2.5 py-1 text-xs font-semibold text-white shadow-sm hover:bg-red-500 sm:mt-0 sm:w-auto"
                        @click="deleteEntry"
                        :disabled="isSaving"
                      >
                        <template v-if="isSaving">
                          <svg class="animate-spin -ml-1 mr-2 h-3 w-3 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                          </svg>
                          Deleting...
                        </template>
                        <template v-else>
                          Delete
                        </template>
                      </button>
                    </div>
                  </DialogPanel>
                </TransitionChild>
              </div>
            </div>
          </Dialog>
        </TransitionRoot>
      </div>
    </div>
  </div>
  
  <!-- Calendar Event Picker Modal -->
  <CalendarEventPicker
    :show="showCalendarPicker"
    :start-date="weekStartDate"
    :end-date="weekEndDate"
    @close="closeCalendarPicker"
    @select="handleCalendarEventSelect"
  />
</template>

<style scoped>
/* Ensure the calendar takes full height of the parent container */
.flex-auto {
  flex: 1 1 auto;
}

/* Styling for calendar cells during drag */
.calendar-cell {
  position: relative;
  z-index: 1;
  min-height: 2rem;
  /* Prevent text selection during drag operations */
  user-select: none;
  -webkit-user-select: none;
  -moz-user-select: none;
}

.calendar-cell:hover {
  background-color: rgba(88, 131, 126, 0.05);
}

/* Make the highlight class more specific and forceful */
.calendar-cell.bg-sage-pale,
div.calendar-cell.bg-sage-pale,
.timesheet-container .calendar-cell.bg-sage-pale {
  background-color: rgba(88, 131, 126, 0.3) !important;
  z-index: 15;
  position: relative;
}

/* Custom color variables for the muted green */
:root {
  --sage-green-primary: #58837e;
  --sage-green-dark: #476b67;
  --sage-green-light: #76a19c;
  --sage-green-pale: #e6efee;
}

/* Add Tailwind utility classes for our custom colors */
.bg-sage-pale {
  background-color: var(--sage-green-pale);
}

.border-sage {
  border-color: var(--sage-green-primary);
}

.bg-sage {
  background-color: var(--sage-green-primary);
}

.bg-sage-dark {
  background-color: var(--sage-green-dark);
}

.text-sage {
  color: var(--sage-green-primary);
}

/* Custom text size for extra small text */
.text-2xs {
  font-size: 0.65rem;
  line-height: 0.9rem;
}

/* Fix for calendar container height */
.timesheet-container {
  height: calc(100vh - 9rem);
  overflow: hidden;
  border-radius: 0.75rem; /* 12px rounded corners */
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
}

@media (min-width: 1024px) {
  .timesheet-container {
    height: calc(100vh - 6.5rem); /* Adjusted for the top padding on desktop */
    border-radius: 0.75rem;
  }
}

/* Allow entries to be positioned absolutely within calendar grid */
.timesheet-container .relative {
  position: relative;
}

/* Style for date inputs to make text size consistent with time inputs */
.date-input {
  font-size: 0.75rem !important;
}

/* Override browser-specific date picker text size */
.date-input::-webkit-datetime-edit {
  font-size: 0.75rem;
}

.date-input::-webkit-calendar-picker-indicator {
  font-size: 0.75rem;
  padding: 0;
  margin-left: 2px;
  width: 14px;
  height: 14px;
}

/* Disabled button styling */
button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Make sure spinner is always visible even in disabled state */
button:disabled svg.animate-spin {
  opacity: 1;
}
</style> 