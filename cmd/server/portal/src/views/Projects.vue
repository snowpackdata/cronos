<template>
  <div class="px-4 sm:px-6 lg:px-8 py-4">
    <div class="sm:flex sm:items-center mb-4">
      <div class="sm:flex-auto">
        <h1 class="text-base font-semibold leading-6 text-gray-900">Your Projects</h1>
        <p class="mt-1 text-sm text-gray-700">
          A list of all projects associated with your account.
        </p>
      </div>
      <!-- No action buttons in the portal view -->
    </div>

    <div class="mt-6 flow-root">
      <div v-if="isLoading" class="text-center py-8">
        <p class="text-gray-500 text-sm">Loading projects...</p>
        <!-- Optional: Add a spinner icon here -->
      </div>
      <div v-else-if="projects.length === 0 && !apiError" class="text-center py-8">
         <div class="flex flex-col items-center justify-center p-6 bg-white rounded-lg shadow-sm">
            <svg class="mx-auto h-10 w-10 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 12.75V12A2.25 2.25 0 014.5 9.75h15A2.25 2.25 0 0121.75 12v.75m-8.69-6.44l-2.12-2.12a1.5 1.5 0 00-1.061-.44H4.5A2.25 2.25 0 002.25 6v12a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9a2.25 2.25 0 00-2.25-2.25h-5.379a1.5 1.5 0 01-1.06-.44z" />
            </svg>
            <h3 class="mt-2 text-sm font-medium text-gray-900">No projects</h3>
            <p class="mt-1 text-xs text-gray-500">You currently have no projects assigned to your account.</p>
          </div>
      </div>
      <div v-else-if="apiError" class="mt-6 rounded-md bg-red-50 p-3">
        <div class="flex">
            <div class="flex-shrink-0">
                <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd" />
                </svg>
            </div>
            <div class="ml-2">
                <h3 class="text-sm font-medium text-red-800">Error loading projects</h3>
                <p class="mt-1 text-xs text-red-700">{{ apiError }}</p>
            </div>
        </div>
      </div>
      <ul v-else role="list" class="space-y-3">
        <li v-for="project in sortedProjects" :key="project.ID" class="bg-white shadow-sm overflow-hidden rounded-md border border-gray-200">
          <div class="px-3 py-2.5 sm:px-4">
            <div class="flex items-center justify-between gap-x-2 mb-1.5">
              <h3 class="text-sm font-semibold leading-snug text-gray-800">{{ project.name }}</h3>
              <span :class="getProjectDisplayStatus(project.active_start, project.active_end).class" 
                    class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium whitespace-nowrap">
                {{ getProjectDisplayStatus(project.active_start, project.active_end).text }}
              </span>
            </div>
            
            <p v-if="project.description" class="text-xs text-gray-500 mb-2 min-h-[30px]">{{ project.description }}</p>
            <p v-else class="text-xs text-gray-400 italic mb-2 min-h-[30px]">No description provided.</p>

            <!-- Main 2-Column Grid for Details, Budget, Team, Billing Codes -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-x-6 gap-y-4 border-t border-gray-200 pt-3 mt-3 text-xs">
              <!-- Col 1: Project Details -->
              <div>
                <h4 class="font-medium text-gray-600 mb-1.5">Details:</h4>
                <dl class="space-y-1">
                  <div class="flex items-baseline">
                    <dt class="text-gray-500 w-16 flex-shrink-0">Billing:</dt>
                    <dd class="font-medium text-gray-700 truncate">{{ formatBillingFrequency(project.billing_frequency) || 'N/A' }}</dd>
                  </div>
                  <div class="flex items-baseline">
                    <dt class="text-gray-500 w-16 flex-shrink-0">Start:</dt>
                    <dd class="font-medium text-gray-700 truncate">{{ formatDate(project.active_start) || 'N/A' }}</dd>
                  </div>
                  <div class="flex items-baseline">
                    <dt class="text-gray-500 w-16 flex-shrink-0">End:</dt>
                    <dd class="font-medium text-gray-700 truncate">{{ formatDate(project.active_end) || 'N/A' }}</dd>
                  </div>
                </dl>
              </div>

              <!-- Col 2: Budget Details -->
              <div>
                <h4 class="font-medium text-gray-600 mb-1.5">Budget:</h4>
                <div v-if="project.budget_hours || project.budget_dollars || project.budget_cap_hours || project.budget_cap_dollars" class="space-y-1">
                  <div v-if="project.budget_hours" class="flex justify-between items-center">
                    <span class="text-gray-500">Periodic Hours:</span>
                    <span class="font-medium text-gray-800 bg-indigo-50 px-1.5 py-0.5 rounded-md ring-1 ring-inset ring-indigo-200">
                      {{ project.budget_hours }} hrs{{ formatBudgetFrequencySuffix(project.billing_frequency) }}
                    </span>
                  </div>
                  <div v-if="project.budget_dollars" class="flex justify-between items-center">
                    <span class="text-gray-500">Periodic $:</span>
                    <span class="font-medium text-gray-800 bg-sage-50 px-1.5 py-0.5 rounded-md ring-1 ring-inset ring-sage-200">
                      {{ formatCurrency(project.budget_dollars) }}{{ formatBudgetFrequencySuffix(project.billing_frequency) }}
                    </span>
                  </div>
                  <div v-if="project.budget_cap_hours" class="flex justify-between items-center">
                    <span class="text-gray-500">Total Hours Cap:</span>
                    <span class="font-medium text-gray-800 bg-sky-50 px-1.5 py-0.5 rounded-md ring-1 ring-inset ring-sky-200">
                      {{ project.budget_cap_hours }} hrs (Cap)
                    </span>
                  </div>
                  <div v-if="project.budget_cap_dollars" class="flex justify-between items-center">
                    <span class="text-gray-500">Total $ Cap:</span>
                    <span class="font-medium text-gray-800 bg-rose-50 px-1.5 py-0.5 rounded-md ring-1 ring-inset ring-rose-200">
                      {{ formatCurrency(project.budget_cap_dollars) }} (Cap)
                    </span>
                  </div>
                </div>
                <p v-else class="text-xs text-gray-500 italic bg-gray-50 p-2 rounded-md border border-gray-100 text-center">No budget information provided.</p>
              </div>

              <!-- Divider between rows -->
              <div class="md:col-span-2 py-2">
                <div class="border-t border-gray-200"></div>
              </div>

              <!-- Col 1 (Row 2): Assigned Team -->
              <div>
                <h4 class="text-xs font-medium text-gray-600 mb-1">Assigned Team:</h4>
                <div v-if="project.staffing_assignments && project.staffing_assignments.length > 0"
                     class="bg-white shadow-sm border border-gray-200 rounded-md p-2">
                  <ul role="list" class="divide-y divide-gray-100 text-xs">
                    <li v-for="assignment in project.staffing_assignments" :key="assignment.ID" class="py-1.5">
                      <div class="flex items-center justify-between gap-x-2">
                        <div class="flex items-center gap-x-2 flex-grow min-w-0">
                          <span 
                            class="inline-flex items-center rounded-md px-1.5 py-0.5 text-xs font-medium ring-1 ring-inset whitespace-nowrap flex-shrink-0"
                            :class="getStaffingAssignmentStatus(assignment).color">
                            {{ getStaffingAssignmentStatus(assignment).label }}
                          </span>
                          <div class="min-w-0">
                            <p class="font-medium text-gray-700 truncate">
                              {{ assignment.employee?.first_name }} {{ assignment.employee?.last_name }}
                              <span v-if="assignment.employee?.title" class="text-gray-500 font-normal"> ({{ assignment.employee?.title }})</span>
                            </p>
                            <p v-if="assignment.role" class="text-xs text-gray-500">Role: {{ assignment.role }}</p>
                          </div>
                        </div>
                      </div>
                      <div class="pl-8 text-xs text-gray-500 mt-0.5">
                        <span class="mr-2">
                          <i class="far fa-clock mr-0.5 text-gray-400"></i> {{ getCommitmentDisplay(assignment) }}
                        </span>
                        <span v-if="assignment.start_date || assignment.end_date">
                          <i class="far fa-calendar-alt mr-0.5 text-gray-400"></i>
                          {{ formatDate(assignment.start_date, true) }} - {{ formatDate(assignment.end_date, true) }}
                        </span>
                      </div>
                    </li>
                  </ul>
                </div>
                <p v-else class="text-xs text-gray-500 italic bg-gray-50 p-2 rounded-md border border-gray-100 text-center">No specific team assignments listed for this project.</p>
              </div>

              <!-- Col 2 (Row 2): Billing Codes -->
              <div>
                <h4 class="text-xs font-medium text-gray-600 mb-1">Billing Codes:</h4>
                <div v-if="project.billing_codes && project.billing_codes.length > 0"
                     class="bg-white shadow-sm border border-gray-200 rounded-md p-2">
                  <ul role="list" class="divide-y divide-gray-100 text-xs">
                    <li v-for="bc in project.billing_codes" :key="bc.ID" class="flex items-center justify-between py-1">
                      <span class="text-gray-600 truncate">{{ bc.name }} ({{ bc.code }})</span>
                      <span class="ml-2 text-xs font-medium text-gray-700 whitespace-nowrap inline-flex items-center rounded-md px-1.5 py-0.5 bg-blue-100 text-blue-700 ring-1 ring-inset ring-blue-200">
                        {{ formatRate(bc.rate, bc.rate_type) }}
                      </span>
                    </li>
                  </ul>
                </div>
                <p v-else class="text-xs text-gray-500 italic bg-gray-50 p-2 rounded-md border border-gray-100 text-center">No specific billing codes for this project.</p>
              </div>
            </div>
            
            <!-- Project Assets Section (Below the grid, full width) -->
            <div class="mt-3 pt-3 border-t border-gray-200">
              <h4 class="text-xs font-semibold text-gray-500 mb-1.5">Project Assets:</h4>
              <div v-if="project.assets && project.assets.length > 0" class="space-y-1">
                <AssetDisplayItem 
                  v-for="asset in project.assets"
                  :key="asset.ID" 
                  :asset="asset"
                  :project-id="project.ID" 
                  :is-read-only="true"
                  @asset-updated="handlePortalAssetUpdated({ projectId: project.ID, asset: $event })"
                />
              </div>
              <p v-else class="text-xs text-gray-500 italic bg-gray-50 p-2 rounded-md border border-gray-100 text-center">
                No assets available for this project.
              </p>
            </div>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { portalAPI } from '../api';
// Adjusted paths to use local portal components and types
import AssetDisplayItem from '../components/assets/AssetDisplayItem.vue'; 
import type { Asset } from '../types/Asset';

interface PortalRate {
  ID: number;
  name: string;
  amount: number;
  // Add other relevant rate fields, e.g., unit (per hour, fixed)
}

interface PortalBillingCode {
  ID: number;
  name: string;
  code: string;
  rate_type: string; // e.g., HOURLY, FIXED_FEE - depends on your cronos.BillingCodeRateType constants
  rate?: PortalRate;
  // Add other fields from cronos.BillingCode if needed
}

// Define PortalStaff interface (similar to AdminStaff)
interface PortalStaff {
  ID: number;
  first_name?: string;
  last_name?: string;
  title?: string; // Or role, depending on what your API provides for portal view
}

// Define PortalStaffingAssignment interface
interface PortalStaffingAssignment {
  ID: number;
  role?: string; // Role specific to this assignment for this project
  commitment?: number; // Legacy weekly commitment hours
  commitment_schedule?: string; // JSON string of segment schedule
  start_date?: string | Date;
  end_date?: string | Date;
  employee?: PortalStaff; // Nested staff details
}

interface PortalProject {
  ID: number;
  name: string;
  description?: string;
  status?: string;
  project_type?: string;
  active_start?: string | Date;
  active_end?: string | Date;
  budget_hours?: number;
  budget_dollars?: number;
  billing_frequency?: string;
  billing_codes?: PortalBillingCode[];
  staffing_assignments?: PortalStaffingAssignment[];
  assets?: Asset[]; // This now uses the local portal Asset type
  budget_cap_dollars?: number;
  budget_cap_hours?: number;
}

const projects = ref<PortalProject[]>([]);
const isLoading = ref(true);
const apiError = ref<string | null>(null);

const handlePortalAssetUpdated = (updatedAssetInfo: { projectId: number, asset: Asset }) => {
  const projectIndex = projects.value.findIndex(p => p.ID === updatedAssetInfo.projectId);
  if (projectIndex !== -1) {
    const project = projects.value[projectIndex];
    if (project.assets) {
      const assetIndex = project.assets.findIndex(a => a.ID === updatedAssetInfo.asset.ID);
      if (assetIndex !== -1) {
        const newAssets = [...project.assets];
        newAssets[assetIndex] = { ...newAssets[assetIndex], ...updatedAssetInfo.asset };
        projects.value[projectIndex] = { ...project, assets: newAssets };
      }
    }
  }
};

const sortedProjects = computed(() => {
  return [...projects.value].sort((a: PortalProject, b: PortalProject) => { 
    const dateA = a.active_end ? new Date(a.active_end).getTime() : 0;
    const dateB = b.active_end ? new Date(b.active_end).getTime() : 0;
    if (isNaN(dateA) && isNaN(dateB)) return 0;
    if (isNaN(dateA)) return 1;
    if (isNaN(dateB)) return -1;
    return dateB - dateA;
  });
});

// Function to determine project status based on dates
const getProjectDisplayStatus = (startDateString?: string | Date, endDateString?: string | Date): { text: string; class: string } => {
  const today = new Date();
  today.setHours(0, 0, 0, 0); // Normalize today to start of day for accurate comparison

  let startDate: Date | null = null;
  if (startDateString) {
    try {
      startDate = new Date(startDateString);
      if (isNaN(startDate.getTime())) startDate = null; // Invalid date string
      else startDate.setHours(0, 0, 0, 0);
    } catch { startDate = null; }
  }

  let endDate: Date | null = null;
  if (endDateString) {
    try {
      endDate = new Date(endDateString);
      if (isNaN(endDate.getTime())) endDate = null; // Invalid date string
      else endDate.setHours(0,0,0,0);
    } catch { endDate = null; }
  }

  if (startDate && startDate > today) {
    return { text: 'Upcoming', class: 'bg-blue-100 text-blue-700' };
  }
  
  if (endDate && endDate < today) {
    return { text: 'Completed', class: 'bg-gray-200 text-gray-700' };
  }
  
  // If it's not upcoming and not completed, it's either active or ongoing.
  // Active: Has a start date that is today or in the past, AND (no end date OR end date is today or in the future)
  if (startDate && startDate <= today && (!endDate || endDate >= today) ) {
     return { text: 'Active', class: 'bg-green-200 text-green-800' };
  }
  
  // Fallback for projects with unclear status based on dates (e.g. only end date in future, no start)
  return { text: 'Ongoing', class: 'bg-yellow-100 text-yellow-700' }; 
};

onMounted(async () => {
  isLoading.value = true;
  apiError.value = null;
  try {
    const data = await portalAPI.fetchPortalProjects();
    projects.value = (data || []).map((p: PortalProject) => ({
      ...p,
      staffing_assignments: (p.staffing_assignments || []).map(sa => ({
        ...sa,
        employee: sa.employee || undefined,
        commitment: sa.commitment !== undefined ? Number(sa.commitment) : undefined,
        start_date: sa.start_date || undefined,
        end_date: sa.end_date || undefined,
      })),
    }));
  } catch (error) {
    console.error('Error fetching portal projects:', error);
    if (error instanceof Error) {
      apiError.value = error.message;
    } else {
      apiError.value = 'An unknown error occurred.';
    }
    projects.value = []; // Clear projects on error
  } finally {
    isLoading.value = false;
  }
});

// Helper function to format dates (example)
const formatDate = (dateString?: string | Date, short: boolean = false) => {
  if (!dateString) return 'N/A';
  if (typeof dateString === 'string' && dateString.startsWith('0001-01-01')) return 'N/A';
  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return 'N/A';
    return date.toLocaleDateString(undefined, { year: 'numeric', month: short ? 'short' : 'long', day: 'numeric' });
  } catch (e) {
    return 'N/A';
  }
};


// Helper function to format currency (example)
const formatCurrency = (amount?: number) => {
  if (amount === undefined || amount === null) return '';
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(amount);
};

// Helper to display commitment (handles segments)
const getCommitmentDisplay = (assignment: PortalStaffingAssignment): string => {
  // Try to parse segments
  if (assignment.commitment_schedule) {
    try {
      const parsed = JSON.parse(assignment.commitment_schedule);
      const segments = parsed.segments || [];
      
      if (segments.length > 0) {
        // Check if all segments have the same commitment
        const firstCommitment = segments[0].commitment;
        const allSame = segments.every((seg: any) => seg.commitment === firstCommitment);
        
        if (allSame) {
          return `${firstCommitment} hr/wk`;
        } else {
          // Variable schedule - show range
          const commitments = segments.map((seg: any) => seg.commitment);
          const min = Math.min(...commitments);
          const max = Math.max(...commitments);
          return `${min}-${max} hr/wk (variable)`;
        }
      }
    } catch (e) {
      // Fall through to legacy commitment
    }
  }
  
  // Fallback to legacy commitment field
  if (assignment.commitment !== undefined) {
    return `${assignment.commitment} hr/wk`;
  }
  
  return 'Not specified';
};

// Helper function to format billing frequency
const formatBillingFrequency = (frequency?: string) => {
  if (!frequency) return 'N/A';
  const frequencies: { [key: string]: string } = {
    'BILLING_TYPE_WEEKLY': 'Weekly',
    'BILLING_TYPE_BIWEEKLY': 'Bi-Weekly',
    'BILLING_TYPE_MONTHLY': 'Monthly',
    'BILLING_TYPE_BIMONTHLY': 'Bi-Monthly',
    'BILLING_TYPE_PROJECT': 'Project-Based'
  };
  return frequencies[frequency] || frequency.replace('BILLING_TYPE_', '').replace(/_/g, ' ');
};

// Helper function to format budget frequency suffix
const formatBudgetFrequencySuffix = (billingFrequency?: string): string => {
  if (!billingFrequency) return '';
  const formattedFrequency = formatBillingFrequency(billingFrequency);
  if (formattedFrequency && formattedFrequency !== 'N/A' && formattedFrequency !== 'Project-Based') {
    // Make it lowercase and prepend " per "
    return ` per ${formattedFrequency.toLowerCase().replace('-based', '').trim()}`;
  }
  return ''; // No suffix for project-based, N/A, or unmapped
};

const formatRate = (rate?: PortalRate, rateType?: string) => {
  if (!rate) return 'N/A';
  let suffix = '';
  // Assuming rate_type might be like 'HOURLY', 'FIXED_FEE' from your cronos.BillingCodeRateType constants
  // Adjust these conditions based on your actual cronos.BillingCodeRateType values
  if (rateType) {
    if (rateType.toUpperCase().includes('HOURLY')) {
      suffix = '/hr';
    } else if (rateType.toUpperCase().includes('FIXED')) {
      suffix = ' fixed';
    }
    // Add more conditions for other rate types if necessary (e.g., per unit, daily)
  }
  // Format rate with 2 decimal places
  const formattedAmount = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(rate.amount);
  return `${formattedAmount}${suffix}`;
};

interface AssignmentStatus {
  label: string;
  color: string;
}

const getStaffingAssignmentStatus = (assignment: PortalStaffingAssignment): AssignmentStatus => {
  const now = new Date();
  now.setHours(0, 0, 0, 0);

  const startDateString = assignment.start_date;
  const endDateString = assignment.end_date;

  if (!startDateString) {
    if (!endDateString) return { label: 'Active', color: 'bg-green-100 text-green-700 ring-green-600/20' };
    const endDate = new Date(endDateString);
    endDate.setHours(0,0,0,0);
    return now <= endDate ? { label: 'Active', color: 'bg-green-100 text-green-700 ring-green-600/20' } : { label: 'Ended', color: 'bg-gray-100 text-gray-600 ring-gray-500/20' };
  }
  const startDate = new Date(startDateString);
  startDate.setHours(0,0,0,0);
  if (now < startDate) {
    return { label: 'Upcoming', color: 'bg-blue-100 text-blue-700 ring-blue-600/20' };
  }
  if (!endDateString) {
    return { label: 'Active', color: 'bg-green-100 text-green-700 ring-green-600/20' };
  }
  const endDate = new Date(endDateString);
  endDate.setHours(0,0,0,0);
  if (now > endDate) {
    return { label: 'Ended', color: 'bg-gray-100 text-gray-600 ring-gray-500/20' };
  }
  return { label: 'Active', color: 'bg-green-100 text-green-700 ring-green-600/20' };
};

</script>

<style scoped>
/* Add any component-specific styles here if needed */
.min-h-\[30px\] {
  min-height: 30px; /* Adjusted for project description */
}
</style> 