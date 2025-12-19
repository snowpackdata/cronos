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
                    <div class="bg-sage-dark px-4 py-4 sm:px-6">
                      <div class="flex items-start justify-between space-x-3">
                        <div class="space-y-1">
                          <DialogTitle class="text-base font-semibold text-white">{{ isEditing ? 'Edit Account' : 'New Account' }}</DialogTitle>
                          <p class="text-sm text-gray-100">
                            {{ isEditing ? 'Update the account details below.' : 'Get started by filling in the information below to create a new account.' }}
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
                    <div class="space-y-4 py-4 sm:space-y-0 sm:divide-y sm:divide-gray-200 sm:py-0">
                      <!-- Account name -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-3">
                        <div>
                          <label for="account-name" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Account name</label>
                        </div>
                        <div class="sm:col-span-2">
                          <input 
                            type="text" 
                            name="account-name" 
                            id="account-name" 
                            v-model="account.name"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6" 
                          />
                        </div>
                      </div>

                      <!-- Legal name -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-3">
                        <div>
                          <label for="account-legal-name" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Legal name</label>
                        </div>
                        <div class="sm:col-span-2">
                          <input 
                            type="text" 
                            name="account-legal-name" 
                            id="account-legal-name" 
                            v-model="account.legal_name"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6" 
                          />
                        </div>
                      </div>

                      <!-- Account type -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-3">
                        <div>
                          <label for="account-type" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Account Type</label>
                        </div>
                        <div class="sm:col-span-2">
                          <select 
                            id="account-type" 
                            name="account-type" 
                            v-model="account.type"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6"
                          >
                            <option value="ACCOUNT_TYPE_CLIENT">Client</option>
                            <option value="ACCOUNT_TYPE_INTERNAL">Internal</option>
                          </select>
                        </div>
                      </div>
                      
                      <!-- Email -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-3">
                        <div>
                          <label for="account-email" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Email</label>
                        </div>
                        <div class="sm:col-span-2">
                          <input 
                            type="email" 
                            name="account-email" 
                            id="account-email" 
                            v-model="account.email"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6" 
                          />
                        </div>
                      </div>

                      <!-- Website -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-3">
                        <div>
                          <label for="account-website" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Website</label>
                        </div>
                        <div class="sm:col-span-2">
                          <input 
                            type="url" 
                            name="account-website" 
                            id="account-website" 
                            v-model="account.website"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6" 
                          />
                        </div>
                      </div>

                      <!-- Address -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-3">
                        <div>
                          <label for="account-address" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Address</label>
                        </div>
                        <div class="sm:col-span-2">
                          <textarea 
                            rows="3" 
                            name="account-address" 
                            id="account-address" 
                            v-model="account.address"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6" 
                          />
                        </div>
                      </div>

                      <!-- Logo Upload (Internal Accounts Only) -->
                      <div v-if="account.type === 'ACCOUNT_TYPE_INTERNAL'" class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-3">
                        <div>
                          <label for="account-logo" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Company Logo</label>
                          <p class="text-xs text-gray-500 mt-1">For invoices and bills</p>
                        </div>
                        <div class="sm:col-span-2">
                          <div v-if="account.logo_asset?.url" class="mb-3">
                            <div class="text-xs text-gray-500 mb-2">Current logo:</div>
                            <img :src="account.logo_asset.url" alt="Current logo" class="h-16 w-auto border border-gray-200 rounded p-2 bg-white">
                          </div>
                          <input 
                            type="file" 
                            id="account-logo" 
                            name="account-logo"
                            accept="image/png,image/jpeg,image/svg+xml"
                            @change="handleLogoChange"
                            class="block w-full text-sm text-gray-900 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-sage-50 file:text-sage-700 hover:file:bg-sage-100"
                          />
                          <p class="mt-1 text-xs text-gray-500">PNG, JPG, or SVG. Max 10MB.</p>
                        </div>
                      </div>

                      <!-- Billing Frequency -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-3">
                        <div>
                          <label for="account-billing-frequency" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Billing Frequency</label>
                        </div>
                        <div class="sm:col-span-2">
                          <select 
                            id="account-billing-frequency" 
                            name="account-billing-frequency" 
                            v-model="account.billing_frequency"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6"
                          >
                            <option value="BILLING_TYPE_WEEKLY">Weekly</option>
                            <option value="BILLING_TYPE_BIWEEKLY">Bi-Weekly</option>
                            <option value="BILLING_TYPE_MONTHLY">Monthly</option>
                            <option value="BILLING_TYPE_BIMONTHLY">Bi-Monthly</option>
                            <option value="BILLING_TYPE_PROJECT">Project-Based</option>
                          </select>
                        </div>
                      </div>

                      <!-- Projects Single Invoice -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-3">
                        <div>
                          <label for="account-projects-single-invoice" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Invoice Type</label>
                        </div>
                        <div class="sm:col-span-2 flex items-center">
                          <div class="flex items-center">
                            <input 
                              type="checkbox" 
                              id="account-projects-single-invoice" 
                              name="account-projects-single-invoice" 
                              v-model="account.projects_single_invoice"
                              class="h-4 w-4 text-sage focus:ring-sage border-gray-300 rounded"
                            />
                            <label for="account-projects-single-invoice" class="ml-2 block text-sm text-gray-900">
                              Combine all projects into a single invoice
                            </label>
                          </div>
                        </div>
                      </div>

                      <!-- Budget Hours -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-3">
                        <div>
                          <label for="account-budget-hours" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Budget Hours</label>
                        </div>
                        <div class="sm:col-span-2">
                          <input 
                            type="number" 
                            name="account-budget-hours" 
                            id="account-budget-hours" 
                            v-model="account.budget_hours"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6" 
                          />
                        </div>
                      </div>

                      <!-- Budget Dollars -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-3">
                        <div>
                          <label for="account-budget-dollars" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Budget Dollars</label>
                        </div>
                        <div class="sm:col-span-2">
                          <input 
                            type="number" 
                            name="account-budget-dollars" 
                            id="account-budget-dollars" 
                            v-model="account.budget_dollars"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6" 
                          />
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
                      <!-- Add delete button only when editing existing account -->
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

const props = defineProps({
  isOpen: {
    type: Boolean,
    default: false
  },
  accountData: {
    type: Object,
    default: () => ({})
  }
});

const emit = defineEmits(['close', 'save', 'delete']);

// Handle close events
const handleClose = () => {
  emit('close');
};

// Handle delete functionality
const handleDelete = () => {
  if (window.confirm('Are you sure you want to delete this account? This action cannot be undone and will also remove all projects associated with this account.')) {
    emit('delete', props.accountData);
  }
};

// Determine if editing or creating new
const isEditing = computed(() => !!(props.accountData?.id || props.accountData?.ID));

// Initialize account with default values or provided data
const account = ref({
  id: props.accountData?.id || null,
  ID: props.accountData?.ID || null,
  name: props.accountData?.name || '',
  type: props.accountData?.type || 'ACCOUNT_TYPE_CLIENT',
  legal_name: props.accountData?.legal_name || '',
  email: props.accountData?.email || '',
  address: props.accountData?.address || '',
  website: props.accountData?.website || '',
  billing_frequency: props.accountData?.billing_frequency || 'BILLING_TYPE_MONTHLY',
  budget_hours: props.accountData?.budget_hours || 0,
  budget_dollars: props.accountData?.budget_dollars || 0,
  projects_single_invoice: props.accountData?.projects_single_invoice || false,
  logo_asset: props.accountData?.logo_asset || null
});

// Logo file handling
const logoFile = ref(null);

const handleLogoChange = (event) => {
  const file = event.target.files[0];
  if (file) {
    // Validate file size (10MB)
    if (file.size > 10 * 1024 * 1024) {
      alert('Logo file must be less than 10MB');
      event.target.value = '';
      return;
    }
    
    // Validate file type
    const validTypes = ['image/png', 'image/jpeg', 'image/svg+xml'];
    if (!validTypes.includes(file.type)) {
      alert('Logo must be PNG, JPG, or SVG');
      event.target.value = '';
      return;
    }
    
    logoFile.value = file;
  }
};

// Update account data when accountData prop changes
watch(() => props.accountData, (newVal) => {
  if (newVal) {
    account.value = {
      id: newVal.id || null,
      ID: newVal.ID || null,
      name: newVal.name || '',
      type: newVal.type || 'ACCOUNT_TYPE_CLIENT',
      legal_name: newVal.legal_name || '',
      email: newVal.email || '',
      address: newVal.address || '',
      website: newVal.website || '',
      billing_frequency: newVal.billing_frequency || 'BILLING_TYPE_MONTHLY',
      budget_hours: newVal.budget_hours || 0,
      budget_dollars: newVal.budget_dollars || 0,
      projects_single_invoice: newVal.projects_single_invoice || false,
      logo_asset: newVal.logo_asset || null
    };
    // Reset logo file when switching accounts
    logoFile.value = null;
  }
}, { deep: true });

// Handle form submission
const handleSubmit = () => {
  // Validate form
  if (!account.value.name || !account.value.type) {
    alert('Please fill in all required fields');
    return;
  }

  console.log('AccountDrawer: Submitting form with account:', account.value);
  console.log('AccountDrawer: Logo file:', logoFile.value);

  // Emit save event with account data and logo file
  emit('save', { accountData: account.value, logoFile: logoFile.value });
};
</script> 