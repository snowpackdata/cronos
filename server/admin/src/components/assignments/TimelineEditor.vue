<template>
  <div class="timeline-editor">
    <div class="mb-2 flex items-center justify-between">
      <label class="block text-sm font-medium text-gray-700">Weekly Commitment Schedule</label>
      <span class="text-xs text-gray-500">Click weeks to adjust hours</span>
    </div>
    
    <!-- Timeline Header -->
    <div class="timeline-container border border-gray-300 rounded-lg overflow-hidden bg-white">
      <!-- Week Headers -->
      <div class="timeline-header-wrapper">
        <div class="timeline-header bg-gray-50 border-b border-gray-300 flex">
          <div class="week-cell header-cell" v-for="(_, index) in weeks" :key="`header-${index}`">
            <div class="text-xs font-medium text-gray-700">{{ formatWeekLabel(weeks[index]) }}</div>
          </div>
        </div>
      </div>
      
      <!-- Commitment Bars -->
      <div class="timeline-body-wrapper">
        <div class="timeline-body p-2">
          <div class="commitment-track flex">
            <div 
              v-for="(_, index) in weeks" 
              :key="`bar-${index}`"
              class="week-bar-container"
              @mousedown="startDrag($event, index)"
            >
              <div 
                class="week-bar"
                :class="{ 'has-commitment': getWeekCommitment(index) > 0, 'dragging': draggingIndex === index }"
                :style="{ height: getBarHeight(getWeekCommitment(index)) }"
              >
                <span class="week-bar-text">{{ getWeekCommitment(index) }}h</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Quick Actions -->
      <div class="timeline-footer bg-gray-50 border-t border-gray-300 p-2 flex gap-2 text-xs">
        <button 
          type="button"
          @click="fillAllWeeks"
          class="px-2 py-1 bg-white border border-gray-300 rounded hover:bg-gray-50"
        >
          Fill All
        </button>
        <button 
          type="button"
          @click="clearAllWeeks"
          class="px-2 py-1 bg-white border border-gray-300 rounded hover:bg-gray-50"
        >
          Clear All
        </button>
        <input 
          v-model.number="bulkCommitment"
          type="number"
          min="0"
          max="40"
          placeholder="Hours"
          class="w-16 px-2 py-1 border border-gray-300 rounded text-xs"
        />
      </div>
    </div>
    
    <!-- Week Editor Modal -->
    <div v-if="editingWeekIndex !== null" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg p-4 w-80 shadow-xl">
        <h3 class="text-sm font-semibold mb-3">Edit Week {{ editingWeekIndex + 1 }}</h3>
        <div class="mb-3">
          <label class="block text-xs text-gray-600 mb-1">{{ formatWeekRange(weeks[editingWeekIndex]) }}</label>
          <input 
            v-model.number="editingCommitment"
            type="number"
            min="0"
            max="40"
            step="5"
            class="w-full px-3 py-2 border border-gray-300 rounded"
            placeholder="Hours per week"
          />
        </div>
        <div class="flex gap-2">
          <button 
            type="button"
            @click="saveWeekEdit"
            class="flex-1 bg-indigo-600 text-white px-3 py-2 rounded text-sm hover:bg-indigo-700"
          >
            Save
          </button>
          <button 
            type="button"
            @click="cancelWeekEdit"
            class="flex-1 bg-gray-200 text-gray-700 px-3 py-2 rounded text-sm hover:bg-gray-300"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';

export interface CommitmentSegment {
  start_date: string;
  end_date: string;
  commitment: number;
}

const props = defineProps<{
  startDate: string;  // YYYY-MM-DD
  endDate: string;    // YYYY-MM-DD
  segments?: CommitmentSegment[];
}>();

const emit = defineEmits<{
  'update:segments': [segments: CommitmentSegment[]]
}>();

// Generate weeks between start and end dates
const weeks = computed(() => {
  const start = new Date(props.startDate);
  const end = new Date(props.endDate);
  const weeksList: Date[] = [];
  
  // Start from the Sunday of the start week
  const current = new Date(start);
  current.setDate(current.getDate() - current.getDay()); // Move to Sunday
  
  const MAX_WEEKS = 104; // Limit to 2 years worth of weeks to prevent crashes
  let weekCount = 0;
  
  while (current <= end && weekCount < MAX_WEEKS) {
    weeksList.push(new Date(current));
    current.setDate(current.getDate() + 7); // Move to next week
    weekCount++;
  }
  
  return weeksList;
});

// Track commitment for each week
const weekCommitments = ref<number[]>([]);
const isInternalUpdate = ref(false); // Flag to prevent infinite loop

// Initialize commitments from segments
watch(() => [props.segments, weeks.value.length], () => {
  if (isInternalUpdate.value) return; // Skip if this is from our own update
  
  weekCommitments.value = weeks.value.map((week) => {
    if (!props.segments || props.segments.length === 0) return 0;
    
    for (const segment of props.segments) {
      const segStart = new Date(segment.start_date);
      const segEnd = new Date(segment.end_date);
      
      if (week >= segStart && week <= segEnd) {
        return segment.commitment;
      }
    }
    return 0;
  });
}, { immediate: true });

// Convert week commitments to segments when they change
watch(weekCommitments, (commitments) => {
  isInternalUpdate.value = true; // Set flag before emitting
  
  const segments: CommitmentSegment[] = [];
  let currentSegment: CommitmentSegment | null = null;
  
  commitments.forEach((commitment, index) => {
    const weekDate = weeks.value[index];
    // Format date without timezone conversion to avoid shifts
    const year = weekDate.getFullYear();
    const month = String(weekDate.getMonth() + 1).padStart(2, '0');
    const day = String(weekDate.getDate()).padStart(2, '0');
    const weekStartStr = `${year}-${month}-${day}`;
    
    // Calculate the Saturday (end of week) for the end_date
    const weekEnd = new Date(weekDate);
    weekEnd.setDate(weekEnd.getDate() + 6);
    const endYear = weekEnd.getFullYear();
    const endMonth = String(weekEnd.getMonth() + 1).padStart(2, '0');
    const endDay = String(weekEnd.getDate()).padStart(2, '0');
    const weekEndStr = `${endYear}-${endMonth}-${endDay}`;
    
    if (commitment > 0) {
      if (currentSegment && currentSegment.commitment === commitment) {
        // Extend current segment - update end_date to Saturday of this week
        currentSegment.end_date = weekEndStr;
      } else {
        // Start new segment
        if (currentSegment) {
          segments.push(currentSegment);
        }
        currentSegment = {
          start_date: weekStartStr,
          end_date: weekEndStr,
          commitment: commitment
        };
      }
    } else {
      // Commitment is 0, close current segment if any
      if (currentSegment) {
        segments.push(currentSegment);
        currentSegment = null;
      }
    }
  });
  
  // Push final segment
  if (currentSegment) {
    segments.push(currentSegment);
  }
  
  emit('update:segments', segments);
  
  // Reset flag after a tick
  setTimeout(() => {
    isInternalUpdate.value = false;
  }, 0);
}, { deep: true });

const getWeekCommitment = (index: number): number => {
  return weekCommitments.value[index] || 0;
};

const getBarHeight = (commitment: number): string => {
  const maxHeight = 60; // px
  const height = (commitment / 40) * maxHeight;
  return `${Math.max(height, 4)}px`;
};

const formatWeekLabel = (date: Date): string => {
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
};

const formatWeekRange = (weekStart: Date): string => {
  const weekEnd = new Date(weekStart);
  weekEnd.setDate(weekEnd.getDate() + 6);
  return `${formatWeekLabel(weekStart)} - ${formatWeekLabel(weekEnd)}`;
};

// Drag to adjust
const draggingIndex = ref<number | null>(null);
const dragStartY = ref(0);
const dragStartCommitment = ref(0);
const isDragging = ref(false);

const startDrag = (event: MouseEvent, index: number) => {
  event.preventDefault();
  event.stopPropagation();
  
  draggingIndex.value = index;
  dragStartY.value = event.clientY;
  dragStartCommitment.value = weekCommitments.value[index] || 0;
  isDragging.value = false; // Will become true on first move
  
  const handleMouseMove = (e: MouseEvent) => {
    if (draggingIndex.value === null) return;
    
    isDragging.value = true;
    
    // Calculate how much the mouse has moved (negative Y = up = more hours)
    const deltaY = dragStartY.value - e.clientY;
    
    // Convert pixels to hours (roughly 2px per hour)
    const hoursChange = Math.round(deltaY / 2);
    
    // Calculate new commitment (clamp between 0 and 40)
    const newCommitment = Math.max(0, Math.min(40, dragStartCommitment.value + hoursChange));
    
    // Update the commitment
    const newCommitments = [...weekCommitments.value];
    newCommitments[draggingIndex.value] = newCommitment;
    weekCommitments.value = newCommitments;
  };
  
  const handleMouseUp = () => {
    // If we didn't actually drag, open the modal
    if (!isDragging.value && draggingIndex.value !== null) {
      openWeekEditor(draggingIndex.value);
    }
    
    draggingIndex.value = null;
    isDragging.value = false;
    document.removeEventListener('mousemove', handleMouseMove);
    document.removeEventListener('mouseup', handleMouseUp);
  };
  
  document.addEventListener('mousemove', handleMouseMove);
  document.addEventListener('mouseup', handleMouseUp);
};

// Week editing modal (for precise input)
const editingWeekIndex = ref<number | null>(null);
const editingCommitment = ref(0);

const openWeekEditor = (index: number) => {
  editingWeekIndex.value = index;
  editingCommitment.value = weekCommitments.value[index] || 0;
};

const saveWeekEdit = () => {
  if (editingWeekIndex.value !== null) {
    // Create a new array to trigger reactivity
    const newCommitments = [...weekCommitments.value];
    newCommitments[editingWeekIndex.value] = editingCommitment.value;
    weekCommitments.value = newCommitments;
  }
  editingWeekIndex.value = null;
};

const cancelWeekEdit = () => {
  editingWeekIndex.value = null;
};

// Bulk actions
const bulkCommitment = ref(40);

const fillAllWeeks = () => {
  weekCommitments.value = weeks.value.map(() => bulkCommitment.value);
};

const clearAllWeeks = () => {
  weekCommitments.value = weeks.value.map(() => 0);
};

// Synchronize scrolling between header and body
let headerScrollHandler: (() => void) | null = null;
let bodyScrollHandler: (() => void) | null = null;

onMounted(() => {
  const headerWrapper = document.querySelector('.timeline-header-wrapper');
  const bodyWrapper = document.querySelector('.timeline-body-wrapper');
  
  if (headerWrapper && bodyWrapper) {
    let isHeaderScrolling = false;
    let isBodyScrolling = false;
    
    headerScrollHandler = () => {
      if (isBodyScrolling) return;
      isHeaderScrolling = true;
      bodyWrapper.scrollLeft = headerWrapper.scrollLeft;
      setTimeout(() => { isHeaderScrolling = false; }, 50);
    };
    
    bodyScrollHandler = () => {
      if (isHeaderScrolling) return;
      isBodyScrolling = true;
      headerWrapper.scrollLeft = bodyWrapper.scrollLeft;
      setTimeout(() => { isBodyScrolling = false; }, 50);
    };
    
    headerWrapper.addEventListener('scroll', headerScrollHandler);
    bodyWrapper.addEventListener('scroll', bodyScrollHandler);
  }
});

// Clean up event listeners on unmount
onBeforeUnmount(() => {
  const headerWrapper = document.querySelector('.timeline-header-wrapper');
  const bodyWrapper = document.querySelector('.timeline-body-wrapper');
  
  if (headerWrapper && headerScrollHandler) {
    headerWrapper.removeEventListener('scroll', headerScrollHandler);
  }
  if (bodyWrapper && bodyScrollHandler) {
    bodyWrapper.removeEventListener('scroll', bodyScrollHandler);
  }
});
</script>

<style scoped>
.timeline-container {
  max-width: 100%;
}

.timeline-header-wrapper {
  overflow-x: auto;
  overflow-y: hidden;
}

.timeline-header-wrapper::-webkit-scrollbar,
.timeline-body-wrapper::-webkit-scrollbar {
  height: 6px;
}

.timeline-header-wrapper::-webkit-scrollbar-thumb,
.timeline-body-wrapper::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 3px;
}

.timeline-header {
  display: flex;
  min-width: fit-content;
}

.timeline-body-wrapper {
  overflow-x: auto;
  overflow-y: hidden;
}

.timeline-body {
  min-width: fit-content;
}

.week-cell {
  width: 80px;
  min-width: 80px;
  flex-shrink: 0;
  padding: 8px 4px;
  text-align: center;
  border-right: 1px solid #e5e7eb;
}

.week-cell:last-child {
  border-right: none;
}

.commitment-track {
  display: flex;
  align-items: flex-end;
  min-height: 80px;
  gap: 0;
}

.week-bar-container {
  width: 80px;
  min-width: 80px;
  flex-shrink: 0;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  cursor: ns-resize;
  padding: 4px 8px;
  user-select: none;
}

.week-bar {
  width: 100%;
  background: #e5e7eb;
  border-radius: 4px 4px 0 0;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 4px;
}

.week-bar.has-commitment {
  background: #6366f1;
}

.week-bar:hover {
  opacity: 0.8;
  transform: scaleY(1.05);
}

.week-bar.dragging {
  opacity: 0.9;
  transition: none;
  box-shadow: 0 0 0 2px #4f46e5;
}

.week-bar-text {
  font-size: 11px;
  font-weight: 600;
  color: white;
  writing-mode: horizontal-tb;
}
</style>

