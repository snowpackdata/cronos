<template>
  <div class="capacity-gantt-container">
    <!-- Loading state -->
    <div v-if="isLoading" class="flex flex-col items-center justify-center p-6 bg-white rounded-lg shadow">
      <i class="fas fa-spinner fa-spin text-3xl text-blue-500 mb-2"></i>
      <span class="text-sm text-gray-600">Loading capacity data...</span>
    </div>
    
    <!-- Error state -->
    <div v-else-if="error" class="flex flex-col items-center justify-center p-6 bg-white rounded-lg shadow">
      <i class="fas fa-exclamation-circle text-3xl text-red-500 mb-2"></i>
      <span class="text-sm text-gray-600 mb-1">{{ error }}</span>
      <button @click="loadCapacityData" class="mt-3 px-3 py-1.5 text-xs font-semibold text-white bg-blue-600 rounded-md hover:bg-blue-700">
        Retry
      </button>
    </div>
    
    <!-- Gantt Chart -->
    <div v-else class="bg-white rounded-lg shadow overflow-hidden">
      <div class="overflow-x-auto">
        <div class="inline-block min-w-full align-middle">
          <div class="gantt-container">
            <!-- Header Row -->
            <div class="gantt-header">
              <div class="gantt-label-column">
                <div class="p-2 text-xs font-semibold text-gray-700">
                  Project / Staff
                </div>
              </div>
              <div class="gantt-timeline">
                <div
                  v-for="(week, index) in weeks"
                  :key="index"
                  :class="[
                    'gantt-week-header',
                    isCurrentWeek(week) ? 'current-week' : ''
                  ]"
                >
                  <div class="week-label">{{ formatDate(week) }}</div>
                  <div v-if="isCurrentWeek(week)" class="current-week-indicator">
                    <i class="fas fa-arrow-down text-xs"></i>
                  </div>
                </div>
              </div>
            </div>

            <!-- Data Rows -->
            <template v-for="item in groupedData" :key="item.type === 'group' ? `group-${item.groupKey}` : `item-${item.id}`">
              <!-- Project Group Row (with aggregated data) -->
              <div
                v-if="item.type === 'group'"
                class="gantt-group-row"
                @click="toggleGroup(item.groupKey!)"
              >
                <div class="gantt-label-column">
                  <div class="px-2 py-1.5 flex items-center cursor-pointer">
                    <i 
                      class="fas text-xs mr-2 text-gray-500 transition-transform"
                      :class="expandedGroups.has(item.groupKey!) ? 'fa-chevron-down' : 'fa-chevron-right'"
                    ></i>
                    <div class="text-xs font-semibold text-gray-700">{{ item.groupName }}</div>
                  </div>
                </div>
                <div class="gantt-timeline">
                  <div
                    v-for="(week, index) in weeks"
                    :key="index"
                    :class="[
                      'gantt-cell',
                      isCurrentWeek(week) ? 'current-week' : ''
                    ]"
                  >
                    <!-- Aggregate all staff for this project for this week -->
                    <div class="relative">
                      <div
                        v-if="getAggregatedCommitment(item.allAssignments!, week) > 0"
                        class="gantt-bar gantt-bar-aggregated"
                        :style="{ 
                          height: getBarHeight(getAggregatedCommitment(item.allAssignments!, week)),
                          background: `linear-gradient(to right, ${getBarColor(getAggregatedUtilization(item.allAssignments!, week))} ${getFillPercentage(getAggregatedUtilization(item.allAssignments!, week))}%, ${getBarColor(getAggregatedUtilization(item.allAssignments!, week))}40 ${getFillPercentage(getAggregatedUtilization(item.allAssignments!, week))}%)`
                        }"
                        :title="`${item.groupName} - ${getAggregatedCommitment(item.allAssignments!, week)}h/week (${getAggregatedActualHours(item.allAssignments!, week).toFixed(1)}h actual, ${getAggregatedUtilization(item.allAssignments!, week).toFixed(0)}%)`"
                        @click.stop="handleAggregatedBarClick(item.allAssignments!, week)"
                      >
                        <span class="gantt-bar-text">{{ getAggregatedCommitment(item.allAssignments!, week) }}h</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              
              <!-- Staff Item Row (shown when project is expanded) -->
              <div
                v-else
                class="gantt-row gantt-row-child"
              >
                <div class="gantt-label-column">
                  <div class="p-2 pl-8">
                    <div class="text-xs font-medium text-gray-600">{{ item.label }}</div>
                  </div>
                </div>
                <div class="gantt-timeline">
                  <div
                    v-for="(week, index) in weeks"
                    :key="index"
                    :class="[
                      'gantt-cell',
                      isCurrentWeek(week) ? 'current-week' : ''
                    ]"
                  >
                    <div
                      v-for="assignment in item.assignments"
                      :key="assignment.ID"
                      class="relative"
                    >
                      <div
                        v-if="dateInRange(week, assignment.start_date, assignment.end_date)"
                        class="gantt-bar"
                        :style="{ 
                          height: getBarHeight(getCommitmentForWeek(assignment, week)),
                          background: `linear-gradient(to right, ${getBarColor(getUtilization(assignment, week))} ${getFillPercentage(getUtilization(assignment, week))}%, ${getBarColor(getUtilization(assignment, week))}40 ${getFillPercentage(getUtilization(assignment, week))}%)`
                        }"
                        :title="`${item.label} - ${getCommitmentForWeek(assignment, week)}h/week (${getActualHours(assignment, week).toFixed(1)}h actual, ${getUtilization(assignment, week).toFixed(0)}%)`"
                        @click.stop="handleBarClick(assignment, week)"
                      >
                        <span class="gantt-bar-text">{{ getCommitmentForWeek(assignment, week) }}h</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </template>

            <!-- Empty State -->
            <div v-if="groupedData.length === 0" class="p-8 text-center">
              <i class="fas fa-calendar-times text-4xl text-gray-400 mb-3"></i>
              <p class="text-sm text-gray-600">No capacity assignments found</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Detail Modal -->
    <div v-if="showDetailModal" class="fixed inset-0 z-50 overflow-y-auto" @click="closeDetailModal">
      <div class="flex items-center justify-center min-h-screen px-4">
        <div class="fixed inset-0 bg-gray-900 bg-opacity-75 transition-opacity"></div>
        
        <div class="relative bg-white rounded-lg shadow-xl max-w-2xl w-full p-6 max-h-[90vh] overflow-y-auto" @click.stop>
          <div class="flex justify-between items-start mb-4">
            <div>
              <h3 class="text-lg font-semibold text-gray-800">
                {{ selectedAssignment?.project.name }}
              </h3>
              <p class="text-sm text-gray-600 mt-0.5">
                {{ selectedAssignment?.employee.first_name }} {{ selectedAssignment?.employee.last_name }}
              </p>
              <p class="text-xs text-gray-500 mt-1">
                Week of {{ selectedWeek ? formatDate(selectedWeek) : '' }}
              </p>
            </div>
            <button @click="closeDetailModal" class="text-gray-500 hover:text-gray-700">
              <i class="fas fa-times text-xl"></i>
            </button>
          </div>

          <div v-if="selectedAssignment && selectedWeek" class="space-y-4">
            <!-- Summary Stats -->
            <div class="grid grid-cols-3 gap-4">
              <div class="bg-gray-50 p-4 rounded-lg">
                <p class="text-xs text-gray-600 mb-1">Commitment</p>
                <p class="text-2xl font-semibold text-gray-800">
                  {{ selectedGroupAssignments 
                    ? getAggregatedCommitment(selectedGroupAssignments, selectedWeek) 
                    : getCommitmentForWeek(selectedAssignment, selectedWeek) 
                  }}h
                </p>
              </div>
              <div class="bg-gray-50 p-4 rounded-lg">
                <p class="text-xs text-gray-600 mb-1">Actual Hours</p>
                <p class="text-2xl font-semibold text-gray-800">
                  {{ selectedGroupAssignments 
                    ? getAggregatedActualHours(selectedGroupAssignments, selectedWeek).toFixed(1) 
                    : getActualHours(selectedAssignment, selectedWeek).toFixed(1) 
                  }}h
                </p>
              </div>
              <div class="bg-gray-50 p-4 rounded-lg">
                <p class="text-xs text-gray-600 mb-1">Utilization</p>
                <p 
                  class="text-2xl font-semibold"
                  :class="(selectedGroupAssignments 
                    ? getAggregatedUtilization(selectedGroupAssignments, selectedWeek) 
                    : getUtilization(selectedAssignment, selectedWeek)) > 100 ? 'text-red-600' : 'text-blue-600'"
                >
                  {{ (selectedGroupAssignments 
                    ? getAggregatedUtilization(selectedGroupAssignments, selectedWeek) 
                    : getUtilization(selectedAssignment, selectedWeek)).toFixed(0) 
                  }}%
                </p>
              </div>
            </div>

            <!-- Time Entries -->
            <div v-if="weekEntries.length > 0" class="border-t pt-4">
              <h4 class="text-sm font-semibold text-gray-800 mb-3">Time Entries</h4>
              <div class="overflow-x-auto">
                <table class="min-w-full divide-y divide-gray-200">
                  <thead class="bg-gray-50">
                    <tr>
                      <th scope="col" class="px-3 py-2 text-left text-xs font-medium text-gray-600 uppercase tracking-wider">Date</th>
                      <th scope="col" class="px-3 py-2 text-left text-xs font-medium text-gray-600 uppercase tracking-wider">Description</th>
                      <th scope="col" class="px-3 py-2 text-right text-xs font-medium text-gray-600 uppercase tracking-wider">Hours</th>
                    </tr>
                  </thead>
                  <tbody class="bg-white divide-y divide-gray-200">
                    <tr v-for="entry in weekEntries" :key="entry.ID" class="hover:bg-gray-50">
                      <td class="px-3 py-2 whitespace-nowrap text-xs text-gray-700">
                        {{ formatEntryDate(entry.start) }}
                      </td>
                      <td class="px-3 py-2 text-xs text-gray-700">
                        {{ entry.notes || 'No description' }}
                      </td>
                      <td class="px-3 py-2 whitespace-nowrap text-right text-xs font-medium text-gray-800">
                        {{ (entry.duration_minutes / 60).toFixed(2) }}h
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
            <div v-else class="border-t pt-4">
              <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
                <p class="text-sm text-yellow-800">
                  <i class="fas fa-info-circle mr-2"></i>
                  No time entries recorded for this week
                </p>
              </div>
            </div>

            <!-- Status Message -->
            <div 
              v-if="(selectedGroupAssignments 
                ? getAggregatedUtilization(selectedGroupAssignments, selectedWeek) 
                : getUtilization(selectedAssignment, selectedWeek)) > 100"
              class="bg-red-50 border border-red-200 rounded-lg p-3"
            >
              <p class="text-sm text-red-800">
                <i class="fas fa-exclamation-triangle mr-2"></i>
                Over-utilized: {{ (selectedGroupAssignments 
                  ? getAggregatedActualHours(selectedGroupAssignments, selectedWeek) - getAggregatedCommitment(selectedGroupAssignments, selectedWeek)
                  : getActualHours(selectedAssignment, selectedWeek) - getCommitmentForWeek(selectedAssignment, selectedWeek)).toFixed(1) 
                }}h over commitment
              </p>
            </div>
          </div>

          <div class="mt-6 flex justify-end">
            <button 
              @click="closeDetailModal"
              class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm font-semibold"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';

interface CommitmentSegment {
  start_date: string;
  end_date: string;
  commitment: number;
}

interface WeeklyUtilization {
  week_start: string;
  actual_hours: number;
  commitment: number;
  utilization: number;
}

interface Entry {
  ID: number;
  start: string;
  notes: string;
  duration_minutes: number;
}

interface CapacityAssignment {
  ID: number;
  employee_id: number;
  project_id: number;
  commitment: number;
  start_date: string;
  end_date: string;
  segments?: CommitmentSegment[];
  weekly_utilization: Record<string, WeeklyUtilization>;
  entries?: Entry[];
  employee: {
    first_name: string;
    last_name: string;
  };
  project: {
    name: string;
    account?: {
      name: string;
    };
  };
}

const isLoading = ref(true);
const error = ref<string | null>(null);
const assignments = ref<CapacityAssignment[]>([]);
const showDetailModal = ref(false);
const selectedAssignment = ref<CapacityAssignment | null>(null);
const selectedWeek = ref<Date | null>(null);
const expandedGroups = ref<Set<string>>(new Set());
const selectedGroupAssignments = ref<CapacityAssignment[] | null>(null); // For aggregated view

// Calculate date range: 12 weeks past, current week, 12 weeks future
const getWeeks = () => {
  const weeks: Date[] = [];
  const today = new Date();
  
  // Work in UTC to match backend
  const currentWeekStart = new Date(Date.UTC(
    today.getUTCFullYear(),
    today.getUTCMonth(),
    today.getUTCDate()
  ));
  
  // Move to start of week (Sunday)
  const dayOfWeek = currentWeekStart.getUTCDay();
  currentWeekStart.setUTCDate(currentWeekStart.getUTCDate() - dayOfWeek);
  
  // Go back 12 weeks, then forward 25 weeks total
  for (let i = -12; i <= 12; i++) {
    const weekStart = new Date(currentWeekStart);
    weekStart.setUTCDate(currentWeekStart.getUTCDate() + (i * 7));
    weeks.push(weekStart);
  }
  
  return weeks;
};

const weeks = ref<Date[]>(getWeeks());

// Check if a date falls within a range
const dateInRange = (date: Date, start: string, end: string): boolean => {
  const dateTime = date.getTime();
  const startTime = new Date(start).getTime();
  const endTime = new Date(end).getTime();
  return dateTime >= startTime && dateTime <= endTime;
};

// Group assignments by project (first tier), then by staff (second tier)
const groupedData = computed(() => {
  // Group by project, with staff assignments
  const projects = new Map<number, { 
    name: string; 
    accountName: string;
    staffAssignments: Map<number, { staffName: string; assignments: CapacityAssignment[] }>
  }>();
  
  assignments.value.forEach(assignment => {
    if (!projects.has(assignment.project_id)) {
      projects.set(assignment.project_id, {
        name: assignment.project.name,
        accountName: assignment.project.account?.name || '',
        staffAssignments: new Map()
      });
    }
    
    const projectData = projects.get(assignment.project_id)!;
    if (!projectData.staffAssignments.has(assignment.employee_id)) {
      projectData.staffAssignments.set(assignment.employee_id, {
        staffName: `${assignment.employee.first_name} ${assignment.employee.last_name}`,
        assignments: []
      });
    }
    projectData.staffAssignments.get(assignment.employee_id)!.assignments.push(assignment);
  });
  
  // Convert to array structure
  const result: Array<{ 
    type: 'group' | 'item'; 
    groupKey?: string;
    groupName?: string; 
    id?: number; 
    label?: string; 
    assignments?: CapacityAssignment[];
    allAssignments?: CapacityAssignment[];
  }> = [];
  
  Array.from(projects.entries())
    .sort((a, b) => a[1].name.localeCompare(b[1].name))
    .forEach(([projectId, projectData]) => {
      const groupKey = `project-${projectId}`;
      const allAssignments: CapacityAssignment[] = [];
      
      // Collect all assignments for this project
      projectData.staffAssignments.forEach(staff => {
        allAssignments.push(...staff.assignments);
      });
      
      // Check if this project has any commitments in the visible period
      const hasCommitments = weeks.value.some(week => 
        getAggregatedCommitment(allAssignments, week) > 0
      );
      
      // Only add project if it has commitments in visible period
      if (!hasCommitments) return;
      
      // Add project group row (with aggregated data)
      result.push({
        type: 'group',
        groupKey,
        groupName: projectData.name,
        allAssignments
      });
      
      // If expanded, add individual staff members
      if (expandedGroups.value.has(groupKey)) {
        Array.from(projectData.staffAssignments.entries())
          .sort((a, b) => a[1].staffName.localeCompare(b[1].staffName))
          .forEach(([employeeId, staff]) => {
            result.push({
              type: 'item',
              id: employeeId,
              label: staff.staffName,
              assignments: staff.assignments
            });
          });
      }
    });
  
  return result;
});

// Get commitment for a specific week
const getCommitmentForWeek = (assignment: CapacityAssignment, week: Date): number => {
  if (assignment.segments && assignment.segments.length > 0) {
    for (const segment of assignment.segments) {
      const segStart = new Date(segment.start_date);
      const segEnd = new Date(segment.end_date);
      
      if (week >= segStart && week <= segEnd) {
        return segment.commitment;
      }
    }
    return 0;
  }
  
  return assignment.commitment;
};

const getBarHeight = (commitment: number): string => {
  const minHeight = 20;
  const maxHeight = 40;
  const minCommitment = 5;
  const maxCommitment = 40;
  
  const normalized = Math.min(Math.max(commitment, minCommitment), maxCommitment);
  const height = minHeight + ((normalized - minCommitment) / (maxCommitment - minCommitment)) * (maxHeight - minHeight);
  
  return `${Math.round(height)}px`;
};

// Helper to get week key
const getWeekKey = (date: Date): string => {
  const year = date.getUTCFullYear();
  const month = String(date.getUTCMonth() + 1).padStart(2, '0');
  const day = String(date.getUTCDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
};

// Get utilization for a specific week
const getUtilization = (assignment: CapacityAssignment, week: Date): number => {
  const weekKey = getWeekKey(week);
  const util = assignment.weekly_utilization?.[weekKey];
  return util?.utilization || 0;
};

// Get actual hours for a specific week
const getActualHours = (assignment: CapacityAssignment, week: Date): number => {
  const weekKey = getWeekKey(week);
  const util = assignment.weekly_utilization?.[weekKey];
  return util?.actual_hours || 0;
};

// Get bar color based on utilization
const getBarColor = (utilization: number): string => {
  if (utilization > 100) return '#dc2626'; // red-600 for over-utilized
  return '#58837e'; // sage green
};

// Get fill percentage for two-tone effect
const getFillPercentage = (utilization: number): number => {
  return Math.min(utilization, 100);
};

// Get aggregated commitment for a group of assignments
const getAggregatedCommitment = (assignments: CapacityAssignment[], week: Date): number => {
  return assignments.reduce((total, assignment) => {
    if (dateInRange(week, assignment.start_date, assignment.end_date)) {
      return total + getCommitmentForWeek(assignment, week);
    }
    return total;
  }, 0);
};

// Get aggregated actual hours for a group of assignments
const getAggregatedActualHours = (assignments: CapacityAssignment[], week: Date): number => {
  return assignments.reduce((total, assignment) => {
    return total + getActualHours(assignment, week);
  }, 0);
};

// Get aggregated utilization for a group of assignments
const getAggregatedUtilization = (assignments: CapacityAssignment[], week: Date): number => {
  const commitment = getAggregatedCommitment(assignments, week);
  const actual = getAggregatedActualHours(assignments, week);
  
  if (commitment === 0) return 0;
  return (actual / commitment) * 100;
};

// Toggle group expansion
const toggleGroup = (groupKey: string) => {
  if (expandedGroups.value.has(groupKey)) {
    expandedGroups.value.delete(groupKey);
  } else {
    expandedGroups.value.add(groupKey);
  }
};

// Format date for display
const formatDate = (date: Date): string => {
  const month = date.toLocaleDateString('en-US', { month: 'short' });
  const day = date.getDate();
  return `${month} ${day}`;
};

// Check if week is current week
const isCurrentWeek = (weekStart: Date): boolean => {
  const today = new Date();
  
  const currentWeekStart = new Date(Date.UTC(
    today.getUTCFullYear(),
    today.getUTCMonth(),
    today.getUTCDate()
  ));
  const dayOfWeek = currentWeekStart.getUTCDay();
  currentWeekStart.setUTCDate(currentWeekStart.getUTCDate() - dayOfWeek);
  
  return weekStart.getTime() === currentWeekStart.getTime();
};

// Get entries for the selected week (from single assignment or all assignments in group)
const weekEntries = computed(() => {
  if (!selectedWeek.value) return [];
  
  const weekStart = new Date(selectedWeek.value);
  const weekEnd = new Date(weekStart);
  weekEnd.setDate(weekEnd.getDate() + 7);
  
  // If viewing aggregated group, get entries from all assignments
  if (selectedGroupAssignments.value) {
    const allEntries: Entry[] = [];
    selectedGroupAssignments.value.forEach(assignment => {
      (assignment.entries || []).forEach(entry => {
        const entryDate = new Date(entry.start);
        if (entryDate >= weekStart && entryDate < weekEnd) {
          allEntries.push(entry);
        }
      });
    });
    // Sort by date
    return allEntries.sort((a, b) => new Date(a.start).getTime() - new Date(b.start).getTime());
  }
  
  // Single assignment view
  if (!selectedAssignment.value) return [];
  return (selectedAssignment.value.entries || []).filter(entry => {
    const entryDate = new Date(entry.start);
    return entryDate >= weekStart && entryDate < weekEnd;
  });
});

// Handle bar click
const handleBarClick = (assignment: CapacityAssignment, week: Date) => {
  selectedAssignment.value = assignment;
  selectedWeek.value = week;
  selectedGroupAssignments.value = null; // Clear group assignments
  showDetailModal.value = true;
};

// Handle aggregated bar click (show all assignments for the group)
const handleAggregatedBarClick = (allAssignments: CapacityAssignment[], week: Date) => {
  // For aggregated view, show combined data from all assignments
  if (allAssignments.length > 0) {
    selectedAssignment.value = allAssignments[0]; // Use first for context (title display)
    selectedGroupAssignments.value = allAssignments; // Store all for entries
    selectedWeek.value = week;
    showDetailModal.value = true;
  }
};

// Close modal
const closeDetailModal = () => {
  showDetailModal.value = false;
  selectedAssignment.value = null;
  selectedWeek.value = null;
  selectedGroupAssignments.value = null;
};

// Format entry date
const formatEntryDate = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', { 
    weekday: 'short',
    month: 'short', 
    day: 'numeric' 
  });
};

// Scroll to current week
const scrollToCurrentWeek = () => {
  setTimeout(() => {
    const currentWeekElement = document.querySelector('.gantt-week-header.current-week');
    if (currentWeekElement) {
      currentWeekElement.scrollIntoView({ 
        behavior: 'smooth', 
        block: 'nearest', 
        inline: 'center' 
      });
    }
  }, 100);
};

// Fetch capacity data
const loadCapacityData = async () => {
  isLoading.value = true;
  error.value = null;
  
  try {
    const token = localStorage.getItem('snowpack_token');
    if (!token) {
      throw new Error('Authentication token not found. Please log in again.');
    }
    
    const response = await fetch('/api/portal/capacity', {
      headers: { 'x-access-token': token },
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ message: response.statusText }));
      throw new Error(`Failed to fetch capacity data: ${errorData.message || response.statusText}`);
    }
    
    assignments.value = await response.json();
  } catch (err: any) {
    console.error('Error loading capacity data:', err);
    error.value = err.message || 'Failed to load capacity data. Please try again.';
  } finally {
    isLoading.value = false;
  }
};

onMounted(async () => {
  await loadCapacityData();
  scrollToCurrentWeek();
});
</script>

<style scoped>
.gantt-container {
  min-width: 100%;
}

.gantt-header {
  display: flex;
  background-color: #f9fafb;
  border-bottom: 1px solid #d1d5db;
  position: sticky;
  top: 0;
  z-index: 10;
}

.gantt-group-row {
  display: flex;
  background-color: #f9fafb;
  border-bottom: 1px solid #d1d5db;
  cursor: pointer;
}

.gantt-group-row:hover {
  background-color: #f3f4f6;
}

.gantt-row {
  display: flex;
  border-bottom: 1px solid #e5e7eb;
}

.gantt-row-child {
  background-color: #ffffff;
}

.gantt-row-child:hover {
  background-color: #fafafa;
}

.gantt-label-column {
  width: 200px;
  min-width: 200px;
  border-right: 2px solid #e5e7eb;
  background-color: white;
  position: sticky;
  left: 0;
  z-index: 5;
}

.gantt-header .gantt-label-column {
  z-index: 15;
}

.gantt-timeline {
  display: flex;
  flex: 1;
}

.gantt-week-header {
  flex: 1;
  min-width: 80px;
  padding: 8px 4px 4px 4px;
  text-align: center;
  font-size: 0.75rem;
  font-weight: 600;
  color: #4b5563;
  position: relative;
}

.gantt-week-header.current-week {
  background-color: #b8d4d0;
  color: #1f3d38;
  font-weight: 700;
  border-left: 4px solid #58837e;
  border-right: 4px solid #58837e;
}

.week-label {
  margin-bottom: 2px;
}

.current-week-indicator {
  color: #58837e;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.gantt-cell {
  flex: 1;
  min-width: 80px;
  min-height: 55px;
  padding: 2px 0;
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  align-items: stretch;
  gap: 1px;
}

.gantt-cell.current-week {
  background-color: #b8d4d0;
  border-left: 4px solid #58837e;
  border-right: 4px solid #58837e;
  padding: 2px 2px;
}

.gantt-bar {
  background-color: #58837e;
  color: white;
  padding: 0;
  border-radius: 2px;
  margin: 0;
  font-size: 0.65rem;
  font-weight: 500;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.gantt-bar:hover {
  background-color: #476b67;
  transform: scale(1.03);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2);
  z-index: 5;
}

.gantt-bar-aggregated {
  opacity: 0.9;
}

.gantt-bar-text {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: block;
}
</style>

