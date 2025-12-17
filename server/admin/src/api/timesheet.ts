import type { TimesheetEntry } from '../types/Timesheet';
import { fetchAll, fetchById, remove, createWithFormData, updateWithFormData } from './apiUtils';
import axios from 'axios';

/**
 * Validates timesheet entry data before sending to the API
 * @param entry - Entry data to validate
 * @returns Object with validation result and any error messages
 */
function validateTimesheetEntry(entry: Partial<TimesheetEntry>): { isValid: boolean; errors: string[] } {
  const errors: string[] = [];
  
  // Check required fields
  if (!entry.billing_code_id) errors.push('Billing code is required');
  if (!entry.start) errors.push('Start time is required');
  if (!entry.end) errors.push('End time is required');
  
  // Validate time range
  if (entry.start && entry.end) {
    const startTime = new Date(entry.start);
    const endTime = new Date(entry.end);
    if (endTime <= startTime) {
      errors.push('End time must be after start time');
    }
    
    // Check if duration is reasonable (optional validation)
    const durationMs = endTime.getTime() - startTime.getTime();
    const durationHours = durationMs / (1000 * 60 * 60);
    if (durationHours > 24) {
      errors.push('Duration exceeds 24 hours, please verify the times');
    }
  }
  
  return { 
    isValid: errors.length === 0,
    errors 
  };
}

/**
 * API service for timesheet-related operations
 */
const timesheetAPI = {
  /**
   * Fetch all timesheet entries, optionally filtered by date range and/or user
   * @param startDate - Optional start date (YYYY-MM-DD)
   * @param endDate - Optional end date (YYYY-MM-DD)
   * @param userId - Optional user ID to view another user's entries (admin only)
   * @returns Promise with timesheet entries
   */
  async getEntries(startDate?: string, endDate?: string, userId?: number | null): Promise<TimesheetEntry[]> {
    let url = 'entries';
    const params = new URLSearchParams();
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);
    if (userId) params.append('user_id', userId.toString());
    if (params.toString()) {
      url += `?${params.toString()}`;
    }
    
    const entries = await fetchAll<TimesheetEntry>(url);
    
    // Ensure all entries have the correct data types
    return entries.map(entry => ({
      ...entry,
      entry_id: Number(entry.entry_id),
      billing_code_id: Number(entry.billing_code_id),
      project_id: Number(entry.project_id),
      user_id: Number(entry.user_id),
      duration_hours: Number(entry.duration_hours),
      fee: Number(entry.fee),
      impersonate_as_user_id: entry.impersonate_as_user_id ? Number(entry.impersonate_as_user_id) : null,
      is_being_impersonated: Boolean(entry.is_being_impersonated)
    }));
  },

  /**
   * Fetch a single timesheet entry by ID
   * @param id - Entry ID
   * @returns Promise with timesheet entry data
   */
  async getEntry(id: number): Promise<TimesheetEntry> {
    const entry = await fetchById<TimesheetEntry>('entries', id);
    
    // Ensure correct data types
    return {
      ...entry,
      entry_id: Number(entry.entry_id),
      billing_code_id: Number(entry.billing_code_id),
      project_id: Number(entry.project_id),
      user_id: Number(entry.user_id),
      duration_hours: Number(entry.duration_hours),
      fee: Number(entry.fee),
      impersonate_as_user_id: entry.impersonate_as_user_id ? Number(entry.impersonate_as_user_id) : null,
      is_being_impersonated: Boolean(entry.is_being_impersonated)
    };
  },

  /**
   * Fetch users who can be impersonated for timesheet entries
   * @returns Promise with users data
   */
  async getUsers(): Promise<any[]> {
    return fetchAll('staff');
  },

  /**
   * Fetch active billing codes for timesheet entries
   * @returns Promise with billing codes data
   */
  async getActiveBillingCodes(): Promise<any[]> {
    const codes = await fetchAll('active_billing_codes');
    
    // Normalize billing code IDs to ensure consistent format
    // Add more robust handling for missing or invalid IDs
    return codes.map((code: any) => {
      // Get ID from one of the possible sources
      const rawId = code.id || code.billing_code_id || code.ID;
      
      // Convert to number and ensure it's valid
      const numericId = Number(rawId);
      const validId = !isNaN(numericId) && numericId > 0 ? numericId : null;
      
      // Log warning if we have an invalid ID
      if (validId === null) {
        console.warn('Found billing code with invalid ID:', code);
      }
      
      return {
        ...code,
        // Use the valid ID or keep original for debugging
        id: validId,
        billing_code_id: validId
      };
    }).filter(code => code.id !== null); // Only return codes with valid IDs
  },

  /**
   * Create a new timesheet entry with validation
   * @param entry - Timesheet entry data
   * @returns Promise with created entry
   * @throws Error if validation fails
   */
  async createEntry(entry: TimesheetEntry): Promise<TimesheetEntry> {
    // Validate entry data
    const validation = validateTimesheetEntry(entry);
    if (!validation.isValid) {
      throw new Error(`Validation failed: ${validation.errors.join(', ')}`);
    }
    
    // Prepare URLSearchParams for compatibility with Go backend
    const formData = new URLSearchParams();
    
    // Ensure billing_code_id is a valid number and convert to string
    const billingCodeId = Number(entry.billing_code_id);
    if (isNaN(billingCodeId) || billingCodeId <= 0) {
      throw new Error('Invalid billing code ID');
    }
    formData.set("billing_code_id", billingCodeId.toString());
    
    // NOTE: We work directly with the ISO string format to avoid timezone issues
    // Extract date parts in local time to preserve the user's selected date
    const formatDateTimeForAPI = (isoString: string) => {
      // Parse the ISO string directly to extract components
      const [datePart, timePart] = isoString.split('T');
      const [timeHours, timeMinutes] = timePart.split(':');
      
      // Format as YYYY-MM-DDThh:mm which is what the Go backend expects
      return `${datePart}T${timeHours}:${timeMinutes}`;
    };
    
    // Format dates exactly as they are in the ISO strings to prevent timezone issues
    formData.set("start", formatDateTimeForAPI(entry.start));
    formData.set("end", formatDateTimeForAPI(entry.end));
    
    // Always include notes (even if empty)
    formData.set("notes", entry.notes || '');
    
    // Handle is_meeting flag
    formData.set("is_meeting", entry.is_meeting ? "true" : "false");
    
    // Handle impersonation
    if (entry.impersonate_as_user_id) {
      formData.set("impersonate_as_user_id", entry.impersonate_as_user_id.toString());
    }
    
    // Log the form data for debugging
    
    try {
      // Use direct API call with URLSearchParams
      const normalizedUrl = `/api/entries/0`;
      
      // Create custom headers for URL-encoded form data
      const headers = {
        'Content-Type': 'application/x-www-form-urlencoded',
        'x-access-token': localStorage.getItem('snowpack_token')
      };
      
      // Make the API call with the correct Content-Type
      const response = await axios.post(normalizedUrl, formData.toString(), { headers });
      return response.data;
    } catch (error) {
      console.error('Failed to create timesheet entry:', error);
      throw error;
    }
  },

  /**
   * Update an existing timesheet entry with validation
   * @param entry - Entry data to update
   * @returns Promise with updated entry
   * @throws Error if validation fails
   */
  async updateEntry(entry: TimesheetEntry): Promise<TimesheetEntry> {
    // Ensure entry_id is valid and convert to a proper number
    const entryId = Number(entry.entry_id);
    if (!entryId || isNaN(entryId) || entryId <= 0) {
      throw new Error(`Invalid entry ID for update: ${entry.entry_id}`);
    }
    
    // Validate entry data
    const validation = validateTimesheetEntry(entry);
    if (!validation.isValid) {
      throw new Error(`Validation failed: ${validation.errors.join(', ')}`);
    }
    
    // Prepare URLSearchParams for compatibility with Go backend
    const formData = new URLSearchParams();
    
    // Ensure billing_code_id is a valid number and convert to string
    const billingCodeId = Number(entry.billing_code_id);
    if (isNaN(billingCodeId) || billingCodeId <= 0) {
      throw new Error('Invalid billing code ID');
    }
    formData.set("billing_code_id", billingCodeId.toString());
    
    // NOTE: We work directly with the ISO string format to avoid timezone issues
    // Extract date parts in local time to preserve the user's selected date
    const formatDateTimeForAPI = (isoString: string) => {
      // Parse the ISO string directly to extract components
      const [datePart, timePart] = isoString.split('T');
      const [timeHours, timeMinutes] = timePart.split(':');
      
      // Format as YYYY-MM-DDThh:mm which is what the Go backend expects
      return `${datePart}T${timeHours}:${timeMinutes}`;
    };
    
    // Format dates exactly as they are in the ISO strings to prevent timezone issues
    formData.set("start", formatDateTimeForAPI(entry.start));
    formData.set("end", formatDateTimeForAPI(entry.end));
    
    // Always include notes (even if empty)
    formData.set("notes", entry.notes || '');
    
    // Handle is_meeting flag
    formData.set("is_meeting", entry.is_meeting ? "true" : "false");
    
    // Handle impersonation
    if (entry.impersonate_as_user_id) {
      const impersonateId = Number(entry.impersonate_as_user_id);
      formData.set("impersonate_as_user_id", impersonateId.toString());
    } else {
      // Explicitly set to 0 to clear impersonation
      formData.set("impersonate_as_user_id", "0");
    }
    
    // Log the form data for debugging
    
    try {
      // Use direct API call with URLSearchParams
      const normalizedUrl = `/api/entries/${entryId}`;
      
      // Create custom headers for URL-encoded form data
      const headers = {
        'Content-Type': 'application/x-www-form-urlencoded',
        'x-access-token': localStorage.getItem('snowpack_token')
      };
      
      // Make the API call with the correct Content-Type
      const response = await axios.put(normalizedUrl, formData.toString(), { headers });
      return response.data;
    } catch (error) {
      console.error(`Failed to update timesheet entry ${entryId}:`, error);
      throw error;
    }
  },

  /**
   * Delete a timesheet entry
   * @param id - ID of entry to delete
   * @returns Promise with deletion result
   */
  async deleteEntry(id: number): Promise<any> {
    try {
      return remove('entries', id);
    } catch (error) {
      console.error('Failed to delete timesheet entry:', error);
      throw error;
    }
  },

  /**
   * Fetch timesheet entries for a specific time period
   * @param startDate - Start date in YYYY-MM-DD format
   * @param endDate - End date in YYYY-MM-DD format
   * @returns Promise with filtered timesheet entries
   */
  async getEntriesByDateRange(startDate: string, endDate: string): Promise<TimesheetEntry[]> {
    return fetchAll<TimesheetEntry>(`entries?start_date=${startDate}&end_date=${endDate}`);
  },

  /**
   * Fetch timesheet entries for a specific user
   * @param userId - User ID
   * @returns Promise with user's timesheet entries
   */
  async getEntriesByUser(userId: number): Promise<TimesheetEntry[]> {
    return fetchAll<TimesheetEntry>(`entries?user_id=${userId}`);
  },

  /**
   * Fetch timesheet entries for a specific project
   * @param projectId - Project ID
   * @returns Promise with project's timesheet entries
   */
  async getEntriesByProject(projectId: number): Promise<TimesheetEntry[]> {
    return fetchAll<TimesheetEntry>(`entries?project_id=${projectId}`);
  }
};

// Export as default
export default timesheetAPI;

// Export individual functions for backward compatibility
export const getEntries = timesheetAPI.getEntries;
export const getEntry = timesheetAPI.getEntry;
export const getUsers = timesheetAPI.getUsers;
export const getActiveBillingCodes = timesheetAPI.getActiveBillingCodes;
export const createEntry = timesheetAPI.createEntry;
export const updateEntry = timesheetAPI.updateEntry;
export const deleteEntry = timesheetAPI.deleteEntry;
export const getEntriesByDateRange = timesheetAPI.getEntriesByDateRange;
export const getEntriesByUser = timesheetAPI.getEntriesByUser;
export const getEntriesByProject = timesheetAPI.getEntriesByProject;

export const createTimesheetEntry = async (formData: Partial<TimesheetEntry>): Promise<TimesheetEntry> => {
  try {
    // Create a FormData object instead of a plain object
    const data = new FormData();
    
    // Add billing_code_id (required)
    if (formData.billing_code_id) {
      data.set("billing_code_id", formData.billing_code_id.toString());
    } else {
      throw new Error("Billing code ID is required");
    }
    
    // Add start time (required)
    if (formData.start) {
      // Format the date/time as expected by the server (YYYY-MM-DDThh:mm)
      const [datePart, timePart] = formData.start.split('T');
      const [timeHours, timeMinutes] = timePart.split(':');
      data.set("start", `${datePart}T${timeHours}:${timeMinutes}`);
    } else {
      throw new Error("Start time is required");
    }
    
    // Add end time (required)
    if (formData.end) {
      // Format the date/time as expected by the server (YYYY-MM-DDThh:mm)
      const [datePart, timePart] = formData.end.split('T');
      const [timeHours, timeMinutes] = timePart.split(':');
      data.set("end", `${datePart}T${timeHours}:${timeMinutes}`);
    } else {
      throw new Error("End time is required");
    }
    
    // Add notes (can be empty)
    data.set("notes", formData.notes || "");
    
    // Handle impersonation
    if (formData.impersonate_as_user_id) {
      data.set("impersonate_as_user_id", formData.impersonate_as_user_id.toString());
    } else {
      // Explicitly set to 0 to clear impersonation
      data.set("impersonate_as_user_id", "0");
    }
    
    // Add project_id if available (optional)
    if (formData.project_id) {
      data.set("project_id", formData.project_id.toString());
    }
    
    // Log the form data for debugging
    console.log("Creating timesheet entry with data:", 
      Array.from(data.entries()).map(([key, value]) => `${key}: ${value}`).join(', ')
    );
    
    // Use the createWithFormData function specialized for FormData
    return await createWithFormData<TimesheetEntry>('entries', data);
  } catch (error) {
    console.error('Failed to create timesheet entry:', error);
    throw error;
  }
};

export const updateTimesheetEntry = async (entryId: number, formData: Partial<TimesheetEntry>): Promise<TimesheetEntry> => {
  try {
    // Create a FormData object instead of a plain object
    const data = new FormData();
    
    // Add billing_code_id (required)
    if (formData.billing_code_id) {
      data.set("billing_code_id", formData.billing_code_id.toString());
    } else {
      throw new Error("Billing code ID is required");
    }
    
    // Add start time (required)
    if (formData.start) {
      // Format the date/time as expected by the server (YYYY-MM-DDThh:mm)
      const [datePart, timePart] = formData.start.split('T');
      const [timeHours, timeMinutes] = timePart.split(':');
      data.set("start", `${datePart}T${timeHours}:${timeMinutes}`);
    } else {
      throw new Error("Start time is required");
    }
    
    // Add end time (required)
    if (formData.end) {
      // Format the date/time as expected by the server (YYYY-MM-DDThh:mm)
      const [datePart, timePart] = formData.end.split('T');
      const [timeHours, timeMinutes] = timePart.split(':');
      data.set("end", `${datePart}T${timeHours}:${timeMinutes}`);
    } else {
      throw new Error("End time is required");
    }
    
    // Add notes (can be empty)
    data.set("notes", formData.notes || "");
    
    // Handle impersonation
    if (formData.impersonate_as_user_id) {
      data.set("impersonate_as_user_id", formData.impersonate_as_user_id.toString());
    } else {
      // Explicitly set to 0 to clear impersonation
      data.set("impersonate_as_user_id", "0");
    }
    
    // Add project_id if available (optional)
    if (formData.project_id) {
      data.set("project_id", formData.project_id.toString());
    }
    
    // Log the form data for debugging
    console.log(`Updating timesheet entry ${entryId} with data:`,
      Array.from(data.entries()).map(([key, value]) => `${key}: ${value}`).join(', ')
    );
    
    // Use the updateWithFormData function specialized for FormData
    return await updateWithFormData<TimesheetEntry>('entries', entryId, data);
  } catch (error) {
    console.error(`Failed to update timesheet entry ${entryId}:`, error);
    throw error;
  }
}; 