<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import type { Rate } from '../../types/Rate';
import { fetchRates, updateRate, createRate, deleteRate } from '../../api/rates';
// @ts-ignore - Ignore type issues with Vue components for now
import RateDrawer from '../../components/rates/RateDrawer.vue';
// @ts-ignore - Ignore type issues with Vue components for now
import ConfirmationModal from '../../components/ConfirmationModal.vue';
import { formatDate } from '../../utils/dateUtils';

// State
const rates = ref<Rate[]>([]);
const isLoading = ref(true);
const error = ref<string | null>(null);
const showDrawer = ref(false);
const selectedRate = ref<Rate | null>(null);
const showDeleteConfirm = ref(false);

// Computed property to sort rates by active status
const sortedRates = computed(() => {
  if (!rates.value || !Array.isArray(rates.value)) {
    return [];
  }
  
  // Return a new sorted array
  return [...rates.value].sort((a, b) => {
    const isAActive = isRateActive(a);
    const isBActive = isRateActive(b);
    
    // Sort active rates first, then inactive
    if (isAActive && !isBActive) return -1;
    if (!isAActive && isBActive) return 1;
    
    // If both have the same active status, keep their original order
    // Alternatively, we could sort by name, ID, or date
    return 0;
  });
});

// Mock API - This would be replaced with actual API calls
// Commented out as it's not being used
/*
const ratesApi = {
  async getRates(): Promise<Rate[]> {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 1000));
    return [
      // ... mock data ...
    ];
  }
};
*/

// Fetch rates on component mount
onMounted(async () => {
  await fetchRatesData();
});

// Fetch all rates
const fetchRatesData = async () => {
  isLoading.value = true;
  error.value = null;
  
  try {
    
    // Add direct fetch attempt for debugging
    try {
      await fetch('/api/rates');
    } catch (directErr) {
      console.error('Direct fetch error:', directErr);
    }
    
    // Use the exported fetchRates function
    const response = await fetchRates();
    
    if (!response || !Array.isArray(response)) {
      console.error('Invalid response format - expected array but got:', typeof response);
      error.value = 'Invalid response format from API';
      rates.value = [];
      return;
    }
    
    rates.value = response;
  } catch (err) {
    console.error('Error fetching rates:', err);
    console.error('Error details:', err instanceof Error ? err.message : String(err));
    if (err instanceof Error && err.stack) {
      console.error('Stack trace:', err.stack);
    }
    error.value = 'Failed to load rates. Please try again.';
    rates.value = [];
  } finally {
    isLoading.value = false;
  }
};

// Save rate
const saveRate = async (rateData: Rate) => {
  try {
    if (rateData.ID && rateData.ID > 0) {
      // Use the exported updateRate function
      await updateRate(rateData);
    } else {
      // Use the exported createRate function
      await createRate(rateData);
    }
    
    // Refresh rates
    await fetchRatesData();
    
    // Close drawer
    closeDrawer();
  } catch (error) {
    console.error('Error saving rate:', error);
    alert('Failed to save rate. Please try again.');
  }
};

// Format currency
const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD'
  }).format(amount);
};

// Drawer functions
const openDrawer = (rate: Rate | null = null) => {
  selectedRate.value = rate;
  showDrawer.value = true;
};

const closeDrawer = () => {
  showDrawer.value = false;
  selectedRate.value = null;
};

const handleDeleteFromDrawer = async (rate: Rate) => {
  if (!rate) return;
  
  try {
    // Use the exported deleteRate function
    await deleteRate(Number(rate.ID));
    
    // Refresh rates
    await fetchRatesData();
    // Close drawer
    closeDrawer();
  } catch (error) {
    console.error('Error deleting rate:', error);
    alert('Failed to delete rate. Please try again.');
  }
};

// Helper function to check if a rate is currently active
const isRateActive = (rate: Rate): boolean => {
  // Ensure we have a valid rate object
  if (!rate) return false;
  
  // Get current date at midnight UTC
  const now = new Date();
  now.setUTCHours(0, 0, 0, 0);
  
  // Parse start and end dates
  let startDate: Date | null = null;
  let endDate: Date | null = null;
  
  // Handle active_from date
  if (rate.active_from) {
    if (typeof rate.active_from === 'string' && rate.active_from.match(/^\d{4}-\d{2}-\d{2}$/)) {
      // For YYYY-MM-DD strings, parse in UTC
      const [year, month, day] = rate.active_from.split('-').map(Number);
      startDate = new Date(Date.UTC(year, month - 1, day));
    } else {
      startDate = new Date(rate.active_from);
    }
    // Normalize to midnight UTC
    startDate.setUTCHours(0, 0, 0, 0);
  }
  
  // Handle active_to date
  if (rate.active_to) {
    if (typeof rate.active_to === 'string' && rate.active_to.match(/^\d{4}-\d{2}-\d{2}$/)) {
      // For YYYY-MM-DD strings, parse in UTC
      const [year, month, day] = rate.active_to.split('-').map(Number);
      endDate = new Date(Date.UTC(year, month - 1, day));
    } else {
      endDate = new Date(rate.active_to);
    }
    // Normalize to midnight UTC
    endDate.setUTCHours(0, 0, 0, 0);
  }
  
  // Rate is active if: 
  // 1. Current date is after or equal to the start date (if a start date exists)
  // 2. Current date is before or equal to the end date (if an end date exists)
  const isAfterStart = startDate ? now >= startDate : true;
  
  // IMPORTANT: For end dates, we want to include the entire day as active
  // So we compare with the day AFTER the end date
  const isBeforeEnd = endDate ? now <= endDate : true;
  
  return isAfterStart && isBeforeEnd;
};
</script>

<template>
  <div class="px-4 py-6 sm:px-6 lg:px-8">
    <div class="sm:flex sm:items-center">
      <div class="sm:flex-auto">
        <h1 class="text-base font-semibold leading-6 text-gray-900">Rates</h1>
        <p class="mt-2 text-sm text-gray-700">
          A list of all rates including their name, amount, and effective dates.
        </p>
      </div>
      <div class="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
        <button
          type="button"
          @click="openDrawer()"
          class="block rounded-md bg-sage px-3 py-2 text-center text-sm font-semibold text-white shadow-sm hover:bg-sage-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sage"
        >
          Create new rate
        </button>
      </div>
    </div>
    
    <!-- Loading state -->
    <div v-if="isLoading" class="flex flex-col items-center justify-center p-10 bg-white rounded-lg shadow mt-6">
      <i class="fas fa-spinner fa-spin text-4xl text-teal mb-4"></i>
      <span class="text-gray-dark">Loading rates...</span>
    </div>
    
    <!-- Error state -->
    <div v-else-if="error" class="flex flex-col items-center justify-center p-10 bg-white rounded-lg shadow mt-6">
      <i class="fas fa-exclamation-circle text-4xl text-red mb-4"></i>
      <span class="text-gray-dark mb-2">{{ error }}</span>
      <button @click="fetchRatesData" class="btn-secondary mt-4">
        <i class="fas fa-sync mr-2"></i> Retry
      </button>
    </div>
    
    <!-- Rate List (New Design) -->
    <div v-else class="bg-white shadow rounded-lg mt-8">
      <ul role="list" class="divide-y divide-gray-100">
        <li v-for="rate in sortedRates" :key="rate.ID" class="flex items-center justify-between gap-x-6 py-5 px-4 hover:bg-gray-50">
          <div class="min-w-0 flex-1">
            <div class="flex items-start gap-x-3">
              <p class="text-sm/6 font-semibold text-gray-900">{{ rate.name }}</p>
              <p :class="[
                isRateActive(rate) ? 'text-green-700 bg-green-50 ring-green-600/20' : 'text-red-700 bg-red-50 ring-red-600/20',
                'mt-0.5 whitespace-nowrap rounded-md px-1.5 py-0.5 text-xs font-medium ring-1 ring-inset'
              ]">
                {{ isRateActive(rate) ? 'Active' : 'Inactive' }}
              </p>
              <p v-if="rate.internal_only" class="text-gray-600 bg-gray-50 ring-gray-500/10 mt-0.5 whitespace-nowrap rounded-md px-1.5 py-0.5 text-xs font-medium ring-1 ring-inset">
                Internal Only
              </p>
            </div>
            <div class="mt-1 flex items-center gap-x-2 text-xs/5 text-gray-500">
              <p class="whitespace-nowrap">
                <span class="font-medium">{{ formatCurrency(rate.amount) }}</span>
              </p>
              <svg viewBox="0 0 2 2" class="size-0.5 fill-current">
                <circle cx="1" cy="1" r="1" />
              </svg>
              <p class="whitespace-nowrap">
                Active from <time :datetime="rate.active_from">{{ formatDate(rate.active_from) }}</time>
              </p>
              <svg v-if="rate.active_to" viewBox="0 0 2 2" class="size-0.5 fill-current">
                <circle cx="1" cy="1" r="1" />
              </svg>
              <p v-if="rate.active_to" class="whitespace-nowrap">
                to <time :datetime="rate.active_to">{{ formatDate(rate.active_to) }}</time>
              </p>
            </div>
          </div>
          <div class="flex flex-none items-center gap-x-4">
            <button
              @click="openDrawer(rate)"
              class="rounded-md bg-white px-2.5 py-1.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
            >
              <i class="fas fa-pencil-alt mr-1"></i> Edit
            </button>
          </div>
        </li>
        <li v-if="sortedRates.length === 0" class="py-5 px-4">
          <div class="flex flex-col items-center justify-center p-10">
            <i class="fas fa-calculator text-5xl text-gray-300 mb-4"></i>
            <p class="text-lg font-medium text-gray-dark">No rates found</p>
            <p class="text-gray mb-4">Click "Create new rate" to add one</p>
          </div>
        </li>
      </ul>
    </div>
    
    <!-- Rate Drawer -->
    <RateDrawer
      :is-open="showDrawer"
      :rate-data="selectedRate"
      @close="closeDrawer"
      @save="saveRate"
      @delete="handleDeleteFromDrawer"
    />
    
    <!-- Delete Confirmation Modal -->
    <ConfirmationModal
      :show="showDeleteConfirm"
      title="Delete Rate"
      message="Are you sure you want to delete this rate? This action cannot be undone and may affect billing codes that use this rate."
      @confirm="selectedRate && handleDeleteFromDrawer(selectedRate)"
      @cancel="showDeleteConfirm = false"
    />
  </div>
</template>

<style scoped>
/* Buttons */
.btn-primary {
  display: inline-flex;
  align-items: center;
  padding: 0.5rem 1rem;
  border: 1px solid transparent;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: white;
  background-color: #2563eb;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
}

.btn-primary:hover {
  background-color: #1d4ed8;
}

.btn-secondary {
  display: inline-flex;
  align-items: center;
  padding: 0.5rem 1rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
  background-color: white;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
}

.btn-secondary:hover {
  background-color: #f9fafb;
}

.btn-danger {
  display: inline-flex;
  align-items: center;
  padding: 0.5rem 1rem;
  border: 1px solid transparent;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: white;
  background-color: #dc2626;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
}

.btn-danger:hover {
  background-color: #b91c1c;
}
</style> 