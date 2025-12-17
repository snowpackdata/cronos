<script setup lang="ts">
import { computed } from 'vue';
import type { TimesheetEntry } from '../../types/Timesheet';
import type { CSSProperties } from 'vue';
import { UserIcon, UserMinusIcon, ExclamationTriangleIcon } from '@heroicons/vue/20/solid';

// Define props
const props = defineProps<{
  entry: TimesheetEntry;
  dayIndex: number;
  day: Date;
  timeRangeStart?: number; // Starting hour (e.g., 8 for 8am)
  timeRangeEnd?: number;   // Ending hour (e.g., 22 for 10pm)
}>();

// Define emits
const emit = defineEmits<{
  (e: 'edit', entry: TimesheetEntry): void;
}>();

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

// Calculate grid row placement for the entry
const gridRowValues = computed(() => {
  if (!props.entry || !props.entry.start || !props.entry.end) {
    return { start: 0, span: 0 };
  }

  try {
    // Extract date parts directly from ISO strings to avoid timezone issues
    const entryDatePart = props.entry.start.split('T')[0];
    
    // Format day as ISO date string and extract just the date part
    const dayDateStr = props.day.toISOString().split('T')[0];
    
    // Check if entry belongs to this day (using date part only)
    if (entryDatePart !== dayDateStr) {
      return { start: 0, span: 0 };
    }
    
    // Handle times timezone-agnostically - extract hours and minutes directly from the ISO string
    const startTimeParts = props.entry.start.split('T')[1].split(':');
    const endTimeParts = props.entry.end.split('T')[1].split(':');
    
    const startHour = parseInt(startTimeParts[0], 10);
    const startMinute = parseInt(startTimeParts[1], 10);
    const endHour = parseInt(endTimeParts[0], 10);
    const endMinute = parseInt(endTimeParts[1], 10);
    
    // Time range settings (default to 8am-10pm if not provided)
    const timeRangeStart = props.timeRangeStart || 8;  // 8am
    
    // GRID POSITIONING CALCULATIONS
    
    // Each hour has 2 rows (for 30-min increments)
    // Grid starts with a header row (1.75rem) and 1-indexed, so add 2 to account for these
    const rowOffset = 2;
    
    // Calculate starting row - snap to nearest 30-minute block
    const startTotalMinutes = (startHour - timeRangeStart) * 60 + startMinute;
    // Round down to nearest 30-minute block
    const startBlockMinutes = Math.floor(startTotalMinutes / 30) * 30;
    const startRow = rowOffset + (startBlockMinutes / 30);
    
    // Calculate span (duration)
    const durationHours = endHour - startHour;
    const durationMinutes = endMinute - startMinute;
    const totalMinutes = (durationHours * 60) + durationMinutes;
    // Round up to nearest 30-minute block, minimum 1 block
    const totalHalfHours = Math.max(1, Math.ceil(totalMinutes / 30));
    
    //   startRow,
    //   span: totalHalfHours,
    //   startHour,
    //   startMinute,
    //   endHour,
    //   endMinute,
    //   timeRangeStart
    // });
    
    return {
      start: startRow,
      span: Math.max(1, totalHalfHours) // Ensure minimum span of 1
    };
  } catch (error) {
    console.error('Error calculating grid position for entry:', props.entry, error);
    return { start: 0, span: 0 };
  }
});

// Get column start based on day index
const colStart = computed(() => {
  // Add 1 because grid is 1-indexed
  return props.dayIndex + 1;
});

// Get color based on entry state
const entryColor = computed(() => {
  if (!props.entry || !props.entry.state) {
    return 'bg-white text-gray-700 hover:bg-gray-50 border border-gray-200';
  }

  switch (props.entry.state) {
    case 'ENTRY_STATE_UNAFFILIATED':
      return 'bg-orange-100 text-orange-800 hover:bg-orange-200 border-2 border-orange-500';
    case 'ENTRY_STATE_DRAFT':
      return 'bg-amber-50 text-amber-700 hover:bg-amber-100 border border-amber-200';
    case 'ENTRY_STATE_REJECTED':
      return 'bg-red-100 text-red-800 hover:bg-red-200 border-2 border-red-400';
    case 'ENTRY_STATE_APPROVED':
      return 'bg-blue-50 text-blue-700 hover:bg-blue-100 border border-blue-200';
    case 'ENTRY_STATE_EXCLUDED':
      return 'bg-gray-100 text-gray-600 hover:bg-gray-200 border border-gray-400';
    case 'ENTRY_STATE_INVOICED':
      return 'bg-purple-50 text-blue-700 hover:bg-blue-100 border border-blue-200';
    case 'ENTRY_STATE_SENT':
      return 'bg-indigo-50 text-indigo-700 hover:bg-indigo-100 border border-indigo-200';
    case 'ENTRY_STATE_PAID':
      return 'bg-gray-50 text-gray-700 hover:bg-gray-200 border border-gray-300';
    case 'ENTRY_STATE_VOID':
      return 'bg-gray-50 text-gray-700 hover:bg-gray-200 border border-gray-300';
    default:
      return 'bg-white text-gray-700 hover:bg-gray-50 border border-gray-200';
  }
});

// Handle edit button click
const handleEdit = () => {
  emit('edit', props.entry);
};

// Calculate a subtle offset for entries based on time to help with overlapping entries
const entryOffset = computed(() => {
  if (!props.entry || !props.entry.start) {
    return { right: '0%' };
  }

  try {
    // Extract the entry ID - higher IDs should be wider and at the bottom
    const entryId = props.entry.entry_id || 0;
    
    // Invert logic: Higher IDs should be wider (at bottom), lower IDs narrower (at top)
    // Use modulo to cycle through different styles (0-4)
    const offsetIndex = entryId % 5;
    
    // Calculate offsets - higher offsetIndex (higher ID) means less offset
    const rightOffset = (4 - offsetIndex) * 6;  // 6% increments for width from right
    
    // Z-index calculation - LOWER values for higher IDs so they appear at bottom
    // Keep z-index below 30 to stay under the sticky nav (z-40)
    const calculatedZIndex = 25 - offsetIndex;
    
    return {
      right: `${rightOffset}%`,
      width: offsetIndex === 4 ? '100%' : `calc(100% - ${rightOffset}%)`, // Full width for highest offsetIndex
      boxShadow: `0 ${(4-offsetIndex) + 1}px ${(4-offsetIndex) * 2 + 2}px rgba(0, 0, 0, 0.0${(4-offsetIndex) + 1})`,
      zIndex: calculatedZIndex
    };
  } catch (error) {
    console.error('Error calculating entry offset:', error);
    return { right: '0%' };
  }
});

// Get grid styles for the entry with better z-index handling
const gridStyle = computed((): CSSProperties => {
  if (gridRowValues.value.start === 0 || gridRowValues.value.span === 0) {
    return { display: 'none' };
  }

  return {
    gridArea: `${gridRowValues.value.start} / ${colStart.value} / span ${gridRowValues.value.span} / span 1`,
    position: 'relative'
  };
});

// NOTE: This computed property is temporarily unused but kept for future implementation
// of visibility filtering
/*
const isVisibleForDay = computed(() => {
  try {
    // Convert entry date to YYYY-MM-DD format
    const entryDate = new Date(props.entry.start);
    const entryDatePart = entryDate.toISOString().split('T')[0];
    
    // Get the current day's date in YYYY-MM-DD format
    const dayDateStr = props.day.toISOString().split('T')[0];
    
    // Compare the date strings directly to avoid timezone issues
    return entryDatePart === dayDateStr;
  } catch (e) {
    console.error('Error comparing entry date:', e);
    return false;
  }
});
*/

// Calculate grid positioning based on start/end times
// NOTE: This computed property is temporarily unused but kept for future implementation
// of alternative grid positioning approach
// const gridArea = computed(() => {
//   if (!isVisibleForDay.value) return '';
//   
//   // Compute the grid row positioning based on hours
//   const startTime = new Date(props.entry.start);
//   const endTime = new Date(props.entry.end);
//   
//   // Get hours and adjust for grid (1-based in grid)
//   const startHour = startTime.getHours();
//   const startMin = startTime.getMinutes();
//   const endHour = endTime.getHours();
//   const endMin = endTime.getMinutes();
//   
//   // Calculate fractional positions
//   const startRow = startHour + 1 + (startMin / 60);
//   const endRow = endHour + 1 + (endMin / 60);
//   
//   // Format for grid-area property
//   const area = `${startRow} / 1 / ${endRow} / 2`;
//   
//   //   date: props.entry.start_date,
//   //   startHour,
//   //   startMin,
//   //   endHour,
//   //   endMin,
//   //   gridArea: area
//   // });
//   
//   return area;
// });

// New computed property to determine if diagonal shading should be applied
const hasImpersonationStyles = computed(() => {
  return props.entry?.is_being_impersonated === true;
});
</script>

<template>
  <li 
    v-if="gridRowValues.span > 0"
    class="relative flex timesheet-entry"
    :style="gridStyle"
  >
    <a 
      href="#" 
      @click.prevent="handleEdit"
      class="group absolute inset-1 flex flex-col overflow-y-auto rounded-lg p-2 text-xs/5 shadow-sm entry-link"
      :class="[
        entryColor,
        { 'impersonation-striping': hasImpersonationStyles }
      ]"
      :style="entryOffset"
    >
      <div class="flex flex-col mb-1">
        <!-- Billing code name on top -->
        <span class="font-semibold truncate text-xs">{{ entry.billing_code_name }}</span>
        
        <!-- Color-coded badge below with impersonation icon if needed -->
        <div class="flex items-center gap-1 mt-0.5">
          <span 
            v-if="entry.billing_code"
            :style="{ backgroundColor: hashCodeToColor(entry.billing_code) }"
            :class="[
              'inline-flex items-center px-1.5 py-0.5 rounded text-[0.6rem] font-medium', 
              getTextColor(hashCodeToColor(entry.billing_code))
            ]"
          >
            {{ entry.billing_code }}
          </span>
          
          <!-- UNAFFILIATED entry warning -->
          <ExclamationTriangleIcon 
            v-if="entry.state === 'ENTRY_STATE_UNAFFILIATED'" 
            class="h-4 w-4 text-orange-600" 
            title="This work is not associated with any invoice - unbilled!"
          />
          
          <!-- REJECTED entry warning -->
          <ExclamationTriangleIcon 
            v-if="entry.state === 'ENTRY_STATE_REJECTED'" 
            class="h-4 w-4 text-red-600" 
            title="This entry was rejected - edit and save to resubmit"
          />
          
          <!-- Impersonating someone else -->
          <UserIcon 
            v-if="entry.impersonate_as_user_id !== null" 
            class="h-3 w-3 text-gray-700" 
            title="You are impersonating another user"
          />
          
          <!-- Someone is impersonating you -->
          <UserMinusIcon 
            v-if="entry.is_being_impersonated" 
            class="h-3 w-3 text-gray-700" 
            title="Someone is impersonating you"
          />
        </div>
      </div>
      <p class="font-semibold text-[0.65rem]">({{ entry.duration_hours }} hrs)</p>
      <div class="mt-1">
        <div class="text-[0.6rem] line-clamp-2">{{ entry.notes }}</div>
      </div>
    </a>
  </li>
</template>

<style scoped>
/* Time entries hover effect */
a.entry-link {
  transition: all 0.2s ease;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  z-index: 20; /* Ensure entries appear above the grid */
  position: absolute; /* Create stacking context */
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  border-width: 1px !important;
  margin: 1px !important; /* Reduced margin for better alignment */
}

a.entry-link:hover {
  box-shadow: 0 6px 10px -2px rgba(0, 0, 0, 0.2), 0 3px 6px -1px rgba(0, 0, 0, 0.1);
  transform: translateY(-1px);
  z-index: 1000 !important; /* Much higher z-index to ensure it's above all other entries */
  margin: 0 !important; /* Remove margin on hover for full size */
  border-width: 2px !important;
}

/* Add a class for parent element to make z-index work */
.timesheet-entry {
  padding: 0;
  margin: 0;
}

.timesheet-entry:hover {
  z-index: 900 !important; /* Apply higher z-index to the parent element when hovering */
}

/* Diagonal striping for entries where user is being impersonated */
.impersonation-striping {
  background-image: repeating-linear-gradient(
    45deg,
    rgba(245, 158, 11, 0.05), /* amber-500 with low opacity */
    rgba(245, 158, 11, 0.05) 10px,
    rgba(255, 255, 255, 0) 10px,
    rgba(255, 255, 255, 0) 20px
  ) !important;
}
</style> 