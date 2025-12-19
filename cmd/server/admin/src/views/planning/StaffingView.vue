<template>
  <div class="px-4 sm:px-6 lg:px-8">
    <div class="sm:flex sm:items-center mb-4">
      <div class="sm:flex-auto">
        <h1 class="text-xl font-semibold text-gray-900">Staffing Assignments</h1>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="mt-8 flex justify-center">
      <div class="flex items-center">
        <i class="fas fa-spinner fa-spin text-sage text-2xl mr-3"></i>
        <span class="text-gray-700">Loading staffing data...</span>
      </div>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="mt-8">
      <div class="rounded-md bg-red-50 p-4">
        <div class="flex">
          <div class="flex-shrink-0">
            <i class="fas fa-exclamation-circle text-red-400"></i>
          </div>
          <div class="ml-3">
            <h3 class="text-sm font-medium text-red-800">Error loading data</h3>
            <div class="mt-2 text-sm text-red-700">
              <p>{{ error }}</p>
            </div>
            <div class="mt-4">
              <button @click="loadData" class="text-sm font-medium text-red-800 hover:text-red-600">
                Try again
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <div v-else class="mt-4">
      <!-- Controls Row -->
      <div class="mb-3 flex items-center gap-3">
        <!-- Left: View Toggle -->
        <div class="flex items-center space-x-2">
          <button
            @click="viewMode = 'by-staff'"
            :class="[
              'px-3 py-1.5 text-xs font-medium rounded-md',
              viewMode === 'by-staff'
                ? 'bg-sage text-white'
                : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
            ]"
          >
            <i class="fas fa-users mr-1.5"></i>
            By Staff
          </button>
          <button
            @click="viewMode = 'by-project'"
            :class="[
              'px-3 py-1.5 text-xs font-medium rounded-md',
              viewMode === 'by-project'
                ? 'bg-sage text-white'
                : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
            ]"
          >
            <i class="fas fa-bars-progress mr-1.5"></i>
            By Project
          </button>
        </div>

        <!-- Center: Filters & Date Range -->
        <div class="flex items-center space-x-2">
          <!-- Status Filter -->
          <select
            v-model="statusFilter"
            class="border border-gray-300 rounded-md bg-white text-gray-700 w-24"
            style="padding: 4px 8px; font-size: 0.75rem; line-height: 1rem;"
          >
            <option value="all">All</option>
            <option value="active">Active</option>
            <option value="upcoming">Upcoming</option>
            <option value="ended">Ended</option>
          </select>

          <!-- Date Range -->
          <select
            v-model="selectedYear"
            @change="updateTimeline"
            class="border border-gray-300 rounded-md bg-white text-gray-700 w-20"
            style="padding: 4px 8px; font-size: 0.75rem; line-height: 1rem;"
          >
            <option v-for="year in availableYears" :key="year" :value="year">{{ year }}</option>
          </select>
        </div>

        <!-- Right: Legend -->
        <div class="flex items-center space-x-2 text-xs ml-auto">
          <div class="flex items-center space-x-1">
            <div class="w-2.5 h-2.5 rounded bg-green-100 border border-green-300"></div>
            <span class="text-gray-600">Active</span>
          </div>
          <div class="flex items-center space-x-1">
            <div class="w-2.5 h-2.5 rounded bg-blue-100 border border-blue-300"></div>
            <span class="text-gray-600">Upcoming</span>
          </div>
          <div class="flex items-center space-x-1">
            <div class="w-2.5 h-2.5 rounded bg-gray-100 border border-gray-300"></div>
            <span class="text-gray-600">Ended</span>
          </div>
        </div>
      </div>

      <!-- Gantt Chart Container with synchronized scrolling -->
      <div class="bg-white border border-gray-200 rounded-lg overflow-hidden">
        <div class="overflow-x-auto">
          <div style="min-width: fit-content;">
            <!-- Timeline Header -->
            <div class="flex border-b border-gray-200">
              <div class="w-64 flex-shrink-0 px-4 py-2 bg-gray-50 border-r border-gray-200">
                <span class="text-xs font-medium text-gray-700">{{ viewMode === 'by-staff' ? 'Staff Member' : 'Project' }}</span>
              </div>
              <div class="flex">
                <div v-for="period in timelinePeriods" :key="period.label" class="min-w-[80px] px-2 py-2 bg-gray-50 border-r border-gray-200 last:border-r-0">
                  <div class="text-xs font-medium text-gray-700 text-center">{{ period.label }}</div>
                </div>
              </div>
            </div>

            <!-- By Staff Member View (Gantt-style) -->
            <div v-if="viewMode === 'by-staff'">
              <div v-for="staff in filteredStaffMembers" :key="staff.ID" class="border-b border-gray-200 last:border-b-0">
                <div 
                  class="flex items-center hover:bg-gray-50 cursor-pointer"
                  :style="{ height: (expandedStaff.has(staff.ID) ? getRowHeight(staff.ID, 'staff') : 48) + 'px' }"
                  @click="toggleStaffExpand(staff.ID)"
                >
                  <div class="w-64 flex-shrink-0 px-4 py-3 border-r border-gray-200 flex items-center justify-between h-full">
                    <div class="flex items-center space-x-2 min-w-0 flex-1">
                      <i 
                        :class="[
                          'fas fa-chevron-right text-xs text-gray-400 transition-transform',
                          expandedStaff.has(staff.ID) ? 'rotate-90' : ''
                        ]"
                      ></i>
                      <StaffAvatar :employee="staff" size="sm" />
                      <div class="min-w-0 flex-1">
                        <h3 class="text-xs font-medium text-gray-900 truncate">
                          {{ staff.first_name }} {{ staff.last_name }}
                        </h3>
                        <p v-if="staff.title" class="text-xs text-gray-500 truncate">{{ staff.title }}</p>
                      </div>
                    </div>
                    <button
                      @click.stop="openAddAssignmentForStaff(staff)"
                      class="ml-2 p-1 text-gray-400 hover:text-sage rounded hover:bg-gray-100"
                      title="Add assignment"
                    >
                      <i class="fas fa-plus text-xs"></i>
                    </button>
                  </div>
                  
                  <!-- Timeline bars -->
                  <div class="flex-1 relative" :style="{ height: (expandedStaff.has(staff.ID) ? getRowHeight(staff.ID, 'staff') : 48) + 'px', minWidth: '960px' }">
                    <!-- Collapsed: Single combined bar -->
                    <div v-if="!expandedStaff.has(staff.ID) && getCombinedTimelineRange(staff.ID, 'staff')" 
                         :style="getCombinedBarStyle(staff.ID, 'staff')"
                         :class="getCombinedBarClass(staff.ID, 'staff')"
                         class="absolute h-8 top-2 rounded px-2 py-1"
                         :title="`${getStaffAssignments(staff.ID).length} assignment(s)`">
                      <div class="text-xs font-medium truncate">
                        {{ getStaffAssignments(staff.ID).length }} project{{ getStaffAssignments(staff.ID).length !== 1 ? 's' : '' }}
                      </div>
                    </div>
                    
                    <!-- Expanded: Individual stacked bars -->
                    <div v-else 
                         v-for="{ assignment, stackIndex } in getStackedAssignments(getStaffAssignments(staff.ID))" 
                         :key="assignment.ID"
                         :style="getTimelineBarStyle(assignment, stackIndex)"
                         :class="[
                           'absolute h-8 rounded px-2 py-1 cursor-pointer',
                           getTimelineBarClass(assignment)
                         ]"
                         :title="`${getProjectName(assignment.project_id)} - ${assignment.commitment || 0} hr/week`"
                         @click.stop="editAssignment(assignment)">
                      <div class="text-xs font-medium truncate">
                        {{ getProjectName(assignment.project_id) }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- By Project View (Gantt-style) -->
            <div v-if="viewMode === 'by-project'">
              <div v-for="project in filteredProjects" :key="project.ID" class="border-b border-gray-200 last:border-b-0">
                <div 
                  class="flex items-center hover:bg-gray-50 cursor-pointer"
                  :style="{ height: (expandedProjects.has(project.ID) ? getRowHeight(project.ID, 'project') : 48) + 'px' }"
                  @click="toggleProjectExpand(project.ID)"
                >
                  <div class="w-64 flex-shrink-0 px-4 py-3 border-r border-gray-200 flex items-center justify-between h-full">
                    <div class="flex items-center space-x-2 min-w-0 flex-1">
                      <i 
                        :class="[
                          'fas fa-chevron-right text-xs text-gray-400 transition-transform',
                          expandedProjects.has(project.ID) ? 'rotate-90' : ''
                        ]"
                      ></i>
                      <div class="min-w-0 flex-1">
                        <a
                          @click.stop="navigateToProject(project)"
                          class="text-xs font-medium text-sage hover:text-sage-dark truncate cursor-pointer flex items-center gap-1 group"
                        >
                          {{ project.name }}
                          <i class="fas fa-external-link-alt text-[8px] opacity-0 group-hover:opacity-100 transition-opacity"></i>
                        </a>
                        <a 
                          v-if="project.account?.name"
                          @click.stop="navigateToAccount(project)"
                          class="text-xs text-gray-500 hover:text-sage truncate cursor-pointer flex items-center gap-1 group"
                        >
                          {{ project.account.name }}
                          <i class="fas fa-external-link-alt text-[7px] opacity-0 group-hover:opacity-100 transition-opacity"></i>
                        </a>
                      </div>
                    </div>
                    <button
                      @click.stop="openAddAssignmentForProject(project)"
                      class="ml-2 p-1 text-gray-400 hover:text-sage rounded hover:bg-gray-100"
                      title="Add staff"
                    >
                      <i class="fas fa-plus text-xs"></i>
                    </button>
                  </div>
                  
                  <!-- Timeline bars -->
                  <div class="flex-1 relative" :style="{ height: (expandedProjects.has(project.ID) ? getRowHeight(project.ID, 'project') : 48) + 'px', minWidth: '960px' }">
                    <!-- Collapsed: Single combined bar -->
                    <div v-if="!expandedProjects.has(project.ID) && getCombinedTimelineRange(project.ID, 'project')" 
                         :style="getCombinedBarStyle(project.ID, 'project')"
                         :class="getCombinedBarClass(project.ID, 'project')"
                         class="absolute h-8 top-2 rounded px-2 py-1"
                         :title="`${getProjectAssignments(project.ID).length} staff member(s)`">
                      <div class="text-xs font-medium truncate">
                        {{ getProjectAssignments(project.ID).length }} staff member{{ getProjectAssignments(project.ID).length !== 1 ? 's' : '' }}
                      </div>
                    </div>
                    
                    <!-- Expanded: Individual stacked bars -->
                    <div v-else 
                         v-for="{ assignment, stackIndex } in getStackedAssignments(getProjectAssignments(project.ID))" 
                         :key="assignment.ID"
                         :style="getTimelineBarStyle(assignment, stackIndex)"
                         :class="[
                           'absolute h-8 rounded px-2 py-1 cursor-pointer',
                           getTimelineBarClass(assignment)
                         ]"
                         :title="`${assignment.employee?.first_name} ${assignment.employee?.last_name} - ${assignment.commitment || 0} hr/week`"
                         @click.stop="editAssignment(assignment)">
                      <div class="text-xs font-medium truncate">
                        {{ assignment.employee?.first_name }} {{ assignment.employee?.last_name }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Assignment Modal -->
    <AssignmentModal
      v-if="isAssignmentModalOpen"
      :project-id="selectedProject?.ID || 0"
      :is-open="isAssignmentModalOpen"
      :assignment-data="editingAssignment"
      :staff-list="staffMembers"
      :projects-list="projects"
      :project-data="selectedProject ? { active_start: selectedProject.active_start, active_end: selectedProject.active_end } : null"
      @close="closeAssignmentModal"
      @save="handleAssignmentSave"
      @delete="handleAssignmentDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { fetchProjects } from '../../api/projects';
import { getUsers } from '../../api/timesheet';
import { createProjectAssignment, updateProjectAssignment } from '../../api/projectAssignments';
import type { Project, Staff, StaffingAssignment } from '../../types/Project';
import StaffAvatar from '../../components/StaffAvatar.vue';
import AssignmentModal from '../../components/assignments/AssignmentModal.vue';

const router = useRouter();

// State
const isLoading = ref(true);
const error = ref<string | null>(null);
const viewMode = ref<'by-staff' | 'by-project'>('by-staff');
const projects = ref<Project[]>([]);
const staffMembers = ref<Staff[]>([]);
const assignments = ref<StaffingAssignment[]>([]);

// Modal state
const isAssignmentModalOpen = ref(false);
const editingAssignment = ref<StaffingAssignment | null>(null);
const selectedProject = ref<Project | null>(null);
const selectedStaff = ref<Staff | null>(null);

// Expansion state
const expandedStaff = ref<Set<number>>(new Set());
const expandedProjects = ref<Set<number>>(new Set());

// Filters
const statusFilter = ref<'all' | 'active' | 'upcoming' | 'ended'>('active');
const selectedYear = ref(new Date().getFullYear());

// Available years (current year +/- 2 years)
const currentYear = new Date().getFullYear();
const availableYears = [currentYear - 2, currentYear - 1, currentYear, currentYear + 1, currentYear + 2];

// Timeline periods (monthly view)
const timelinePeriods = computed(() => {
  return [
    { index: 0, label: 'Jan', start: 0, end: 1 },
    { index: 1, label: 'Feb', start: 1, end: 2 },
    { index: 2, label: 'Mar', start: 2, end: 3 },
    { index: 3, label: 'Apr', start: 3, end: 4 },
    { index: 4, label: 'May', start: 4, end: 5 },
    { index: 5, label: 'Jun', start: 5, end: 6 },
    { index: 6, label: 'Jul', start: 6, end: 7 },
    { index: 7, label: 'Aug', start: 7, end: 8 },
    { index: 8, label: 'Sep', start: 8, end: 9 },
    { index: 9, label: 'Oct', start: 9, end: 10 },
    { index: 10, label: 'Nov', start: 10, end: 11 },
    { index: 11, label: 'Dec', start: 11, end: 12 },
  ];
});

// Load all data
const loadData = async () => {
  isLoading.value = true;
  error.value = null;

  try {
    const [projectsData, staffData] = await Promise.all([
      fetchProjects(),
      getUsers()
    ]);

    projects.value = projectsData;
    staffMembers.value = staffData;

    // Extract all assignments from projects
    const allAssignments: StaffingAssignment[] = [];
    projectsData.forEach((project: Project) => {
      if (project.staffing_assignments) {
        project.staffing_assignments.forEach((assignment: StaffingAssignment) => {
          allAssignments.push(assignment);
        });
      }
    });
    assignments.value = allAssignments;

    // Initialize all staff and projects as expanded by default
    expandedStaff.value = new Set(staffData.map((s: Staff) => s.ID));
    expandedProjects.value = new Set(projectsData.map((p: Project) => p.ID));

  } catch (err) {
    console.error('Error loading staffing data:', err);
    error.value = 'Failed to load staffing data. Please try again.';
  } finally {
    isLoading.value = false;
  }
};

// Filtered staff and projects based on status
const filteredStaffMembers = computed(() => {
  if (statusFilter.value === 'all') return staffMembers.value;
  
  return staffMembers.value.filter(staff => {
    const staffAssignments = getStaffAssignments(staff.ID);
    if (staffAssignments.length === 0) return statusFilter.value === 'ended';
    
    return staffAssignments.some(assignment => {
      const status = getAssignmentStatus(assignment);
      return status.toLowerCase() === statusFilter.value;
    });
  });
});

const filteredProjects = computed(() => {
  if (statusFilter.value === 'all') return projects.value;
  
  return projects.value.filter(project => {
    const projectAssignments = getProjectAssignments(project.ID);
    if (projectAssignments.length === 0) return statusFilter.value === 'ended';
    
    return projectAssignments.some(assignment => {
      const status = getAssignmentStatus(assignment);
      return status.toLowerCase() === statusFilter.value;
    });
  });
});

// Get assignments for a specific staff member
const getStaffAssignments = (staffId: number) => {
  return assignments.value.filter(a => a.employee_id === staffId);
};

// Get assignments for a specific project
const getProjectAssignments = (projectId: number) => {
  return assignments.value.filter(a => a.project_id === projectId);
};

// Get project name by ID
const getProjectName = (projectId: number) => {
  const project = projects.value.find(p => p.ID === projectId);
  return project?.name || 'Unknown Project';
};


// Get assignment status
const getAssignmentStatus = (assignment: StaffingAssignment) => {
  if (!assignment.start_date || !assignment.end_date) return 'Unknown';
  
  const now = new Date();
  const startDate = new Date(assignment.start_date);
  const endDate = new Date(assignment.end_date);

  if (isNaN(startDate.getTime()) || isNaN(endDate.getTime())) return 'Unknown';
  if (now < startDate) return 'Upcoming';
  if (now > endDate) return 'Ended';
  return 'Active';
};


// Get combined timeline range (for collapsed view)
const getCombinedTimelineRange = (id: number, type: 'staff' | 'project'): { start: Date; end: Date } | null => {
  const assigns = type === 'staff' ? getStaffAssignments(id) : getProjectAssignments(id);
  if (assigns.length === 0) return null;
  
  let minStart: Date | null = null;
  let maxEnd: Date | null = null;
  
  assigns.forEach(assignment => {
    if (assignment.start_date && assignment.end_date) {
      const start = new Date(assignment.start_date);
      const end = new Date(assignment.end_date);
      
      if (!minStart || start < minStart) minStart = start;
      if (!maxEnd || end > maxEnd) maxEnd = end;
    }
  });
  
  return minStart && maxEnd ? { start: minStart, end: maxEnd } : null;
};

// Get combined bar style (for collapsed view - single bar spanning full range)
const getCombinedBarStyle = (id: number, type: 'staff' | 'project') => {
  const range = getCombinedTimelineRange(id, type);
  if (!range) return {};
  
  const yearStart = new Date(selectedYear.value, 0, 1);
  const yearEnd = new Date(selectedYear.value, 11, 31);
  
  const yearDuration = yearEnd.getTime() - yearStart.getTime();
  const startOffset = Math.max(0, range.start.getTime() - yearStart.getTime());
  const endOffset = Math.min(yearDuration, range.end.getTime() - yearStart.getTime());
  
  const left = (startOffset / yearDuration) * 100;
  const width = ((endOffset - startOffset) / yearDuration) * 100;
  
  return {
    left: `${left}%`,
    width: `${Math.max(width, 1)}%`
  };
};

// Get timeline bar style (position and width based on dates)
const getTimelineBarStyle = (assignment: StaffingAssignment, stackIndex: number = 0) => {
  if (!assignment.start_date || !assignment.end_date) return {};
  
  const yearStart = new Date(selectedYear.value, 0, 1);
  const yearEnd = new Date(selectedYear.value, 11, 31);
  const startDate = new Date(assignment.start_date);
  const endDate = new Date(assignment.end_date);
  
  // Calculate position as percentage of year
  const yearDuration = yearEnd.getTime() - yearStart.getTime();
  const startOffset = Math.max(0, startDate.getTime() - yearStart.getTime());
  const endOffset = Math.min(yearDuration, endDate.getTime() - yearStart.getTime());
  
  const left = (startOffset / yearDuration) * 100;
  const width = ((endOffset - startOffset) / yearDuration) * 100;
  
  // Stack bars vertically with 2px gap
  const barHeight = 32;
  const gap = 2;
  const top = stackIndex * (barHeight + gap);
  
  return {
    left: `${left}%`,
    width: `${Math.max(width, 1)}%`,
    top: `${top}px`
  };
};

// Get stacked assignments (separate overlapping assignments into rows)
const getStackedAssignments = (assigns: StaffingAssignment[]) => {
  if (assigns.length === 0) return [];
  
  // Sort by start date
  const sorted = [...assigns].sort((a, b) => {
    const aStart = new Date(a.start_date || 0);
    const bStart = new Date(b.start_date || 0);
    return aStart.getTime() - bStart.getTime();
  });
  
  const stacks: StaffingAssignment[][] = [];
  
  sorted.forEach(assignment => {
    const assignStart = new Date(assignment.start_date || 0);
    const assignEnd = new Date(assignment.end_date || 0);
    
    // Find a stack where this assignment doesn't overlap
    let placed = false;
    for (const stack of stacks) {
      const overlaps = stack.some(existing => {
        const existStart = new Date(existing.start_date || 0);
        const existEnd = new Date(existing.end_date || 0);
        return !(assignEnd < existStart || assignStart > existEnd);
      });
      
      if (!overlaps) {
        stack.push(assignment);
        placed = true;
        break;
      }
    }
    
    if (!placed) {
      stacks.push([assignment]);
    }
  });
  
  // Flatten with stack indices
  const result: Array<{ assignment: StaffingAssignment; stackIndex: number }> = [];
  stacks.forEach((stack, stackIndex) => {
    stack.forEach(assignment => {
      result.push({ assignment, stackIndex });
    });
  });
  
  return result;
};

// Calculate row height based on number of stacks needed
const getRowHeight = (id: number, type: 'staff' | 'project') => {
  const assigns = type === 'staff' ? getStaffAssignments(id) : getProjectAssignments(id);
  
  if (assigns.length > 0) {
    const stacked = getStackedAssignments(assigns);
    const maxStack = Math.max(...stacked.map(s => s.stackIndex), 0);
    const barHeight = 32;
    const gap = 2;
    return Math.max(48, (maxStack + 1) * (barHeight + gap) + 16);
  }
  
  return 48;
};

// Get timeline bar color class
const getTimelineBarClass = (assignment: StaffingAssignment) => {
  const status = getAssignmentStatus(assignment);
  if (status === 'Active') return 'bg-green-100 border border-green-300 text-green-800 hover:bg-green-200';
  if (status === 'Upcoming') return 'bg-blue-100 border border-blue-300 text-blue-800 hover:bg-blue-200';
  return 'bg-gray-100 border border-gray-300 text-gray-600 hover:bg-gray-200';
};

// Get combined bar color class based on overall status
const getCombinedBarClass = (id: number, type: 'staff' | 'project') => {
  const assigns = type === 'staff' ? getStaffAssignments(id) : getProjectAssignments(id);
  
  // Check if any assignment is active
  const hasActive = assigns.some(a => getAssignmentStatus(a) === 'Active');
  if (hasActive) return 'bg-green-100 border border-green-300 text-green-800';
  
  // Check if any assignment is upcoming
  const hasUpcoming = assigns.some(a => getAssignmentStatus(a) === 'Upcoming');
  if (hasUpcoming) return 'bg-blue-100 border border-blue-300 text-blue-800';
  
  // All ended
  return 'bg-gray-100 border border-gray-300 text-gray-600';
};

// Toggle expansion
const toggleStaffExpand = (staffId: number) => {
  if (expandedStaff.value.has(staffId)) {
    expandedStaff.value.delete(staffId);
  } else {
    expandedStaff.value.add(staffId);
  }
};

const toggleProjectExpand = (projectId: number) => {
  if (expandedProjects.value.has(projectId)) {
    expandedProjects.value.delete(projectId);
  } else {
    expandedProjects.value.add(projectId);
  }
};

// Update timeline when year changes
const updateTimeline = () => {
  // Timeline periods is computed, so it will update automatically
};

// Open modal to add assignment for staff (pre-populate employee)
const openAddAssignmentForStaff = (staff: Staff) => {
  selectedStaff.value = staff;
  // Create partial assignment with just the employee_id pre-filled
  editingAssignment.value = {
    employee_id: staff.ID,
    employee: staff,
    start_date: '',
    end_date: ''
  } as any;
  // Default to first active project
  selectedProject.value = projects.value.find(p => {
    const now = new Date();
    return new Date(p.active_start) <= now && new Date(p.active_end) >= now;
  }) || projects.value[0] || { ID: 0, active_start: '', active_end: '', name: '', account: null, staffing_assignments: [] };
  isAssignmentModalOpen.value = true;
};

// Open modal to add assignment for project (pre-populate project)
const openAddAssignmentForProject = (project: Project) => {
  selectedProject.value = project;
  selectedStaff.value = null;
  // Create partial assignment with just the project_id pre-filled
  editingAssignment.value = {
    project_id: project.ID,
    start_date: '',
    end_date: ''
  } as any;
  isAssignmentModalOpen.value = true;
};

// Edit existing assignment
const editAssignment = (assignment: StaffingAssignment) => {
  editingAssignment.value = JSON.parse(JSON.stringify(assignment));
  const project = projects.value.find(p => p.ID === assignment.project_id);
  selectedProject.value = project || null;
  isAssignmentModalOpen.value = true;
};


// Close modal
const closeAssignmentModal = () => {
  isAssignmentModalOpen.value = false;
  editingAssignment.value = null;
  selectedProject.value = null;
  selectedStaff.value = null;
};

// Handle assignment save
const handleAssignmentSave = async (assignmentData: any) => {
  try {
    if (editingAssignment.value && editingAssignment.value.ID) {
      // Update existing assignment
      await updateProjectAssignment(editingAssignment.value.ID, assignmentData);
    } else {
      // Create new assignment - use project_id from assignmentData
      const projectId = assignmentData.project_id || selectedProject.value?.ID;
      if (!projectId) {
        alert('Please select a project');
        return;
      }
      await createProjectAssignment(projectId, assignmentData);
    }
    
    // Reload data to reflect changes
    await loadData();
    closeAssignmentModal();
  } catch (err) {
    console.error('Error saving assignment:', err);
    alert('Failed to save assignment. Please try again.');
  }
};

// Handle assignment delete
const handleAssignmentDelete = async (assignmentId: number) => {
  if (!confirm('Are you sure you want to delete this assignment?')) {
    return;
  }
  
  try {
    const { deleteProjectAssignment } = await import('../../api/projectAssignments');
    await deleteProjectAssignment(assignmentId);
    await loadData();
    closeAssignmentModal();
  } catch (err) {
    console.error('Error deleting assignment:', err);
    alert('Failed to delete assignment. Please try again.');
  }
};

// Navigate to account in organization view
const navigateToAccount = (project: Project) => {
  if (!project.account_id) {
    router.push('/organization');
    return;
  }
  
  router.push({
    path: '/organization',
    query: {
      accountId: project.account_id.toString()
    }
  });
};

// Navigate to project in organization view
const navigateToProject = (project: Project) => {
  if (!project.account_id) {
    // If no account, just go to organization page
    router.push('/organization');
    return;
  }
  
  // Navigate to organization with account and project IDs in query params
  router.push({
    path: '/organization',
    query: {
      accountId: project.account_id.toString(),
      projectId: project.ID.toString()
    }
  });
};

// Load data on mount
onMounted(() => {
  loadData();
});
</script>
