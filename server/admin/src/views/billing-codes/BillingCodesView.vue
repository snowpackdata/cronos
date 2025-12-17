<template>
  <div class="px-4 py-6 sm:px-6 lg:px-8">
    <div class="sm:flex sm:items-center">
      <div class="sm:flex-auto">
        <h1 class="text-base font-semibold leading-6 text-gray-900">Billing Codes</h1>
        <p class="mt-2 text-sm text-gray-700">
          A list of all billing codes associated with projects and rates.
        </p>
      </div>
      <div class="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
        <button
          type="button"
          @click="openBillingCodeDrawer()"
          class="block rounded-md bg-sage px-3 py-2 text-center text-sm font-semibold text-white shadow-sm hover:bg-sage-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sage"
        >
          Create new billing code
        </button>
      </div>
    </div>
    
    <!-- Project Filter (Commented out for now, can be re-enabled if needed) -->
    <!-- <div class="mt-4">
      <label for="project-filter" class="block text-sm font-medium text-gray-700">Filter by Project</label>
      <select
        id="project-filter"
        v-model="selectedProjectId"
        class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-sage focus:border-sage sm:text-sm rounded-md"
        @change="handleProjectChange"
      >
        <option value="">All Projects</option>
        <option v-for="project in projects" :key="project.id" :value="project.id">
          {{ project.name }}
        </option>
      </select>
    </div> -->
    
    <!-- Loading state -->
    <div v-if="isLoading" class="flex flex-col items-center justify-center p-10 bg-white rounded-lg shadow mt-6">
      <i class="fas fa-spinner fa-spin text-4xl text-teal mb-4"></i>
      <span class="text-gray-dark">Loading billing codes...</span>
    </div>
    
    <!-- Error state -->
    <div v-else-if="error" class="flex flex-col items-center justify-center p-10 bg-white rounded-lg shadow mt-6">
      <i class="fas fa-exclamation-circle text-4xl text-red mb-4"></i>
      <span class="text-gray-dark mb-2">{{ error }}</span>
      <button @click="fetchBillingCodesData" class="btn-secondary mt-4">
        <i class="fas fa-sync mr-2"></i> Retry
      </button>
    </div>
    
    <!-- Billing Codes List (New Design, similar to RatesView) -->
    <div v-else class="bg-white shadow rounded-lg mt-8">
      <ul role="list" class="divide-y divide-gray-100">
        <li v-for="billingCode in sortedBillingCodes" :key="billingCode.ID" class="flex items-center justify-between gap-x-6 py-5 px-4 hover:bg-gray-50">
          <div class="min-w-0 flex-1">
            <div class="flex items-start gap-x-3">
              <p class="text-sm/6 font-semibold text-gray-900">{{ billingCode.name }}</p>
              <p :class="[
                isBillingCodeActive(billingCode) ? 'text-green-700 bg-green-50 ring-green-600/20' : 'text-red-700 bg-red-50 ring-red-600/20',
                'mt-0.5 whitespace-nowrap rounded-md px-1.5 py-0.5 text-xs font-medium ring-1 ring-inset'
              ]">
                {{ isBillingCodeActive(billingCode) ? 'Active' : 'Inactive' }}
              </p>
              <p class="mt-0.5 whitespace-nowrap rounded-md px-1.5 py-0.5 text-xs font-medium bg-indigo-50 text-indigo-700 ring-1 ring-inset ring-indigo-600/20">
                {{ formatCategory(billingCode.category) }}
              </p>
            </div>
            <div class="mt-1 flex items-center gap-x-2 text-xs/5 text-gray-500">
              <p class="whitespace-nowrap">
                <span class="font-medium">Code: {{ billingCode.code }}</span>
              </p>
              <svg viewBox="0 0 2 2" class="size-0.5 fill-current">
                <circle cx="1" cy="1" r="1" />
              </svg>
              <p class="whitespace-nowrap">
                Project: <span class="font-medium">{{ getProjectName(billingCode.project) }}</span>
              </p>
              <svg viewBox="0 0 2 2" class="size-0.5 fill-current">
                <circle cx="1" cy="1" r="1" />
              </svg>
              <p class="whitespace-nowrap">
                Rate: <span class="font-medium">{{ getRateName(billingCode.rate_id) }}</span>
              </p>
            </div>
            <div v-if="billingCode.active_start || billingCode.active_end" class="mt-1 flex items-center gap-x-2 text-xs/5 text-gray-500">
              <p v-if="billingCode.active_start" class="whitespace-nowrap">
                Active from <time :datetime="billingCode.active_start">{{ formatDate(billingCode.active_start) }}</time>
              </p>
              <svg v-if="billingCode.active_start && billingCode.active_end" viewBox="0 0 2 2" class="size-0.5 fill-current">
                <circle cx="1" cy="1" r="1" />
              </svg>
              <p v-if="billingCode.active_end" class="whitespace-nowrap">
                to <time :datetime="billingCode.active_end">{{ formatDate(billingCode.active_end) }}</time>
              </p>
            </div>
          </div>
          <div class="flex flex-none items-center gap-x-4">
            <button
              @click="openBillingCodeDrawer(billingCode)"
              class="rounded-md bg-white px-2.5 py-1.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
            >
              <i class="fas fa-pencil-alt mr-1"></i> Edit
            </button>
          </div>
        </li>
        <li v-if="sortedBillingCodes.length === 0" class="py-5 px-4">
          <div class="flex flex-col items-center justify-center p-10">
            <i class="fas fa-tags text-5xl text-gray-300 mb-4"></i>
            <p class="text-lg font-medium text-gray-dark">No billing codes found</p>
            <p class="text-gray mb-4">Click "Create new billing code" to add one</p>
          </div>
        </li>
      </ul>
    </div>
    
    <!-- Billing Code Drawer -->
    <BillingCodeDrawer
      :is-open="isBillingCodeDrawerOpen"
      :billing-code-data="selectedBillingCode"
      @close="closeBillingCodeDrawer"
      @save="saveBillingCode"
      @delete="confirmDelete"
    />
    
    <!-- Delete Confirmation Modal -->
    <ConfirmationModal
      :show="showDeleteModal"
      title="Delete Billing Code"
      message="Are you sure you want to delete this billing code? This action cannot be undone."
      @confirm="deleteBillingCode"
      @cancel="showDeleteModal = false"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { 
  billingCodesAPI,
  getBillingCodes, 
  fetchProjects, 
  fetchRates, 
  createBillingCode, 
  updateBillingCode, 
  deleteBillingCode as deleteBillingCodeAPI 
} from '../../api';
import BillingCodeDrawer from '../../components/billing-codes/BillingCodeDrawer.vue';
import ConfirmationModal from '../../components/ConfirmationModal.vue';
import { formatDate, parseServerDate, getCurrentDate, formatDateForServer } from '../../utils/dateUtils';
import { createWithFormData, updateWithFormData } from '../../api/apiUtils';

// State
const billingCodes = ref([]);
const projects = ref([]);
const rates = ref([]);
const isBillingCodeDrawerOpen = ref(false);
const selectedBillingCode = ref(null);
const showDeleteModal = ref(false);
const billingCodeToDelete = ref(null);
const selectedProjectId = ref('');
const isLoading = ref(true);
const error = ref(null);

// Computed property to sort billing codes by active status
const sortedBillingCodes = computed(() => {
  if (!billingCodes.value || !Array.isArray(billingCodes.value)) {
    return [];
  }
  
  // Return a new sorted array
  return [...billingCodes.value].sort((a, b) => {
    const isAActive = isBillingCodeActive(a);
    const isBActive = isBillingCodeActive(b);
    
    // Sort active billing codes first, then inactive
    if (isAActive && !isBActive) return -1;
    if (!isAActive && isBActive) return 1;
    
    // If both have the same active status, keep their original order
    return 0;
  });
});

// Fetch data
onMounted(async () => {
  await fetchBillingCodesData();
});

// Fetch billing codes data
const fetchBillingCodesData = async () => {
  isLoading.value = true;
  error.value = null;
  
  try {
    // Fetch billing codes, projects, and rates
    const [billingCodesData, projectsData, ratesData] = await Promise.all([
      getBillingCodes(),
      fetchProjects(),
      fetchRates()
    ]);
    
    billingCodes.value = billingCodesData || [];
    projects.value = projectsData || [];
    rates.value = ratesData || [];
  } catch (err) {
    console.error('Error loading data:', err);
    error.value = 'Failed to load billing codes. Please try again.';
  } finally {
    isLoading.value = false;
  }
};

// Helper functions
const getProjectName = (projectId) => {
  if (!projects.value || !Array.isArray(projects.value)) {
    return 'Unknown';
  }
  const project = projects.value.find(p => p.ID === projectId);
  return project ? project.name : 'Unknown';
};

const getRateName = (rateId) => {
  if (!rates.value || !Array.isArray(rates.value)) {
    return 'Unknown';
  }
  const rate = rates.value.find(r => r.ID === rateId);
  if (!rate) return 'Unknown';
  return `${rate.name} (${formatCurrency(rate.amount)})`;
};

const formatCurrency = (amount) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD'
  }).format(amount);
};

const formatCategory = (category) => {
  if (!category) return 'Unknown';
  
  // Define a mapping for category constants to human-readable names
  const categoryMap = {
    'BILLING_CODE_CATEGORY_DEVELOPMENT': 'Development',
    'BILLING_CODE_CATEGORY_ANALYSIS': 'Analysis',
    'BILLING_CODE_CATEGORY_AI': 'AI & Machine Learning',
    'BILLING_CODE_CATEGORY_CONSULTING': 'Consulting',
    'BILLING_CODE_CATEGORY_OPERATIONS': 'Operations',
    'BILLING_CODE_CATEGORY_ADMINISTRATIVE': 'Administrative',
    'BILLING_CODE_CATEGORY_SUPPORT': 'Support',
    'BILLING_CODE_CATEGORY_DESIGN': 'Design',
    'BILLING_CODE_CATEGORY_TESTING': 'Testing & QA',
    'BILLING_CODE_CATEGORY_MANAGEMENT': 'Project Management',
    'BILLING_CODE_CATEGORY_EXTERNAL_CLIENT': 'External Client',
  };
  
  // Check if we have a direct mapping
  if (categoryMap[category]) {
    return categoryMap[category];
  }
  
  // Fall back to converting from enum-like format to human-readable if needed
  // This handles both legacy data and new format
  const formattedCategory = category
    .replace('BILLING_CODE_CATEGORY_', '')
    .replace(/_/g, ' ')
    .toLowerCase();
  
  return formattedCategory.charAt(0).toUpperCase() + formattedCategory.slice(1);
};

// Project filter
const handleProjectChange = async () => {
  try {
    isLoading.value = true;
    const billingCodesData = await getBillingCodes(selectedProjectId.value || undefined);
    billingCodes.value = billingCodesData || [];
  } catch (error) {
    console.error('Error filtering billing codes:', error);
  } finally {
    isLoading.value = false;
  }
};

// Helper function to check if a billing code is currently active
const isBillingCodeActive = (billingCode) => {
  // Ensure we have a valid billing code object
  if (!billingCode) return false;
  
  // Parse start and end dates using our utility functions
  const startDate = billingCode.active_start ? new Date(billingCode.active_start) : null;
  const endDate = billingCode.active_end ? new Date(billingCode.active_end) : null;
  
  // Get current date at midnight in UTC
  const now = new Date();
  now.setUTCHours(0, 0, 0, 0);
  
  // Billing code is active if: 
  // 1. Current date is after or equal to the start date (if a start date exists)
  // 2. Current date is before or equal to the end date (if an end date exists)
  const isAfterStart = startDate ? now >= startDate : true;
  const isBeforeEnd = endDate ? now <= endDate : true;
  
  return isAfterStart && isBeforeEnd;
};

// Drawer functions
const openBillingCodeDrawer = (billingCode = null) => {
  if (billingCode) {
    selectedBillingCode.value = {
      id: billingCode.id,
      ID: billingCode.ID,
      name: billingCode.name,
      code: billingCode.code || '',
      category: billingCode.category,
      project: billingCode.project,
      projectId: billingCode.project,
      rate_id: billingCode.rate_id,
      rateId: billingCode.rate_id,
      active_start: billingCode.active_start || getCurrentDate(),
      active_end: billingCode.active_end || '',
      internal_rate_id: billingCode.internal_rate_id || ''
    };
    
  } else {
    selectedBillingCode.value = null;
  }
  
  isBillingCodeDrawerOpen.value = true;
};

const closeBillingCodeDrawer = () => {
  isBillingCodeDrawerOpen.value = false;
  selectedBillingCode.value = null;
};

// Save billing code
const saveBillingCode = async (billingCodeData) => {
  try {
    // Check if the code already exists (for uniqueness)
    const codeExists = billingCodes.value.some(code => 
      code.code === billingCodeData.code && 
      code.ID !== billingCodeData.id // Skip checking against the current billing code when editing
    );
    
    if (codeExists) {
      alert('This billing code ID already exists. Please choose a different ID.');
      return;
    }
    
    
    // Ensure rateId is a valid number, default to 0 if undefined/null
    const rateId = billingCodeData.rateId !== undefined ? Number(billingCodeData.rateId) : 
                   billingCodeData.rate_id !== undefined ? Number(billingCodeData.rate_id) : 0;
    
    // Ensure internal_rate_id is a valid number, default to 0 if undefined/null
    const internalRateId = billingCodeData.internal_rate_id !== undefined ? Number(billingCodeData.internal_rate_id) : 0;
    
    
    // Create a FormData object directly with all required fields
    const formData = new FormData();
    formData.append('name', billingCodeData.name || '');
    formData.append('code', billingCodeData.code || '');
    formData.append('type', rateId === internalRateId ? 'RATE_TYPE_INTERNAL_PROJECT' : 'RATE_TYPE_EXTERNAL_CLIENT_BILLABLE');
    formData.append('category', billingCodeData.category || '');
    
    // Project ID - ensure it's a string and valid
    if (billingCodeData.projectId) {
      formData.append('project_id', String(billingCodeData.projectId));
    }
    
    // Always include both rate IDs as strings
    formData.append('rate_id', String(rateId));
    formData.append('internal_rate_id', String(internalRateId));
    
    // Dates
    if (billingCodeData.active_start) {
      formData.append('active_start', billingCodeData.active_start);
    }
    
    if (billingCodeData.active_end) {
      formData.append('active_end', billingCodeData.active_end);
    }
    
    // Default rounded_to to 15 minutes
    formData.append('rounded_to', '15');
    
    // Log the form data for debugging
    for (const [key, value] of formData.entries()) {
    }
    
    let newBillingCode;
    if (billingCodeData.id) {
      // Update existing billing code
      newBillingCode = await updateWithFormData('billing_codes', billingCodeData.id, formData);
    } else {
      // Create new billing code
      newBillingCode = await createWithFormData('billing_codes', formData);
    }
    
    
    // Refresh billing codes
    await fetchBillingCodesData();
    
    // Close drawer
    closeBillingCodeDrawer();
  } catch (error) {
    console.error('Error saving billing code:', error);
    alert('Failed to save billing code. Please try again.');
  }
};

// Delete billing code
const confirmDelete = (billingCode) => {
  if (!billingCode) {
    console.error('No billing code provided for deletion');
    return;
  }
  
  // Ensure the billing code has a proper ID for deletion
  billingCodeToDelete.value = {
    ID: billingCode.ID || billingCode.id,
    id: billingCode.id || billingCode.ID,
    name: billingCode.name
  };
  
  showDeleteModal.value = true;
  // Close the drawer when confirming delete
  isBillingCodeDrawerOpen.value = false;
};

const deleteBillingCode = async () => {
  if (!billingCodeToDelete.value) return;
  
  try {
    // Use either ID or id property, whichever exists
    const billingCodeId = billingCodeToDelete.value.ID || billingCodeToDelete.value.id;
    
    if (!billingCodeId) {
      console.error('Invalid billing code ID for deletion:', billingCodeToDelete.value);
      alert('Failed to delete billing code: Invalid ID');
      return;
    }
    
    await deleteBillingCodeAPI(billingCodeId);
    
    // Refresh billing codes
    await fetchBillingCodesData();
    
    // Close modal
    showDeleteModal.value = false;
    billingCodeToDelete.value = null;
  } catch (error) {
    console.error('Error deleting billing code:', error);
    alert('Failed to delete billing code. Please try again.');
  }
};
</script>

<style scoped>
.btn-secondary {
  display: inline-flex;
  align-items: center;
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-gray-700);
  background-color: white;
  border: 1px solid var(--color-gray-300);
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background-color: var(--color-gray-50);
}
</style> 