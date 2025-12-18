<template>
  <div class="bg-white">
    <header class="border-b border-gray-200">
      <!-- Secondary navigation -->
      <nav class="flex overflow-x-auto py-4">
        <ul role="list" class="flex min-w-full flex-none gap-x-6 px-4 text-sm/6 font-semibold text-gray-600 sm:px-6 lg:px-8">
          <li>
            <button @click="scrollToSection('organization-settings')" class="text-sage cursor-pointer hover:underline">Organization Settings</button>
          </li>
          <li>
            <button @click="scrollToSection('company-settings')" class="text-gray-500 hover:text-sage cursor-pointer hover:underline">Company Settings</button>
          </li>
          <li>
            <button @click="scrollToSection('billing-defaults')" class="text-gray-500 hover:text-sage cursor-pointer hover:underline">Billing Defaults</button>
          </li>
        </ul>
      </nav>
    </header>

    <!-- Settings forms -->
    <div class="divide-y divide-gray-200">
      <!-- Organization Settings -->
      <div id="organization-settings" class="grid max-w-7xl grid-cols-1 gap-x-8 gap-y-10 px-4 py-16 sm:px-6 md:grid-cols-3 lg:px-8">
        <div>
          <h2 class="text-base/7 font-semibold text-gray-900">Organization Settings</h2>
          <p class="mt-1 text-sm/6 text-gray-500">
            Configure your organization's domain, branding, and global settings.
          </p>
        </div>

        <form @submit.prevent="saveTenantSettings" class="md:col-span-2">
          <div class="grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
            <div class="sm:col-span-4">
              <label for="org-name" class="block text-sm/6 font-medium text-gray-900">Organization Name</label>
              <div class="mt-2">
                <input type="text" name="org-name" id="org-name" v-model="tenantForm.name"
                  class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm/6" />
              </div>
            </div>

            <div class="sm:col-span-4">
              <label for="slug" class="block text-sm/6 font-medium text-gray-900">Organization Slug (Subdomain)</label>
              <div class="mt-2">
                <input type="text" name="slug" id="slug" v-model="tenantForm.slug"
                  placeholder="mycompany"
                  class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm/6" />
              </div>
              <p class="mt-2 text-sm text-gray-500">Subdomain identifier (e.g., mycompany.localhost or mycompany.yourdomain.com)</p>
            </div>

            <div class="sm:col-span-4">
              <label for="domain" class="block text-sm/6 font-medium text-gray-900">Email Domain</label>
              <div class="mt-2">
                <input type="text" name="domain" id="domain" v-model="tenantForm.domain"
                  placeholder="example.com"
                  class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm/6" />
              </div>
              <p class="mt-2 text-sm text-gray-500">Domain used for Google OAuth login (e.g., snowpack-data.com)</p>
            </div>

            <div class="sm:col-span-4">
              <label for="billing-email" class="block text-sm/6 font-medium text-gray-900">Billing Contact Email</label>
              <div class="mt-2">
                <input type="email" name="billing-email" id="billing-email" v-model="tenantSettings.billing_email"
                  placeholder="billing@example.com"
                  class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm/6" />
              </div>
            </div>

            <div class="sm:col-span-3">
              <label for="commission-rate" class="block text-sm/6 font-medium text-gray-900">Default Commission Rate (%)</label>
              <div class="mt-2">
                <input type="number" name="commission-rate" id="commission-rate" v-model.number="tenantSettings.default_commission_rate"
                  min="0" max="100" step="0.01"
                  class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm/6" />
              </div>
            </div>
          </div>

          <div class="mt-8 flex">
            <button type="submit"
              class="rounded-md bg-sage px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-sage-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sage">
              Save Organization Settings
            </button>
          </div>
        </form>
      </div>

      <!-- Company Settings -->
      <div id="company-settings" class="grid max-w-7xl grid-cols-1 gap-x-8 gap-y-10 px-4 py-16 sm:px-6 md:grid-cols-3 lg:px-8">
        <div>
          <h2 class="text-base/7 font-semibold text-gray-900">Company Information</h2>
          <p class="mt-1 text-sm/6 text-gray-500">
            Manage your company settings and contact information.
          </p>
        </div>

        <div class="md:col-span-2" v-if="internalAccounts.length > 0">
          <div v-for="account in internalAccounts" :key="account.id" 
               class="overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm hover:shadow-md transition">
            <div class="p-6">
              <div class="flex items-start justify-between mb-4">
                <div class="flex items-start gap-x-3">
                  <div>
                    <div class="flex items-center gap-x-2">
                      <h3 class="text-xl font-semibold text-gray-900">{{ account.name }}</h3>
                      <span class="inline-flex items-center rounded-md bg-sage-50 px-2 py-1 text-xs font-medium text-sage-700 ring-1 ring-inset ring-sage-600/20">
                        Internal
                      </span>
                    </div>
                    <p class="mt-1 text-sm text-gray-500">{{ account.legal_name || 'No legal name provided' }}</p>
                  </div>
                </div>
                <button
                  @click="openAccountDrawer(account)"
                  class="rounded-md bg-white px-2.5 py-1.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                >
                  <i class="fas fa-pencil-alt mr-1"></i> Edit
                </button>
              </div>
              
              <div class="mt-6 space-y-6 border-t border-gray-200 pt-6">
                <div class="grid grid-cols-1 gap-x-4 gap-y-4 sm:grid-cols-2">
                  <!-- Budget Information -->
                  <div v-if="account.budget_dollars || account.budget_hours" class="sm:col-span-2">
                    <div class="text-sm font-medium text-gray-500 mb-2">Budget</div>
                    <div class="flex flex-wrap gap-2">
                      <span v-if="account.budget_dollars" class="inline-flex items-center rounded-md bg-sage-50 px-2 py-1 text-xs font-medium text-sage-700 ring-1 ring-inset ring-sage-600/20">
                        <i class="fas fa-dollar-sign mr-1.5"></i>
                        {{ formatCurrency(account.budget_dollars) }}{{ formatBudgetFrequency(account.billing_frequency) }}
                      </span>
                      <span v-if="account.budget_hours" class="inline-flex items-center rounded-md bg-indigo-50 px-2 py-1 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-600/20">
                        <i class="fas fa-clock mr-1.5"></i>
                        {{ account.budget_hours }} hours{{ formatBudgetFrequency(account.billing_frequency) }}
                      </span>
                    </div>
                  </div>
                  
                  <div>
                    <div class="text-sm font-medium text-gray-500">Billing Frequency</div>
                    <div class="mt-1 text-sm text-gray-900">{{ formatBillingFrequency(account.billing_frequency) }}</div>
                  </div>
                  <div>
                    <div class="text-sm font-medium text-gray-500">Invoice Type</div>
                    <div class="mt-1 text-sm text-gray-900">
                      <span class="inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset"
                        :class="account.projects_single_invoice ? 
                          'bg-green-50 text-green-700 ring-green-600/20' : 
                          'bg-amber-50 text-amber-700 ring-amber-600/20'">
                        {{ account.projects_single_invoice ? 'Combined' : 'Separate' }}
                      </span>
                    </div>
                  </div>
                  <div v-if="account.address" class="sm:col-span-2">
                    <div class="text-sm font-medium text-gray-500">Address</div>
                    <div class="mt-1 text-sm text-gray-900 whitespace-pre-line">{{ account.address }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          
          <div v-if="internalAccounts.length === 0" class="text-center p-8 bg-gray-50 rounded-lg">
            <i class="fas fa-building text-4xl text-gray-300 mb-3"></i>
            <p class="text-gray-600">No internal accounts found.</p>
            <button
              @click="openAccountDrawer()"
              class="mt-4 inline-flex items-center rounded-md bg-sage px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-sage-dark"
            >
              <i class="fas fa-plus mr-1.5"></i> Create internal account
            </button>
          </div>
        </div>
      </div>

      <!-- Billing Defaults Section -->
      <div id="billing-defaults" class="grid max-w-7xl grid-cols-1 gap-x-8 gap-y-10 px-4 py-16 sm:px-6 md:grid-cols-3 lg:px-8">
        <div>
          <h2 class="text-base/7 font-semibold text-gray-900">Billing Defaults</h2>
          <p class="mt-1 text-sm/6 text-gray-500">
            Configure default billing settings for new projects and clients.
          </p>
        </div>

        <form class="md:col-span-2">
          <div class="grid grid-cols-1 gap-x-6 gap-y-8 sm:max-w-xl sm:grid-cols-6">
            <div class="col-span-full">
              <label for="default-billing-frequency" class="block text-sm/6 font-medium text-gray-900">Default Billing Frequency</label>
              <div class="mt-2">
                <select id="default-billing-frequency" class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm/6">
                  <option value="BILLING_TYPE_WEEKLY">Weekly</option>
                  <option value="BILLING_TYPE_BIWEEKLY">Bi-Weekly</option>
                  <option value="BILLING_TYPE_MONTHLY">Monthly</option>
                  <option value="BILLING_TYPE_BIMONTHLY">Bi-Monthly</option>
                  <option value="BILLING_TYPE_PROJECT">Project-Based</option>
                </select>
              </div>
            </div>

            <div class="col-span-full">
              <label for="default-invoice-type" class="block text-sm/6 font-medium text-gray-900">Default Invoice Type</label>
              <div class="mt-2">
                <select id="default-invoice-type" class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm/6">
                  <option value="true">Combined (all projects in one invoice)</option>
                  <option value="false">Separate (one invoice per project)</option>
                </select>
              </div>
            </div>
          </div>

          <div class="mt-8 flex">
            <button type="submit" class="rounded-md bg-sage px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-sage-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sage">
              Save defaults
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>

  <!-- Account Drawer -->
  <AccountDrawer
    :is-open="isAccountDrawerOpen"
    :account-data="selectedAccount"
    @close="closeAccountDrawer"
    @save="saveAccount"
    @delete="handleDeleteFromDrawer"
  />
  
  <!-- Delete Confirmation Modal -->
  <ConfirmationModal
    :show="showDeleteModal"
    title="Delete Account"
    message="Are you sure you want to delete this account? This action cannot be undone and will also remove all projects associated with this account."
    @confirm="deleteAccount"
    @cancel="showDeleteModal = false"
  />
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { fetchAccounts, createAccount, updateAccount, deleteAccount as deleteAccountAPI, getTenant, updateTenant } from '../../api';
import AccountDrawer from '../../components/accounts/AccountDrawer.vue';
import ConfirmationModal from '../../components/ConfirmationModal.vue';

// State
const accounts = ref([]);
const isAccountDrawerOpen = ref(false);
const selectedAccount = ref(null);
const showDeleteModal = ref(false);
const accountToDelete = ref(null);

// Tenant settings state
const tenantForm = ref({
  name: '',
  slug: '',
  domain: ''
});
const tenantSettings = ref({
  billing_email: '',
  default_commission_rate: 0
});

// Filter accounts to show only internal accounts
const internalAccounts = computed(() => {
  return accounts.value.filter(account => account.type === 'ACCOUNT_TYPE_INTERNAL');
});

// Fetch data
onMounted(async () => {
  try {
    const accountsData = await fetchAccounts();
    accounts.value = accountsData || [];
  } catch (error) {
    console.error('Error loading accounts:', error);
  }
  
  // Load tenant settings
  await loadTenantSettings();
});

// Tenant settings functions
const loadTenantSettings = async () => {
  try {
    const tenant = await getTenant();
    tenantForm.value = {
      name: tenant.name || '',
      slug: tenant.slug || '',
      domain: tenant.domain || ''
    };
    
    // Parse settings JSON
    if (tenant.settings) {
      const settings = typeof tenant.settings === 'string' ? JSON.parse(tenant.settings) : tenant.settings;
      tenantSettings.value = {
        billing_email: settings.billing_email || '',
        default_commission_rate: settings.default_commission_rate || 0
      };
    }
  } catch (error) {
    console.error('Error loading tenant settings:', error);
  }
};

const saveTenantSettings = async () => {
  try {
    const payload = {
      name: tenantForm.value.name,
      slug: tenantForm.value.slug,
      domain: tenantForm.value.domain,
      settings: JSON.stringify(tenantSettings.value)
    };
    
    await updateTenant(payload);
    alert('Organization settings saved successfully');
  } catch (error) {
    console.error('Error saving tenant settings:', error);
    alert('Failed to save organization settings');
  }
};

// Scroll to section function for nav links
const scrollToSection = (id) => {
  const element = document.getElementById(id);
  if (element) {
    element.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }
};

// Helper functions
const formatBillingFrequency = (frequency) => {
  const frequencies = {
    'BILLING_TYPE_WEEKLY': 'Weekly',
    'BILLING_TYPE_BIWEEKLY': 'Bi-Weekly',
    'BILLING_TYPE_MONTHLY': 'Monthly',
    'BILLING_TYPE_BIMONTHLY': 'Bi-Monthly',
    'BILLING_TYPE_PROJECT': 'Project-Based'
  };
  return frequencies[frequency] || frequency;
};

// Format budget frequency suffix based on billing frequency
const formatBudgetFrequency = (frequency) => {
  switch (frequency) {
    case 'BILLING_TYPE_WEEKLY':
      return '/wk';
    case 'BILLING_TYPE_BIWEEKLY':
      return '/2wk';
    case 'BILLING_TYPE_MONTHLY':
      return '/mo';
    case 'BILLING_TYPE_BIMONTHLY':
      return '/2mo';
    case 'BILLING_TYPE_PROJECT':
      return ' total';
    default:
      return '';
  }
};

// Format currency
const formatCurrency = (amount) => {
  if (!amount) return '$0';
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    maximumFractionDigits: 0
  }).format(amount);
};

// Drawer functions
const openAccountDrawer = (account = null) => {
  if (account) {
    // Make sure we include both id and ID properties
    selectedAccount.value = {
      ...account,
      id: account.ID, // Include lowercase 'id' for compatibility with the drawer
      type: 'ACCOUNT_TYPE_INTERNAL' // Force internal type in this view
    };
  } else {
    // For new accounts, set default type to internal
    selectedAccount.value = {
      type: 'ACCOUNT_TYPE_INTERNAL'
    };
  }
  isAccountDrawerOpen.value = true;
};

const closeAccountDrawer = () => {
  isAccountDrawerOpen.value = false;
  selectedAccount.value = null;
};

// Save account
const saveAccount = async (accountData) => {
  try {
    // Ensure we're always saving an internal account
    accountData.type = 'ACCOUNT_TYPE_INTERNAL';
    
    if (accountData.id) {
      // Create a copy of the account data with both id and ID properties
      const accountToUpdate = {
        ...accountData,
        ID: accountData.id // Ensure ID property is set for API compatibility
      };
      await updateAccount(accountToUpdate); // Pass the full object instead of id, accountData
    } else {
      await createAccount(accountData);
    }
    
    // Refresh accounts
    const updatedAccounts = await fetchAccounts();
    accounts.value = updatedAccounts;
    
    // Close drawer
    closeAccountDrawer();
  } catch (error) {
    console.error('Error saving account:', error);
    alert('Failed to save account. Please try again.');
  }
};

// Delete account
const confirmDelete = (account) => {
  accountToDelete.value = account;
  showDeleteModal.value = true;
};

const deleteAccount = async () => {
  if (!accountToDelete.value) return;
  
  try {
    await deleteAccountAPI(accountToDelete.value.id);
    
    // Refresh accounts
    const updatedAccounts = await fetchAccounts();
    accounts.value = updatedAccounts;
    
    // Close modal
    showDeleteModal.value = false;
    accountToDelete.value = null;
  } catch (error) {
    console.error('Error deleting account:', error);
    alert('Failed to delete account. Please try again.');
  }
};

const handleDeleteFromDrawer = (account) => {
  confirmDelete(account);
};
</script>

<style scoped>
.bg-sage-50 {
  background-color: #F0F4F0;
}
.text-sage-700 {
  color: #2E6E32;
}
</style> 