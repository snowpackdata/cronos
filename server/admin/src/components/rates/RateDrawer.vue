<template>
  <TransitionRoot as="template" :show="isOpen">
    <Dialog class="relative z-10" @close="handleClose">
      <div class="fixed inset-0" />

      <div class="fixed inset-0 overflow-hidden">
        <div class="absolute inset-0 overflow-hidden">
          <div class="pointer-events-none fixed inset-y-0 right-0 flex max-w-full pl-10 sm:pl-16">
            <TransitionChild as="template" enter="transform transition ease-in-out duration-500 sm:duration-700" enter-from="translate-x-full" enter-to="translate-x-0" leave="transform transition ease-in-out duration-500 sm:duration-700" leave-from="translate-x-0" leave-to="translate-x-full">
              <DialogPanel class="pointer-events-auto w-screen max-w-2xl">
                <form class="flex h-full flex-col overflow-y-scroll bg-white shadow-xl" @submit.prevent="handleSubmit">
                  <div class="flex-1">
                    <!-- Header -->
                    <div class="bg-sage-dark px-4 py-6 sm:px-6">
                      <div class="flex items-start justify-between space-x-3">
                        <div class="space-y-1">
                          <DialogTitle class="text-base font-semibold text-white">{{ isEditing ? 'Edit Rate' : 'New Rate' }}</DialogTitle>
                          <p class="text-sm text-gray-100">
                            {{ isEditing ? 'Update the rate details below.' : 'Get started by filling in the information below to create a new rate.' }}
                          </p>
                        </div>
                        <div class="flex h-7 items-center">
                          <button type="button" class="relative text-white hover:text-gray-200" @click="handleClose">
                            <span class="absolute -inset-2.5" />
                            <span class="sr-only">Close panel</span>
                            <XMarkIcon class="h-6 w-6" aria-hidden="true" />
                          </button>
                        </div>
                      </div>
                    </div>

                    <!-- Divider container -->
                    <div class="space-y-6 py-6 sm:space-y-0 sm:divide-y sm:divide-gray-200 sm:py-0">
                      <!-- Rate name -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="rate-name" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Rate name</label>
                        </div>
                        <div class="sm:col-span-2">
                          <input 
                            type="text" 
                            name="rate-name" 
                            id="rate-name" 
                            v-model="rate.name"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6" 
                          />
                        </div>
                      </div>

                      <!-- Rate type -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="rate-type" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Rate Type</label>
                        </div>
                        <div class="sm:col-span-2">
                          <select 
                            id="rate-type" 
                            name="rate-type" 
                            v-model="rate.type"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6"
                          >
                            <option value="RATE_TYPE_INTERNAL_CLIENT_NON_BILLABLE">Internal Client Non-Billable</option>
                            <option value="RATE_TYPE_INTERNAL_CLIENT_BILLABLE">Internal Client Billable</option>
                            <option value="RATE_TYPE_INTERNAL_ADMINISTRATIVE">Internal Administrative</option>
                            <option value="RATE_TYPE_INTERNAL_ADMINISTRATIVE_NON_BILLABLE">Internal Administrative Non-Billable</option>
                            <option value="RATE_TYPE_EXTERNAL_CLIENT_BILLABLE">External Client Billable</option>
                            <option value="RATE_TYPE_EXTERNAL_CLIENT_NON_BILLABLE">External Client Non-Billable</option>
                            <option value="RATE_TYPE_INTERNAL_PROJECT">Internal Project</option>
                          </select>
                        </div>
                      </div>
                      
                      <!-- Rate amount -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="rate-amount" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Rate Amount ($)</label>
                        </div>
                        <div class="sm:col-span-2">
                          <input 
                            type="number" 
                            name="rate-amount" 
                            id="rate-amount" 
                            v-model="rate.amount"
                            min="0" 
                            step="0.01"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6" 
                          />
                        </div>
                      </div>

                      <!-- Start date -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="rate-start-date" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Start Date</label>
                        </div>
                        <div class="sm:col-span-2">
                          <input 
                            type="date" 
                            name="rate-start-date" 
                            id="rate-start-date" 
                            v-model="rate.startDate"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6" 
                          />
                        </div>
                      </div>

                      <!-- End date -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="rate-end-date" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">End Date</label>
                        </div>
                        <div class="sm:col-span-2">
                          <input 
                            type="date" 
                            name="rate-end-date" 
                            id="rate-end-date" 
                            v-model="rate.endDate"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6" 
                          />
                        </div>
                      </div>

                      <!-- Internal Only -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="internal-only" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Internal Only</label>
                        </div>
                        <div class="sm:col-span-2 flex items-center">
                          <input 
                            type="checkbox" 
                            name="internal-only" 
                            id="internal-only" 
                            v-model="rate.internal_only"
                            class="h-4 w-4 rounded border-gray-300 text-sage focus:ring-sage" 
                          />
                          <label for="internal-only" class="ml-2 text-sm text-gray-600">
                            Mark this rate as internal only
                          </label>
                        </div>
                      </div>
                    </div>
                  </div>

                  <!-- Action buttons -->
                  <div class="shrink-0 border-t border-gray-200 px-4 py-5 sm:px-6">
                    <div class="flex justify-end space-x-3">
                      <button 
                        type="button" 
                        class="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50" 
                        @click="handleClose"
                      >
                        Cancel
                      </button>
                      <!-- Add delete button only when editing existing rate -->
                      <button 
                        v-if="isEditing"
                        type="button" 
                        class="rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-red-600"
                        @click="handleDelete"
                      >
                        Delete
                      </button>
                      <button 
                        type="submit" 
                        class="inline-flex justify-center rounded-md bg-sage px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-sage-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sage"
                      >
                        {{ isEditing ? 'Update' : 'Create' }}
                      </button>
                    </div>
                  </div>
                </form>
              </DialogPanel>
            </TransitionChild>
          </div>
        </div>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue';
import { XMarkIcon } from '@heroicons/vue/24/outline';
import { formatDate, parseServerDate, formatDateForServer, getCurrentDate } from '../../utils/dateUtils';

const props = defineProps({
  isOpen: {
    type: Boolean,
    default: false
  },
  rateData: {
    type: Object,
    default: () => ({})
  }
});

const emit = defineEmits(['close', 'save', 'delete']);

// Handle close events
const handleClose = () => {
  emit('close');
};

// Handle delete button click
const handleDelete = () => {
  if (window.confirm('Are you sure you want to delete this rate? This action cannot be undone and may affect billing codes that use this rate.')) {
    emit('delete', props.rateData);
  }
};

// Determine if editing or creating new
const isEditing = computed(() => !!props.rateData?.ID);

// Initialize rate with default values or provided data
const rate = ref({
  ID: props.rateData?.ID || null,
  name: props.rateData?.name || '',
  type: props.rateData?.type || 'RATE_TYPE_EXTERNAL_CLIENT_BILLABLE',
  amount: props.rateData?.amount || 0,
  startDate: props.rateData?.active_from ? parseServerDate(props.rateData.active_from) : getCurrentDate(),
  endDate: props.rateData?.active_to ? parseServerDate(props.rateData.active_to) : '',
  internal_only: props.rateData?.internal_only || false
});

// Update rate data when rateData prop changes
watch(() => props.rateData, (newVal) => {
  if (newVal) {
    rate.value = {
      ID: newVal.ID || null,
      name: newVal.name || '',
      type: newVal.type || 'RATE_TYPE_EXTERNAL_CLIENT_BILLABLE',
      amount: newVal.amount || 0,
      startDate: newVal.active_from ? parseServerDate(newVal.active_from) : getCurrentDate(),
      endDate: newVal.active_to ? parseServerDate(newVal.active_to) : ''
    };
  } else {
    // Reset to empty values if no rate data is provided
    resetForm();
  }
}, { deep: true });

// Handle form submission
const handleSubmit = () => {
  
  // Validate form
  if (!rate.value.name || !rate.value.amount || !rate.value.startDate) {
    alert('Please fill in all required fields');
    console.error('Validation failed - missing required fields', {
      name: rate.value.name,
      amount: rate.value.amount,
      startDate: rate.value.startDate
    });
    return;
  }

  // Format the datetime ranges for proper display and API submission
  const formattedRate = computed(() => {
    // Start with base data
    const formattedData = {
      ID: rate.value.ID,
      name: rate.value.name.trim(),
      type: rate.value.type,
      amount: parseFloat(rate.value.amount.toString()),
      active_from: formatDateForServer(rate.value.startDate),
      active_to: formatDateForServer(rate.value.endDate)
    };
    
    return formattedData;
  });
  
  // For date inputs (which are already in YYYY-MM-DD format),
  // we can directly set the values to avoid any timezone issues
  if (rate.value.startDate) {
    // Directly set the YYYY-MM-DD string from the date input
    formattedRate.value.active_from = rate.value.startDate;
  } else {
    // Fallback to today if somehow missing
    const today = new Date();
    formattedRate.value.active_from = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;
  }
  
  // Only set endDate if provided
  if (rate.value.endDate) {
    // Directly set the YYYY-MM-DD string from the date input
    formattedRate.value.active_to = rate.value.endDate;
  } else {
    formattedRate.value.active_to = '';
  }
  
  emit('save', formattedRate.value);
};
</script> 