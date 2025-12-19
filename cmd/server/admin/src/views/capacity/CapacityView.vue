<script setup lang="ts">
import { ref, onMounted, computed, watch, nextTick } from 'vue';
import { fetchCapacityData, fetchCapacityEntries, type CapacityAssignment, type EntryDetail } from '../../api/capacity';

// State
const isLoading = ref(true);
const error = ref<string | null>(null);
const assignments = ref<CapacityAssignment[]>([]);
const viewMode = ref<'project' | 'staff'>('project');
const selectedAssignment = ref<CapacityAssignment | null>(null);
const selectedWeek = ref<Date | null>(null);
const showDetailModal = ref(false);
const expandedGroups = ref<Set<string>>(new Set());
const selectedGroupAssignments = ref<CapacityAssignment[] | null>(null); // For aggregated view
const modalEntries = ref<EntryDetail[]>([]); // Detailed entries for modal
const modalElement = ref<HTMLElement | null>(null); // For focusing modal on open

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

// Toggle group expansion
const toggleGroup = (groupKey: string) => {
  if (expandedGroups.value.has(groupKey)) {
    expandedGroups.value.delete(groupKey);
  } else {
    expandedGroups.value.add(groupKey);
  }
  // Force reactivity
  expandedGroups.value = new Set(expandedGroups.value);
};

// Group assignments by account (for projects) or by staff (with their projects)
const groupedData = computed(() => {
  if (viewMode.value === 'project') {
    // Group projects by account
    const accountGroups = new Map<string, Map<number, { name: string; assignments: CapacityAssignment[] }>>();
    
    assignments.value.forEach(assignment => {
      const accountName = assignment.project.account?.name || 'Unknown';
      
      if (!accountGroups.has(accountName)) {
        accountGroups.set(accountName, new Map());
      }
      
      const projectsInAccount = accountGroups.get(accountName)!;
      if (!projectsInAccount.has(assignment.project_id)) {
        projectsInAccount.set(assignment.project_id, {
          name: assignment.project.name,
          assignments: []
        });
      }
      projectsInAccount.get(assignment.project_id)!.assignments.push(assignment);
    });
    
    // Convert to array structure with account groups
    const result: Array<{ 
      type: 'group' | 'item'; 
      groupKey?: string;
      groupName?: string; 
      id?: number; 
      label?: string; 
      assignments?: CapacityAssignment[];
      allAssignments?: CapacityAssignment[]; // For aggregated bars
    }> = [];
    
    Array.from(accountGroups.entries())
      .sort((a, b) => a[0].localeCompare(b[0]))
      .forEach(([accountName, projects]) => {
        const groupKey = `account-${accountName}`;
        const allAssignmentsInAccount: CapacityAssignment[] = [];
        
        // Collect all assignments for this account
        projects.forEach(project => {
          allAssignmentsInAccount.push(...project.assignments);
        });
        
        // Check if this group has any commitments in the visible period
        const hasCommitments = weeks.value.some(week => 
          getAggregatedCommitment(allAssignmentsInAccount, week) > 0
        );
        
        // Only add group if it has commitments in visible period
        if (!hasCommitments) return;
        
        // Add account group row (with aggregated data)
        result.push({ 
          type: 'group', 
          groupKey,
          groupName: accountName,
          allAssignments: allAssignmentsInAccount
        });
        
        // If expanded, add individual projects
        if (expandedGroups.value.has(groupKey)) {
          Array.from(projects.entries())
            .sort((a, b) => a[1].name.localeCompare(b[1].name))
            .forEach(([id, data]) => {
              result.push({
                type: 'item',
                id,
                label: data.name,
                assignments: data.assignments
              });
            });
        }
      });
    
    return result;
  } else {
    // For staff view, group by staff with their projects
    const staffGroups = new Map<number, { 
      name: string; 
      projectAssignments: Map<number, { projectName: string; accountName: string; assignments: CapacityAssignment[] }>
    }>();
    
    assignments.value.forEach(assignment => {
      if (!staffGroups.has(assignment.employee_id)) {
        staffGroups.set(assignment.employee_id, {
          name: `${assignment.employee.first_name} ${assignment.employee.last_name}`,
          projectAssignments: new Map()
        });
      }
      
      const staffData = staffGroups.get(assignment.employee_id)!;
      if (!staffData.projectAssignments.has(assignment.project_id)) {
        staffData.projectAssignments.set(assignment.project_id, {
          projectName: assignment.project.name,
          accountName: assignment.project.account?.name || 'Unknown',
          assignments: []
        });
      }
      staffData.projectAssignments.get(assignment.project_id)!.assignments.push(assignment);
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
    
    Array.from(staffGroups.entries())
      .sort((a, b) => a[1].name.localeCompare(b[1].name))
      .forEach(([employeeId, staffData]) => {
        const groupKey = `staff-${employeeId}`;
        const allAssignments: CapacityAssignment[] = [];
        
        // Collect all assignments for this staff member
        staffData.projectAssignments.forEach(project => {
          allAssignments.push(...project.assignments);
        });
        
        // Check if this staff member has any commitments in the visible period
        const hasCommitments = weeks.value.some(week => 
          getAggregatedCommitment(allAssignments, week) > 0
        );
        
        // Only add staff if they have commitments in visible period
        if (!hasCommitments) return;
        
        // Add staff group row (with aggregated data)
        result.push({
          type: 'group',
          groupKey,
          groupName: staffData.name,
          allAssignments
        });
        
        // If expanded, add individual projects
        if (expandedGroups.value.has(groupKey)) {
          Array.from(staffData.projectAssignments.entries())
            .sort((a, b) => a[1].projectName.localeCompare(b[1].projectName))
            .forEach(([projectId, project]) => {
              result.push({
                type: 'item',
                id: projectId,
                label: `${project.projectName} (${project.accountName})`,
                assignments: project.assignments
              });
            });
        }
      });
    
    return result;
  }
});

// Calculate totals per week from ALL assignments (not just visible ones)
const weeklyTotals = computed(() => {
  return weeks.value.map(week => {
    let committed = 0;
    let billed = 0;
    
    // Iterate through ALL assignments directly, not the grouped data
    assignments.value.forEach(assignment => {
      if (dateInRange(week, assignment.start_date, assignment.end_date)) {
        committed += getCommitmentForWeek(assignment, week);
        billed += getActualHours(assignment, week);
      }
    });
    
    const utilization = committed > 0 ? (billed / committed) * 100 : 0;
    
    return {
      committed,
      billed,
      utilization
    };
  });
});

// Get commitment for a specific week, using segments if available
const getCommitmentForWeek = (assignment: CapacityAssignment, week: Date): number => {
  // If we have segments, find the one that contains this week
  if (assignment.segments && assignment.segments.length > 0) {
    // Get week as YYYY-MM-DD string for comparison
    const weekStr = getWeekKey(week);
    
    for (const segment of assignment.segments) {
      // Compare date strings directly to avoid timezone issues
      if (weekStr >= segment.start_date && weekStr <= segment.end_date) {
        return segment.commitment;
      }
    }
    return 0; // Week not in any segment
  }
  
  // Fallback to simple commitment
  return assignment.commitment;
};

// Get aggregated commitment for a week across multiple assignments
const getAggregatedCommitment = (assignments: CapacityAssignment[], week: Date): number => {
  let total = 0;
  assignments.forEach(assignment => {
    if (dateInRange(week, assignment.start_date, assignment.end_date)) {
      total += getCommitmentForWeek(assignment, week);
    }
  });
  return total;
};

// Get aggregated actual hours for a week across multiple assignments
const getAggregatedActualHours = (assignments: CapacityAssignment[], week: Date): number => {
  let total = 0;
  const details: any[] = [];
  assignments.forEach(assignment => {
    const hours = getActualHours(assignment, week);
    if (hours > 0) {
      details.push({
        assignment_id: assignment.ID,
        project: assignment.project?.name,
        employee: `${assignment.employee?.first_name} ${assignment.employee?.last_name}`,
        hours
      });
    }
    total += hours;
  });
  
  // Debug: Log if there's a discrepancy
  if (details.length > 1 && total > 0) {
    console.log(`Week ${getWeekKey(week)} aggregated hours:`, total, 'from', details);
  }
  
  return total;
};

// Get aggregated utilization for a week across multiple assignments
const getAggregatedUtilization = (assignments: CapacityAssignment[], week: Date): number => {
  const totalCommitment = getAggregatedCommitment(assignments, week);
  const totalActual = getAggregatedActualHours(assignments, week);
  return totalCommitment > 0 ? (totalActual / totalCommitment) * 100 : 0;
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

// Helper to get week key in YYYY-MM-DD format (UTC normalized)
const getWeekKey = (date: Date): string => {
  // Get UTC components to avoid timezone shifts
  const year = date.getUTCFullYear();
  const month = String(date.getUTCMonth() + 1).padStart(2, '0');
  const day = String(date.getUTCDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
};

// Get utilization for a specific week
const getUtilization = (assignment: CapacityAssignment, week: Date): number => {
  const weekKey = getWeekKey(week);
  const util = assignment.weekly_utilization?.[weekKey];
  
  // Debug logging
  if (!util && getCommitmentForWeek(assignment, week) > 0) {
    const availableKeys = Object.keys(assignment.weekly_utilization || {});
    if (availableKeys.length > 0) {
      console.log('Week key mismatch! Looking for:', weekKey, 'but have:', availableKeys.slice(0, 3));
      console.log('Week date:', week, 'UTC:', week.toISOString());
    }
  }
  
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
  return '#58837e'; // default sage green
};

// Get fill percentage for two-tone effect
const getFillPercentage = (utilization: number): number => {
  return Math.min(utilization, 100);
};

// Handle bar click to show details

const handleBarClick = async (assignment: CapacityAssignment, week: Date) => {
  selectedAssignment.value = assignment;
  selectedWeek.value = week;
  selectedGroupAssignments.value = null; // Clear group assignments
  showDetailModal.value = true;
  
  // Fetch detailed entries for the selected assignment and week
  try {
    modalEntries.value = await fetchCapacityEntries(assignment.ID, getWeekKey(week));
  } catch (err) {
    console.error("Failed to fetch detailed entries:", err);
    modalEntries.value = [];
  }
};

// Handle aggregated bar click (show all assignments for the group)
const handleAggregatedBarClick = async (allAssignments: CapacityAssignment[], week: Date) => {
  // For aggregated view, show combined data from all assignments
  if (allAssignments.length > 0) {
    selectedAssignment.value = allAssignments[0]; // Use first for context (title display)
    selectedGroupAssignments.value = allAssignments; // Store all for entries
    selectedWeek.value = week;
    showDetailModal.value = true;

    // Fetch detailed entries for all assignments in the group for the selected week
    const fetchedEntries: EntryDetail[] = [];
    for (const assignment of allAssignments) {
      try {
        const entries = await fetchCapacityEntries(assignment.ID, getWeekKey(week));
        fetchedEntries.push(...entries);
      } catch (err) {
        console.error(`Failed to fetch detailed entries for assignment ${assignment.ID}:`, err);
      }
    }
    // Sort entries by date
    modalEntries.value = fetchedEntries.sort((a, b) => new Date(a.start).getTime() - new Date(b.start).getTime());
  }
};

// Close detail modal
const closeDetailModal = () => {
  showDetailModal.value = false;
  selectedAssignment.value = null;
  selectedWeek.value = null;
  selectedGroupAssignments.value = null;
  modalEntries.value = [];
};

// Watch for modal open and focus it for keyboard events
watch(showDetailModal, (isOpen) => {
  if (isOpen) {
    nextTick(() => {
      modalElement.value?.focus();
    });
  }
});

// Format entry date
const formatEntryDate = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', { 
    weekday: 'short',
    month: 'short', 
    day: 'numeric' 
  });
};

// Format date for display
const formatDate = (date: Date): string => {
  // Use UTC methods since weeks are generated in UTC
  const month = date.toLocaleDateString('en-US', { month: 'short', timeZone: 'UTC' });
  const day = date.getUTCDate();
  return `${month} ${day}`;
};

// Check if week is current week (using UTC to match our week generation)
const isCurrentWeek = (weekStart: Date): boolean => {
  const today = new Date();
  
  // Get current week start in UTC (matching getWeeks logic)
  const currentWeekStart = new Date(Date.UTC(
    today.getUTCFullYear(),
    today.getUTCMonth(),
    today.getUTCDate()
  ));
  const dayOfWeek = currentWeekStart.getUTCDay();
  currentWeekStart.setUTCDate(currentWeekStart.getUTCDate() - dayOfWeek);
  
  // Compare UTC timestamps
  return weekStart.getTime() === currentWeekStart.getTime();
};

// Fetch capacity data
const loadCapacityData = async () => {
  isLoading.value = true;
  error.value = null;
  
  try {
    const data = await fetchCapacityData();
    assignments.value = data;
    
    // Debug: Check if utilization data is present
    console.log('Loaded assignments:', data.length);
    if (data.length > 0) {
      console.log('Sample assignment:', data[0]);
      console.log('Weekly utilization keys:', Object.keys(data[0].weekly_utilization || {}));
      console.log('Sample week data:', Object.values(data[0].weekly_utilization || {})[0]);
    }
  } catch (err) {
    console.error('Error loading capacity data:', err);
    error.value = 'Failed to load capacity data. Please try again.';
  } finally {
    isLoading.value = false;
  }
};

// Scroll to current week
const scrollToCurrentWeek = () => {
  // Find current week index
  const currentWeekIndex = weeks.value.findIndex(week => isCurrentWeek(week));
  
  if (currentWeekIndex >= 0) {
    // Find the current week element and scroll it into view
    setTimeout(() => {
      const currentWeekElement = document.querySelector('.gantt-week-header.current-week');
      if (currentWeekElement) {
        currentWeekElement.scrollIntoView({ 
          behavior: 'smooth', 
          block: 'nearest', 
          inline: 'center' 
        });
      }
    }, 100); // Small delay to ensure DOM is rendered
  }
};

// Initialize component
onMounted(async () => {
  await loadCapacityData();
  scrollToCurrentWeek();
});
</script>

<template>
  <div class="px-4 py-4 sm:px-6 lg:px-8">
    <div class="sm:flex sm:items-center">
      <div class="sm:flex-auto">
        <h1 class="text-base font-semibold leading-6 text-blue">Capacity Management</h1>
      </div>
      <div class="mt-3 sm:ml-16 sm:mt-0 sm:flex-none">
        <div class="inline-flex rounded-md shadow-sm" role="group">
          <button
            type="button"
            @click="viewMode = 'project'"
            :class="[
              viewMode === 'project'
                ? 'bg-sage text-white'
                : 'bg-white text-gray-dark hover:bg-gray-50',
              'px-3 py-1.5 text-xs font-semibold rounded-l-md border border-gray-300'
            ]"
          >
            By Project
          </button>
          <button
            type="button"
            @click="viewMode = 'staff'"
            :class="[
              viewMode === 'staff'
                ? 'bg-sage text-white'
                : 'bg-white text-gray-dark hover:bg-gray-50',
              'px-3 py-1.5 text-xs font-semibold rounded-r-md border border-gray-300 border-l-0'
            ]"
          >
            By Staff
          </button>
        </div>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="isLoading" class="flex flex-col items-center justify-center p-6 bg-white rounded-lg shadow mt-4">
      <i class="fas fa-spinner fa-spin text-3xl text-teal mb-2"></i>
      <span class="text-sm text-gray-dark">Loading capacity data...</span>
    </div>
    
    <!-- Error state -->
    <div v-else-if="error" class="flex flex-col items-center justify-center p-6 bg-white rounded-lg shadow mt-4">
      <i class="fas fa-exclamation-circle text-3xl text-red mb-2"></i>
      <span class="text-sm text-gray-dark mb-1">{{ error }}</span>
      <button @click="loadCapacityData" class="mt-3 px-3 py-1.5 text-xs font-semibold text-white bg-sage rounded-md hover:bg-sage-dark">
        Retry
      </button>
    </div>
    
    <!-- Gantt Chart -->
    <div v-else class="mt-4 bg-white rounded-lg shadow overflow-hidden">
      <div class="overflow-x-auto">
        <div class="inline-block min-w-full align-middle">
          <div class="gantt-container">
            <!-- Header Row -->
            <div class="gantt-header">
              <div class="gantt-label-column">
                <div class="p-2 text-xs font-semibold text-gray-dark">
                  {{ viewMode === 'project' ? 'Project' : 'Staff Member' }}
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
              <!-- Group Row (Account or Staff with aggregated data) -->
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
                    <!-- Aggregate all assignments in this group for this week -->
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
              
              <!-- Project/Assignment Item Row (shown when group is expanded) -->
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
                        :title="`${viewMode === 'project' ? assignment.employee.first_name + ' ' + assignment.employee.last_name : assignment.project.name} - ${getCommitmentForWeek(assignment, week)}h/week (${getActualHours(assignment, week).toFixed(1)}h actual, ${getUtilization(assignment, week).toFixed(0)}%)`"
                        @click.stop="handleBarClick(assignment, week)"
                      >
                        <span class="gantt-bar-text">{{ getCommitmentForWeek(assignment, week) }}h</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </template>
            
            <!-- Totals Rows -->
            <div class="gantt-totals-row">
              <div class="gantt-label-column">
                <div class="p-2 text-xs font-semibold text-gray-dark">Committed</div>
              </div>
              <div class="gantt-timeline">
                <div
                  v-for="(week, index) in weeks"
                  :key="index"
                  :class="[
                    'gantt-total-cell',
                    isCurrentWeek(week) ? 'current-week' : ''
                  ]"
                >
                  <span class="text-xs font-semibold text-gray-dark">{{ weeklyTotals[index].committed }}h</span>
                </div>
              </div>
            </div>
            
            <div class="gantt-totals-row">
              <div class="gantt-label-column">
                <div class="p-2 text-xs font-semibold text-gray-dark">Billed</div>
              </div>
              <div class="gantt-timeline">
                <div
                  v-for="(week, index) in weeks"
                  :key="index"
                  :class="[
                    'gantt-total-cell',
                    isCurrentWeek(week) ? 'current-week' : ''
                  ]"
                >
                  <span class="text-xs font-semibold text-gray-dark">{{ weeklyTotals[index].billed.toFixed(1) }}h</span>
                </div>
              </div>
            </div>
            
            <div class="gantt-totals-row">
              <div class="gantt-label-column">
                <div class="p-2 text-xs font-semibold text-gray-dark">Utilization</div>
              </div>
              <div class="gantt-timeline">
                <div
                  v-for="(week, index) in weeks"
                  :key="index"
                  :class="[
                    'gantt-total-cell',
                    isCurrentWeek(week) ? 'current-week' : ''
                  ]"
                >
                  <span 
                    class="text-xs font-semibold"
                    :class="weeklyTotals[index].utilization > 100 ? 'text-red-600' : 'text-gray-dark'"
                  >
                    {{ weeklyTotals[index].utilization.toFixed(0) }}%
                  </span>
                </div>
              </div>
            </div>

            <!-- Empty State -->
            <div v-if="groupedData.length === 0" class="p-8 text-center">
              <i class="fas fa-calendar-times text-4xl text-gray mb-3"></i>
              <p class="text-sm text-gray-dark">No capacity assignments found</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Detail Modal -->
    <div v-if="showDetailModal" ref="modalElement" class="fixed inset-0 z-50 flex items-center justify-center px-4" @click="closeDetailModal" @keydown.esc="closeDetailModal" tabindex="-1">
      <div class="fixed inset-0 bg-black/30 backdrop-blur-sm transition-opacity"></div>
      
      <div class="relative bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] flex flex-col" @click.stop>
        <!-- Header -->
        <div class="bg-white px-6 py-4 border-b border-gray-200 rounded-t-lg">
          <div class="flex justify-between items-start">
            <div>
              <h3 class="text-lg font-semibold text-gray-900">
                {{ viewMode === 'project' ? selectedAssignment?.project.name : selectedAssignment?.employee.first_name + ' ' + selectedAssignment?.employee.last_name }}
              </h3>
              <p class="text-sm text-gray-600 mt-1">
                Week of {{ selectedWeek ? formatDate(selectedWeek) : '' }}
              </p>
            </div>
            <button @click="closeDetailModal" class="text-gray-400 hover:text-gray-600">
              <i class="fas fa-times text-xl"></i>
            </button>
          </div>
        </div>

        <!-- Scrollable Content -->
        <div class="overflow-y-auto flex-1 px-6 py-4" v-if="selectedAssignment && selectedWeek">
          <!-- Summary Stats -->
          <div class="grid grid-cols-3 gap-4 mb-4">
            <div class="bg-gray-50 border border-gray-200 p-3 rounded-lg">
              <p class="text-xs text-gray-600 mb-1">Commitment</p>
              <p class="text-2xl font-semibold text-gray-900">
                {{ selectedGroupAssignments 
                  ? getAggregatedCommitment(selectedGroupAssignments, selectedWeek) 
                  : getCommitmentForWeek(selectedAssignment, selectedWeek) 
                }}h
              </p>
            </div>
            <div class="bg-gray-50 border border-gray-200 p-3 rounded-lg">
              <p class="text-xs text-gray-600 mb-1">Actual Hours</p>
              <p class="text-2xl font-semibold text-gray-900">
                {{ selectedGroupAssignments 
                  ? getAggregatedActualHours(selectedGroupAssignments, selectedWeek).toFixed(1) 
                  : getActualHours(selectedAssignment, selectedWeek).toFixed(1) 
                }}h
              </p>
            </div>
            <div class="bg-gray-50 border border-gray-200 p-3 rounded-lg">
              <p class="text-xs text-gray-600 mb-1">Utilization</p>
              <p 
                class="text-2xl font-semibold"
                :class="(selectedGroupAssignments 
                  ? getAggregatedUtilization(selectedGroupAssignments, selectedWeek) 
                  : getUtilization(selectedAssignment, selectedWeek)) > 100 ? 'text-red-600' : 'text-sage'"
              >
                {{ (selectedGroupAssignments 
                  ? getAggregatedUtilization(selectedGroupAssignments, selectedWeek) 
                  : getUtilization(selectedAssignment, selectedWeek)).toFixed(0) 
                }}%
              </p>
            </div>
          </div>

          <!-- Assignment Details -->
          <div class="border-t pt-4 mt-4">
              <h4 class="text-sm font-semibold text-gray-dark mb-2">Assignment Details</h4>
              <div class="space-y-2 text-sm">
                <div class="flex justify-between">
                  <span class="text-gray">{{ viewMode === 'project' ? 'Staff Member' : 'Project' }}:</span>
                  <span class="text-gray-dark font-medium">
                    {{ viewMode === 'project' 
                      ? `${selectedAssignment.employee.first_name} ${selectedAssignment.employee.last_name}`
                      : selectedAssignment.project.name
                    }}
                  </span>
                </div>
                <div v-if="viewMode === 'staff' && selectedAssignment.project.account" class="flex justify-between">
                  <span class="text-gray">Account:</span>
                  <span class="text-gray-dark font-medium">{{ selectedAssignment.project.account.name }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray">Assignment Period:</span>
                  <span class="text-gray-dark font-medium">
                    {{ new Date(selectedAssignment.start_date).toLocaleDateString() }} - {{ new Date(selectedAssignment.end_date).toLocaleDateString() }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Time Entries -->
            <div v-if="modalEntries.length > 0" class="border-t pt-4">
              <h4 class="text-sm font-semibold text-gray-dark mb-3">Time Entries</h4>
              <div class="overflow-x-auto">
                <table class="min-w-full divide-y divide-gray-200">
                  <thead class="bg-gray-50">
                    <tr>
                      <th scope="col" class="px-3 py-2 text-left text-xs font-medium text-gray uppercase tracking-wider">Date</th>
                      <th scope="col" class="px-3 py-2 text-left text-xs font-medium text-gray uppercase tracking-wider">Description</th>
                      <th scope="col" class="px-3 py-2 text-right text-xs font-medium text-gray uppercase tracking-wider">Hours</th>
                    </tr>
                  </thead>
                  <tbody class="bg-white divide-y divide-gray-200">
                    <tr v-for="entry in modalEntries" :key="entry.id" class="hover:bg-gray-50">
                      <td class="px-3 py-2 whitespace-nowrap text-xs text-gray-dark">
                        {{ formatEntryDate(entry.start) }}
                      </td>
                      <td class="px-3 py-2 text-xs text-gray-dark">
                        {{ entry.notes || 'No description' }}
                      </td>
                      <td class="px-3 py-2 whitespace-nowrap text-right text-xs font-medium text-gray-dark">
                        {{ (entry.duration_minutes / 60).toFixed(2) }}h
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          <div v-else class="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
            <p class="text-sm text-yellow-800">
              <i class="fas fa-info-circle mr-2"></i>
              No time entries recorded for this week.
            </p>
          </div>
        </div>

        <!-- Fixed Footer -->
        <div class="border-t border-gray-200 px-6 py-4 bg-gray-50 rounded-b-lg flex justify-end">
          <button 
            @click="closeDetailModal"
            class="px-4 py-2 bg-sage text-white rounded-md hover:bg-sage-dark text-sm font-semibold"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

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

.gantt-group-header {
  display: flex;
  background-color: #f9fafb;
  border-top: 2px solid #d1d5db;
  border-bottom: 1px solid #d1d5db;
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
  border-bottom: 1px solid #f3f4f6;
}

.gantt-row-child {
  background-color: #fafafa;
}

.gantt-row:hover, .gantt-row-child:hover {
  background-color: #f5f5f5;
}

.gantt-bar-aggregated {
  opacity: 0.9;
}

.gantt-totals-row {
  display: flex;
  background-color: #f9fafb;
  border-top: 1px solid #d1d5db;
  font-weight: 600;
  position: sticky;
  bottom: 0;
  z-index: 10;
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

.gantt-totals-row .gantt-label-column {
  z-index: 15;
  background-color: #f9fafb;
}

.gantt-total-column {
  width: 80px;
  min-width: 80px;
  border-left: 2px solid #e5e7eb;
  background-color: #f9fafb;
  display: flex;
  align-items: center;
  justify-content: center;
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

.gantt-total-cell {
  flex: 1;
  min-width: 80px;
  padding: 8px 4px;
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
}

.gantt-total-cell.current-week {
  background-color: #b8d4d0;
  border-left: 4px solid #58837e;
  border-right: 4px solid #58837e;
  font-weight: 700;
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

.gantt-bar-text {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: block;
}
</style>

