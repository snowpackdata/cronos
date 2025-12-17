<template>
  <div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow transition hover:shadow-md">
    <!-- Project Header - Clickable to expand/collapse -->
    <div 
      class="flex items-start justify-between p-4 cursor-pointer hover:bg-gray-50 transition-colors"
      @click="toggleExpanded"
    >
      <div class="flex-1">
        <div class="flex items-center gap-x-2">
          <!-- Expand/Collapse Chevron -->
          <i 
            class="fas transition-transform text-gray-400 text-xs"
            :class="isExpanded ? 'fa-chevron-down' : 'fa-chevron-right'"
          ></i>
          
          <h3 class="text-base font-semibold text-gray-900">{{ project.name }}</h3>
          
          <!-- Project Status Badge -->
          <span 
            class="inline-flex items-center rounded-md px-1.5 py-0.5 text-xs font-medium ring-1 ring-inset"
            :class="getProjectStatus(project).color">
            {{ getProjectStatus(project).label }}
          </span>
          
          <!-- Internal Badge if applicable -->
          <span 
            v-if="project.internal"
            class="inline-flex items-center rounded-md bg-purple-50 px-1.5 py-0.5 text-xs font-medium text-purple-700 ring-1 ring-inset ring-purple-700/10">
            Internal
          </span>
        </div>
        <p class="mt-0.5 text-xs text-gray-500 ml-5">{{ project.account ? project.account.name : 'No Account' }}</p>
        <!-- Display Project Description when collapsed -->
        <p v-if="!isExpanded && project.description" class="mt-1 text-xs text-gray-600 truncate ml-5" :title="project.description">
          {{ project.description }}
        </p>
        <!-- Display Staffing Assignments Count when collapsed -->
        <p v-if="!isExpanded && project.staffing_assignments && project.staffing_assignments.length > 0" class="mt-1 text-xs text-gray-500 ml-5">
          {{ project.staffing_assignments.length }} staff assigned
        </p>
      </div>
      <div class="flex items-center space-x-2" @click.stop>
        <button
          @click="$emit('edit', project)"
          class="inline-flex items-center rounded-md bg-blue-50 px-2 py-1 text-xs font-medium text-blue-700 ring-1 ring-inset ring-blue-700/10 hover:bg-blue-100 transition-colors"
          title="Edit Project"
        >
          <i class="fas fa-pencil-alt mr-1"></i>
          Edit
        </button>
      </div>
    </div>
    
    <!-- Expandable Card Body -->
    <div v-if="isExpanded" class="border-t border-gray-100 p-4 pt-3">
      <!-- Three-Column Layout for Card Body -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-x-4 gap-y-3">
      <!-- Column 1: Project Details (Budget, Timeframe, Billing Codes) -->
      <div class="space-y-3">
        <!-- Budget Information -->
        <div v-if="project.budget_dollars || project.budget_hours || project.budget_cap_dollars || project.budget_cap_hours" class="mb-3">
          <div class="mb-1 text-xs font-medium text-gray-500">Budget:</div>
          <div class="flex flex-wrap gap-1.5">
            <div v-if="project.budget_dollars" class="flex items-center">
              <span class="inline-flex items-center rounded-md bg-sage-50 px-1.5 py-0.5 text-xs font-medium text-sage-700 ring-1 ring-inset ring-sage-600/20">
                <i class="fas fa-dollar-sign mr-1"></i>
                {{ formatCurrency(project.budget_dollars) }}{{ formatBudgetFrequency(project.billing_frequency) }}
              </span>
            </div>
            <div v-if="project.budget_hours" class="flex items-center">
              <span class="inline-flex items-center rounded-md bg-indigo-50 px-1.5 py-0.5 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-600/20">
                <i class="fas fa-clock mr-1"></i>
                {{ project.budget_hours }} hours{{ formatBudgetFrequency(project.billing_frequency) }}
              </span>
            </div>
            <div v-if="project.budget_cap_dollars" class="flex items-center">
              <span class="inline-flex items-center rounded-md bg-rose-50 px-1.5 py-0.5 text-xs font-medium text-rose-700 ring-1 ring-inset ring-rose-600/20">
                <i class="fas fa-shield-alt mr-1"></i>
                {{ formatCurrency(project.budget_cap_dollars) }} (Cap)
              </span>
            </div>
            <div v-if="project.budget_cap_hours" class="flex items-center">
              <span class="inline-flex items-center rounded-md bg-sky-50 px-1.5 py-0.5 text-xs font-medium text-sky-700 ring-1 ring-inset ring-sky-600/20">
                <i class="fas fa-stopwatch-20 mr-1"></i>
                {{ project.budget_cap_hours }} hours (Cap)
              </span>
            </div>
          </div>
        </div>
        
        <!-- Timeframe Information -->
        <dl class="grid grid-cols-1 gap-y-2 text-xs">
          <div class="flex items-center">
            <dt class="font-medium text-gray-500 w-20">Timeframe:</dt>
            <dd class="text-gray-900">
              {{ formatDate(project.active_start) }} to {{ formatDate(project.active_end) }}
            </dd>
          </div>
          <div class="flex items-center">
            <dt class="font-medium text-gray-500 w-20">Status:</dt>
            <dd class="text-gray-900">{{ getProjectTimeframe(project) }}</dd>
          </div>
          <div class="flex items-center">
            <dt class="font-medium text-gray-500 w-20">Type:</dt>
            <dd class="text-gray-900">{{ formatProjectType(project.project_type) }}</dd>
          </div>
          <div class="flex items-center">
            <dt class="font-medium text-gray-500 w-20">Billing:</dt>
            <dd class="text-gray-900">{{ formatBillingFrequency(project.billing_frequency) }}</dd>
          </div>
        </dl>

        <!-- Billing Codes Section -->
        <div>
          <h5 class="text-xs font-medium text-gray-900 mb-2">Billing Codes</h5>
          <div v-if="project.billing_codes && project.billing_codes.length > 0" class="space-y-0 divide-y divide-gray-100 border border-gray-200 rounded-lg overflow-hidden">
            <div v-for="code in project.billing_codes" :key="code.ID" 
                  class="flex items-center justify-between gap-x-4 py-2 px-3 hover:bg-gray-50">
              <div class="min-w-0 flex-1">
                <div class="flex items-start gap-x-2">
                  <p class="text-xs font-semibold text-gray-900">{{ code.name }}</p>
                  <p v-if="isBillingCodeActive(code)" class="text-green-700 bg-green-50 ring-green-600/20 mt-0.5 whitespace-nowrap rounded-md px-1 py-0.5 text-xs font-medium ring-1 ring-inset">
                    Active
                  </p>
                  <p v-else class="text-red-700 bg-red-50 ring-red-600/20 mt-0.5 whitespace-nowrap rounded-md px-1 py-0.5 text-xs font-medium ring-1 ring-inset">
                    Inactive
                  </p>
                </div>
                <div class="mt-0.5 flex items-center gap-x-2 text-xs text-gray-500">
                  <p class="whitespace-nowrap">
                    <span class="font-medium">{{ code.code }}</span>
                  </p>
                  <svg viewBox="0 0 2 2" class="h-0.5 w-0.5 fill-current">
                    <circle cx="1" cy="1" r="1" />
                  </svg>
                  <!-- Rates Section -->
                  <p v-if="isLoadingRates" class="whitespace-nowrap italic">
                    Loading rates...
                  </p>
                  <div v-else class="flex flex-row items-center text-xs">
                    <!-- Client Rate -->
                    <div v-if="getRateById(code.rate_id)">
                        <p class="whitespace-nowrap">
                            <span class="font-medium text-green-700">{{ formatCurrency(getRateById(code.rate_id)?.amount || 0) }}</span>
                        </p>
                    </div>

                    <!-- Separator 1 -->
                    <span v-if="getRateById(code.rate_id) && getRateById(code.internal_rate_id)" class="text-gray-400 mx-1">/</span>

                    <!-- Internal Rate -->
                    <div v-if="getRateById(code.internal_rate_id)">
                        <p class="whitespace-nowrap">
                            <span class="font-medium text-blue-700">{{ formatCurrency(getRateById(code.internal_rate_id)?.amount || 0) }}</span>
                        </p>
                    </div>

                    <!-- Separator 2: Show if margin will be shown (i.e., if both client and internal rates exist) -->
                    <span v-if="getRateById(code.rate_id) && getRateById(code.internal_rate_id)" class="text-gray-400 mx-1">/</span>

                    <!-- Margin -->
                    <div v-if="getRateById(code.rate_id) && getRateById(code.internal_rate_id)">
                        <p class="whitespace-nowrap">
                            <span class="font-medium text-indigo-600">
                                {{ formatProfit(getRateById(code.rate_id)?.amount || 0, getRateById(code.internal_rate_id)?.amount || 0) }}
                            </span>
                        </p>
                    </div>
                    
                    <!-- Fallback if NO rates at all -->
                    <p v-if="!getRateById(code.rate_id) && !getRateById(code.internal_rate_id)" class="whitespace-nowrap italic text-gray-500">
                        No rate data
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="text-xs text-gray-700 italic bg-gray-50 p-3 rounded-lg border border-gray-200 text-center">
            <p>No billing codes available for this project.</p>
          </div>
        </div>
      </div>
      
      <!-- Column 2: Staffing Information -->
      <div class="space-y-3">
        <!-- Account Team Information -->
        <div>
          <h5 class="text-xs font-medium text-gray-900 mb-2">Account Team</h5>
          <div class="space-y-0 divide-y divide-gray-100 border border-gray-200 rounded-lg overflow-hidden">
            <div class="flex items-center justify-between py-2 px-3 hover:bg-gray-50">
              <div class="min-w-0 flex-1">
                <div class="flex items-start">
                  <p class="text-xs font-semibold text-gray-900">Account Executive</p>
                </div>
                <div class="mt-0.5 text-xs text-gray-500">
                  <p>{{ getStaffName(project.ae_id) }}</p>
                </div>
              </div>
            </div>
            <div class="flex items-center justify-between py-2 px-3 hover:bg-gray-50">
              <div class="min-w-0 flex-1">
                <div class="flex items-start">
                  <p class="text-xs font-semibold text-gray-900">Sales Development Rep</p>
                </div>
                <div class="mt-0.5 text-xs text-gray-500">
                  <p>{{ getStaffName(project.sdr_id) }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Assigned Staff Section -->
        <div>
          <div class="flex justify-between items-center mb-2">
            <h5 class="text-xs font-medium text-gray-900">Assigned Staff</h5>
            <button
              @click="openAddAssignmentModal"
              :disabled="isLoadingStaff || staffList.length === 0"
              class="inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset transition-colors"
              :class="isLoadingStaff || staffList.length === 0 
                ? 'bg-gray-50 text-gray-400 ring-gray-300/50 cursor-not-allowed' 
                : 'bg-blue-50 text-blue-700 ring-blue-700/10 hover:bg-blue-100'"
              :title="isLoadingStaff ? 'Loading staff...' : 'Add Staff Assignment'"
            >
              <i class="mr-1" :class="isLoadingStaff ? 'fas fa-spinner fa-spin' : 'fas fa-plus'"></i>
              Staff
            </button>
          </div>
          <div v-if="project.staffing_assignments && project.staffing_assignments.length > 0" class="space-y-0 divide-y divide-gray-100 border border-gray-200 rounded-lg overflow-hidden">
            <div v-for="assignment in project.staffing_assignments" :key="assignment.ID" 
                  class="flex items-start justify-between gap-x-4 py-2.5 px-3 hover:bg-gray-50">
              <div class="min-w-0 flex-1 flex items-start gap-x-2">
                <StaffAvatar v-if="assignment.employee" :employee="assignment.employee" size="sm" />
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-x-2 mb-0.5">
                    <p class="text-xs font-semibold text-gray-900">
                      {{ assignment.employee?.first_name }} {{ assignment.employee?.last_name }}
                    </p>
                    <span 
                      class="inline-flex items-center rounded-md px-1.5 py-0.5 text-xs font-medium ring-1 ring-inset whitespace-nowrap flex-shrink-0"
                      :class="getStaffingAssignmentStatus(assignment).color">
                      {{ getStaffingAssignmentStatus(assignment).label }}
                    </span>
                  </div>
                  <p v-if="assignment.employee?.title" class="text-xs text-gray-500 mb-1">{{ assignment.employee?.title }}</p>
                  <div class="pl-0 text-xs text-gray-500 mt-0.5">
                    <span v-if="assignment.commitment !== undefined" class="mr-2">
                      <i class="far fa-clock mr-0.5 text-gray-400"></i> 
                      <span class="font-medium">{{ assignment.commitment }} hr/week</span>
                    </span>
                    <span v-if="assignment.start_date || assignment.end_date" class="whitespace-nowrap text-gray-400">
                      <i class="far fa-calendar-alt mr-1"></i>
                      {{ formatDate(assignment.start_date, true) }} - {{ formatDate(assignment.end_date, true) }}
                    </span>
                  </div>
                </div>
              </div>
              <div class="flex-shrink-0 flex items-center gap-x-1.5">
                <button @click.prevent="openEditAssignmentModal(assignment)" type="button" class="p-1 text-xs text-gray-400 hover:text-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 rounded">
                  <span class="sr-only">Edit assignment</span>
                  <i class="fas fa-pencil-alt h-3 w-3"></i>
                </button>
                <!-- Delete button could be added here if needed, or keep it in drawer -->
              </div>
            </div>
          </div>
          <div v-else class="text-xs text-gray-700 italic bg-gray-50 p-3 rounded-lg border border-gray-200 text-center">
            <p>No staff assigned to this project.</p>
          </div>
        </div>
      </div>

      <!-- Column 3: Assets -->
      <div class="space-y-2">
        <div class="flex justify-between items-center mb-2">
            <h5 class="text-xs font-medium text-gray-900">Project Assets</h5>
            <button
              @click="$emit('add-asset', project.ID)"
              class="inline-flex items-center rounded-md bg-sage-50 px-2 py-1 text-xs font-medium text-sage-700 ring-1 ring-inset ring-sage-600/20 hover:bg-sage-100 transition-colors"
              title="Add Asset to this Project"
            >
              <i class="fas fa-plus mr-1"></i>
              Asset
            </button>
        </div>
        <div v-if="project.assets && project.assets.length > 0" class="space-y-1">
          <AssetDisplayItem 
            v-for="asset in sortedAssets.slice(0, maxAssetsToShow)" 
            :key="asset.ID" 
            :asset="asset"
            :project-id="project.ID"
            @delete-asset="handleDeleteAssetLocally"
            @asset-updated="handleAssetUpdate" />
        </div>
        <div v-else class="text-xs text-gray-700 italic bg-gray-50 p-3 rounded-lg border border-gray-200 text-center">
          <p>No assets available for this project.</p>
        </div>
        <div v-if="project.assets && project.assets.length > maxAssetsToShow" class="text-xs text-gray-500 text-center pt-1">
          +{{ project.assets.length - maxAssetsToShow }} more asset(s)
        </div>
      </div>

    </div>
    
    <!-- View Details Button - Full Width -->
    <div class="mt-4 pt-2 border-t border-gray-100">
      <button 
        @click="toggleExpanded"
        class="text-xs flex items-center justify-center w-full px-2 py-1.5 font-medium rounded-md bg-gray-50 text-gray-700 hover:bg-gray-100 transition-colors"
      >
        <span>{{ isExpanded ? 'Hide details' : 'View details' }}</span>
        <i :class="[isExpanded ? 'fa-chevron-up' : 'fa-chevron-down', 'fas ml-2']"></i>
      </button>
    </div>
    
    <!-- Assignment Add/Edit Modal -->
    <AssignmentModal 
      v-if="project"
      :project-id="project.ID" 
      :is-open="isAssignmentModalOpen"
      :assignment-data="editingAssignment" 
      :staff-list="staffList" 
      :project-data="{ active_start: project.active_start, active_end: project.active_end }"
      @close="handleAssignmentModalClose"
      @save="handleAssignmentSave"
    />

    <!-- Expandable Details Section - Full Width -->
    <div 
      v-if="isExpanded" 
      class="mt-3 border-t border-gray-100 pt-3 text-xs"
    >
      <h4 class="font-medium text-gray-900 mb-2">Billable Details</h4>
      
      <!-- Loading State -->
      <div v-if="isLoading" class="py-2 flex justify-center">
        <i class="fas fa-circle-notch fa-spin text-teal"></i>
        <span class="ml-2 text-gray-700">Loading details...</span>
      </div>
      
      <!-- Error State -->
      <div v-else-if="analyticsError" class="py-2 text-center">
        <div class="text-red-500 mb-1.5">{{ analyticsError }}</div>
        <button 
          @click="fetchAnalyticsData" 
          class="text-xs inline-flex items-center px-2 py-1 rounded bg-gray-100 text-gray-700 hover:bg-gray-200"
        >
          <i class="fas fa-sync-alt mr-1"></i> Retry
        </button>
      </div>
      
      <!-- Details Content -->
      <div v-else-if="analytics" class="space-y-4">
        <!-- Project Totals -->
        <div>
          <h5 class="text-xs font-medium text-gray-900 mb-1.5">Project Totals</h5>
          <div class="grid grid-cols-2 gap-3">
            <div class="bg-gray-50 p-2 rounded-lg">
              <div class="text-xs text-gray-500">Total Hours</div>
              <div class="text-sm font-semibold text-gray-900">{{ analytics.total_hours.toFixed(2) }}</div>
            </div>
            <div class="bg-gray-50 p-2 rounded-lg">
              <div class="text-xs text-gray-500">Total Fees</div>
              <div class="text-sm font-semibold text-gray-900">{{ formatCurrency(analytics.total_fees) }}</div>
            </div>
          </div>
        </div>
        
        <!-- Current Billing Period -->
        <div>
          <h5 class="text-xs font-medium text-gray-900 mb-1.5">Current Billing Period</h5>
          <div class="grid grid-cols-2 gap-3">
            <div class="bg-gray-50 p-2 rounded-lg">
              <div class="text-xs text-gray-500">Period Hours</div>
              <div class="text-sm font-semibold text-gray-900">{{ analytics.period_hours ? analytics.period_hours.toFixed(2) : '0.00' }}</div>
            </div>
            <div class="bg-gray-50 p-2 rounded-lg">
              <div class="text-xs text-gray-500">Period Fees</div>
              <div class="text-sm font-semibold text-gray-900">{{ formatCurrency(analytics.period_fees || 0) }}</div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- No Data State -->
      <div v-else class="py-3 text-center text-gray-500">
        <p>No analytics data available for this project.</p>
        <button 
          @click="fetchAnalyticsData" 
          class="text-xs inline-flex items-center px-2 py-1 mt-2 rounded bg-gray-100 text-gray-700 hover:bg-gray-200"
        >
          <i class="fas fa-sync-alt mr-1"></i> Try Again
        </button>
      </div>
    </div> <!-- Close nested v-if="isExpanded" for billable details -->
  </div> <!-- Close first v-if="isExpanded" for main card body -->
  </div> <!-- Close card -->
</template>

<script setup lang="ts">
import { ref, defineProps, defineEmits, onMounted, computed } from 'vue';
import { getProjectAnalytics } from '../../api/projects';
import { fetchRates } from '../../api/rates';
import { getUsers } from '../../api/timesheet';
import type { Project, Staff, StaffingAssignment } from '../../types/Project';
import type { Rate } from '../../types/Rate';
import type { Asset } from '../../types/Asset';
import AssetDisplayItem from '../assets/AssetDisplayItem.vue';
import AssignmentModal from '../assignments/AssignmentModal.vue';
import StaffAvatar from '../StaffAvatar.vue';
import { createProjectAssignment, updateProjectAssignment } from '../../api/projectAssignments';

const props = defineProps<{
  project: Project;
  staffList?: Staff[]; // Staff list passed from parent to avoid redundant API calls
}>();

const emit = defineEmits(['edit', 'add-asset', 'project-asset-updated', 'project-updated']);

const isExpanded = ref(false);
const isLoading = ref(false);
const rates = ref<Rate[]>([]);
const staffList = ref<Staff[]>(props.staffList || []); // Use prop if available
const isLoadingRates = ref(false);
const isLoadingStaff = ref(false);
const maxAssetsToShow = 5;

// State for Assignment Modal
const isAssignmentModalOpen = ref(false);
const editingAssignment = ref<StaffingAssignment | null>(null);

onMounted(async () => {
  isLoadingRates.value = true;
  // Only fetch staff if not provided via props
  if (!props.staffList || props.staffList.length === 0) {
    isLoadingStaff.value = true;
  }
  try {
    rates.value = await fetchRates();
    // Only fetch staff if not provided via props
    if (!props.staffList || props.staffList.length === 0) {
      staffList.value = await getUsers();
    }
  } catch (error) {
    console.error("Error fetching initial data for project card:", error);
  } finally {
    isLoadingRates.value = false;
    isLoadingStaff.value = false;
  }
});

const sortedAssets = computed(() => {
  if (!props.project.assets) return [];
  return [...props.project.assets].sort((a, b) => {
    return (a.name || '').localeCompare(b.name || '');
  });
});

const getRateById = (rateId: number | undefined): Rate | undefined => {
  if (rateId === undefined) return undefined;
  return rates.value.find(r => r.ID === rateId);
};

const getStaffName = (staffId: number | undefined | null): string => {
  if (staffId === undefined || staffId === null) return 'N/A';
  const staffMember = staffList.value.find(s => s.ID === staffId);
  return staffMember ? `${staffMember.first_name} ${staffMember.last_name}` : 'Unknown Staff';
};

const formatDate = (dateInput: string | Date | undefined, short: boolean = false): string => {
  if (!dateInput) return 'N/A';
  try {
    const date = new Date(dateInput);
    if (isNaN(date.getTime())) return 'Invalid Date';
    if (short) return new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric' }).format(date);
    return new Intl.DateTimeFormat('en-US', { year: 'numeric', month: 'short', day: 'numeric' }).format(date);
  } catch (error) {
    return 'Invalid Date';
  }
};

const formatCurrency = (value: number | undefined): string => {
  if (value === undefined) return '$0.00';
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(value);
};

const formatProfit = (clientRate: number, internalRate: number): string => {
    if (clientRate === undefined || internalRate === undefined) return 'N/A';
    const profit = clientRate - internalRate;
    const percentage = clientRate > 0 ? (profit / clientRate) * 100 : 0;
    return `${formatCurrency(profit)} (${percentage.toFixed(0)}%)`;
};

const formatProjectType = (type: string | undefined): string => {
  if (!type) return 'N/A';
  return type.replace('PROJECT_TYPE_', '').replace(/_/g, ' ').split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase()).join(' ');
};

const formatBillingFrequency = (frequency: string | undefined): string => {
  if (!frequency) return 'N/A';
  return frequency.replace('BILLING_TYPE_', '').replace(/_/g, ' ').toLowerCase()
    .split(' ').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ');
};

const formatBudgetFrequency = (billingFrequency: string | undefined): string => {
  if (!billingFrequency) return '';
  switch (billingFrequency) {
    case 'BILLING_TYPE_WEEKLY':
      return ' weekly';
    case 'BILLING_TYPE_BIWEEKLY':
      return ' bi-weekly';
    case 'BILLING_TYPE_MONTHLY':
      return ' monthly';
    case 'BILLING_TYPE_BIMONTHLY':
      return ' bi-monthly';
    default:
      return ''; // For project-based or other frequencies, no suffix needed here
  }
};

const getProjectTimeframe = (project: Project): string => {
  const now = new Date();
  const start = new Date(project.active_start);
  const end = new Date(project.active_end);
  if (end < now) return 'Ended';
  if (start > now) return 'Upcoming';
  return 'Active';
};

const isBillingCodeActive = (code: any): boolean => {
  const now = new Date();
  const start = new Date(code.active_start);
  const end = new Date(code.active_end);
  return start <= now && end >= now;
};

// Function to get status and color for project
const getProjectStatus = (project: Project) => {
  const now = new Date();
  const startDate = new Date(project.active_start);
  const endDate = new Date(project.active_end);

  if (endDate < now) {
    return { label: 'Ended', color: 'bg-gray-100 text-gray-600 ring-gray-500/10' };
  } else if (startDate > now) {
    return { label: 'Upcoming', color: 'bg-blue-50 text-blue-700 ring-blue-600/20' };
  } else {
    return { label: 'Active', color: 'bg-green-50 text-green-700 ring-green-600/20' };
  }
};

const getStaffingAssignmentStatus = (assignment: any) => {
  const now = new Date();
  const startDate = new Date(assignment.start_date);
  const endDate = new Date(assignment.end_date);

  if (!assignment.start_date || !assignment.end_date) {
    return { label: 'Date N/A', color: 'bg-yellow-50 text-yellow-700 ring-yellow-600/20' };
  }

  if (endDate < now) {
    return { label: 'Past', color: 'bg-gray-100 text-gray-600 ring-gray-500/10' };
  } else if (startDate > now) {
    return { label: 'Future', color: 'bg-blue-50 text-blue-700 ring-blue-600/20' };
  } else {
    return { label: 'Current', color: 'bg-green-50 text-green-700 ring-green-600/20' };
  }
};

const handleDeleteAssetLocally = (assetIdToDelet: number) => {
  if (props.project.assets) {
    props.project.assets = props.project.assets.filter(asset => asset.ID !== assetIdToDelet);
    // Optionally, emit an event to parent if the change needs to be persisted or cause wider updates
    // For now, just updating local display based on successful delete in AssetDisplayItem which calls API
  }
};

const handleAssetUpdate = (updatedAsset: Asset) => {
  // When an asset is updated (e.g., URL refreshed), emit an event to the parent (ProjectsView)
  // so it can update the central projects array.
  emit('project-asset-updated', { projectId: props.project.ID, asset: updatedAsset });
};

// Assignment Modal Functions
const openAddAssignmentModal = () => {
  editingAssignment.value = null;
  isAssignmentModalOpen.value = true;
};

const openEditAssignmentModal = (assignment: StaffingAssignment) => {
  // Ensure we are working with a mutable copy and dates are in string YYYY-MM-DD format if needed by modal
  // The modal might handle date formatting internally, but good to be mindful.
  const assignmentCopy = JSON.parse(JSON.stringify(assignment));
  // If modal expects date strings and they are Date objects or complex strings, format them here.
  // For now, assuming modal handles it or they are already suitable strings.
  editingAssignment.value = assignmentCopy;
  isAssignmentModalOpen.value = true;
};

const handleAssignmentModalClose = () => {
  isAssignmentModalOpen.value = false;
  editingAssignment.value = null;
};

type SavedAssignmentData = {
  employee_id: number;
  project_id: number; // This will be sourced from props.project.ID
  commitment?: number | undefined;
  start_date: string; // Assuming string format from modal
  end_date: string;   // Assuming string format from modal
};

const handleAssignmentSave = async (assignmentDataFromModal: Omit<SavedAssignmentData, 'project_id'>) => {
  if (!props.project || !props.project.ID) {
    alert("Cannot save assignment: Project context is missing.");
    return;
  }

  const fullAssignmentData = {
    ...assignmentDataFromModal,
    project_id: props.project.ID,
  };

  try {
    if (editingAssignment.value && editingAssignment.value.ID) {
      await updateProjectAssignment(editingAssignment.value.ID, fullAssignmentData);
      console.log('Assignment updated successfully');
    } else {
      await createProjectAssignment(props.project.ID, fullAssignmentData); // create needs project_id in payload
      console.log('Assignment created successfully');
    }
    emit('project-updated'); // Notify parent to refresh project data
    handleAssignmentModalClose(); 
  } catch (error) {
     console.error("Error saving assignment:", error);
     alert('Failed to save assignment.');
  }
};

// Toggle expanded state
const toggleExpanded = () => {
  isExpanded.value = !isExpanded.value;
  
  if (isExpanded.value) {
    fetchAnalyticsData();
  }
};

const analytics = ref<{
  total_hours: number;
  total_fees: number;
  period_hours: number;
  period_fees: number;
} | null>(null);
const analyticsError = ref<string | null>(null);

// Fetch analytics data
const fetchAnalyticsData = async () => {
  if (!props.project.ID) return;
  
  isLoading.value = true;
  analyticsError.value = null;
  
  try {
    
    const data = await getProjectAnalytics(props.project.ID);
    
    if (!data) {
      analyticsError.value = 'No analytics data available';
      analytics.value = null;
      return;
    }
    
    analytics.value = data;
  } catch (error) {
    console.error('Failed to fetch analytics:', error);
    analyticsError.value = 'Failed to load analytics data';
    analytics.value = null;
  } finally {
    isLoading.value = false;
  }
};
</script>

<style scoped>
/* Additional styles if needed */
.truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style> 