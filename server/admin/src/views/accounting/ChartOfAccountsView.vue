<template>
  <div class="p-2 bg-white min-h-screen">
    <!-- Header -->
    <div class="mb-1 pb-1 border-b border-gray-900 flex justify-between items-center">
      <div>
        <h1 class="text-sm font-bold text-gray-900 uppercase">Chart of Accounts</h1>
      </div>
      <div class="flex gap-1">
        <button
          @click="openCreateModal"
          class="px-2 py-0.5 text-2xs font-medium text-white bg-sky-700 hover:bg-sky-800 rounded"
        >
          + Account
        </button>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-gray-50 border border-gray-300 rounded px-1.5 py-0.5 mb-1 flex gap-2 items-center">
      <div class="flex-1">
        <select
          v-model="filters.accountType"
          @change="fetchData"
          class="block w-full rounded border-gray-300 text-2xs py-0.5"
        >
          <option value="">All Types</option>
          <option value="ASSET">Assets</option>
          <option value="LIABILITY">Liabilities</option>
          <option value="EQUITY">Equity</option>
          <option value="REVENUE">Revenue</option>
          <option value="EXPENSE">Expenses</option>
        </select>
      </div>
      <div class="flex-1">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search..."
          class="block w-full rounded border-gray-300 text-2xs py-0.5"
        />
      </div>
      <div class="flex items-center gap-1">
        <input
          id="activeOnly"
          v-model="filters.activeOnly"
          @change="fetchData"
          type="checkbox"
          class="h-3 w-3 rounded border-gray-300 text-sky-700"
        />
        <label for="activeOnly" class="text-2xs text-gray-700 whitespace-nowrap">Active only</label>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="text-center py-2">
      <p class="text-2xs text-gray-500">Loading...</p>
    </div>

    <!-- Accounts List -->
    <div v-else class="space-y-2">
      <div v-for="type in accountTypes" :key="type" class="border border-gray-900">
        <div class="bg-gray-900 text-white px-3 py-1">
          <h2 class="text-xs font-bold uppercase tracking-wide">{{ type }}</h2>
        </div>
        
        <table class="w-full text-xs">
          <thead class="bg-gray-100 border-b border-gray-300">
            <tr>
              <th class="px-3 py-1 text-left font-semibold text-gray-700 uppercase w-8"></th>
              <th class="px-3 py-1 text-left font-semibold text-gray-700 uppercase">Code</th>
              <th class="px-3 py-1 text-left font-semibold text-gray-700 uppercase">Name</th>
              <th class="px-3 py-1 text-left font-semibold text-gray-700 uppercase">Description</th>
              <th class="px-3 py-1 text-center font-semibold text-gray-700 uppercase w-20">Actions</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="account in getAccountsByType(type)" :key="account.ID">
              <!-- Account Row -->
              <tr class="border-b border-gray-100 hover:bg-gray-50 cursor-pointer" @click="toggleAccount(account.account_code)">
                <td class="px-3 py-1 text-gray-400">
                  <i 
                    v-if="getSubaccountsByAccount(account.account_code).length > 0"
                    :class="expandedAccounts.has(account.account_code) ? 'fas fa-chevron-down' : 'fas fa-chevron-right'" 
                    class="text-xs"
                  ></i>
                </td>
                <td class="px-3 py-1 font-mono text-gray-900">
                  {{ account.account_code }}
                  <span v-if="account.is_system_defined" class="ml-1 text-gray-400">(Sys)</span>
                </td>
                <td class="px-3 py-1 font-medium text-gray-900">{{ account.account_name }}</td>
                <td class="px-3 py-1 text-gray-600">{{ account.description }}</td>
                <td class="px-3 py-1 text-center" @click.stop>
                  <button
                    @click="openCreateSubaccountModal(account.account_code)"
                    class="text-sage hover:text-sage-dark transition-colors"
                    title="Add subaccount"
                  >
                    <i class="fas fa-plus text-xs"></i>
                  </button>
                  <button
                    v-if="!account.is_system_defined"
                    @click="openEditModal(account)"
                    class="ml-2 text-sky-700 hover:text-sky-900 transition-colors"
                    title="Edit account"
                  >
                    <i class="fas fa-edit text-xs"></i>
                  </button>
                  <button
                    v-if="!account.is_system_defined"
                    @click="deactivateAccount(account.account_code)"
                    class="ml-2 text-red hover:text-red-700 transition-colors"
                    title="Deactivate account"
                  >
                    <i class="fas fa-trash-alt text-xs"></i>
                  </button>
                </td>
              </tr>
              
              <!-- Subaccounts (Expanded) -->
              <tr v-if="expandedAccounts.has(account.account_code)" v-for="sub in getSubaccountsByAccount(account.account_code)" :key="sub.ID" class="bg-gray-50 border-b border-gray-100">
                <td class="px-3 py-1"></td>
                <td class="px-3 py-1 pl-8 font-mono text-gray-700">
                  {{ sub.code }}
                </td>
                <td class="px-3 py-1 text-gray-600">{{ sub.name }}</td>
                <td class="px-3 py-1 text-gray-600">{{ sub.type }}</td>
                <td class="px-3 py-1 text-center">
                  <button
                    @click="openEditSubaccountModal(sub)"
                    class="text-sky-700 hover:text-sky-900 transition-colors"
                    title="Edit subaccount"
                  >
                    <i class="fas fa-edit text-xs"></i>
                  </button>
                  <button
                    @click="deactivateSubaccount(sub.code)"
                    class="ml-2 text-red hover:text-red-700 transition-colors"
                    title="Deactivate subaccount"
                  >
                    <i class="fas fa-trash-alt text-xs"></i>
                  </button>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
        
        <div v-if="getAccountsByType(type).length === 0" class="px-3 py-2 text-center text-xs text-gray-500">
          No {{ type.toLowerCase() }} accounts found
        </div>
      </div>
    </div>

    <!-- Account Create/Edit Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-gray-500/75 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-md w-full mx-2">
        <div class="bg-sage px-3 py-2 rounded-t-lg flex justify-between items-center">
          <h3 class="text-xs font-semibold text-white">
            {{ isEditing ? 'Edit Account' : 'Create Account' }}
          </h3>
          <button @click="closeModal" class="text-white hover:text-gray-200">
            <span class="text-lg">×</span>
          </button>
        </div>

        <form @submit.prevent="submitForm" class="p-3">
          <div v-if="modalError" class="mb-2 p-2 bg-red-50 border border-red-200 rounded">
            <p class="text-2xs text-red-600">{{ modalError }}</p>
          </div>

          <div class="space-y-2">
            <div>
              <label class="block text-2xs font-medium text-gray-700 mb-0.5">Account Code*</label>
              <input
                v-model="formData.account_code"
                :disabled="isEditing"
                required
                type="text"
                class="block w-full rounded border-gray-300 text-2xs py-0.5 disabled:bg-gray-100"
              />
            </div>

            <div>
              <label class="block text-2xs font-medium text-gray-700 mb-0.5">Account Name*</label>
              <input
                v-model="formData.account_name"
                required
                type="text"
                class="block w-full rounded border-gray-300 text-2xs py-0.5"
              />
            </div>

            <div>
              <label class="block text-2xs font-medium text-gray-700 mb-0.5">Account Type*</label>
              <select
                v-model="formData.account_type"
                :disabled="isEditing"
                required
                class="block w-full rounded border-gray-300 text-2xs py-0.5 disabled:bg-gray-100"
              >
                <option value="ASSET">Asset</option>
                <option value="LIABILITY">Liability</option>
                <option value="EQUITY">Equity</option>
                <option value="REVENUE">Revenue</option>
                <option value="EXPENSE">Expense</option>
              </select>
            </div>

            <div>
              <label class="block text-2xs font-medium text-gray-700 mb-0.5">Description</label>
              <textarea
                v-model="formData.description"
                rows="2"
                class="block w-full rounded border-gray-300 text-2xs py-0.5"
              ></textarea>
            </div>

            <div class="flex items-center gap-1.5">
              <input
                id="is_active_account"
                v-model="formData.is_active"
                type="checkbox"
                class="h-3 w-3 rounded border-gray-300 text-sky-700"
              />
              <label for="is_active_account" class="text-2xs text-gray-700">Active</label>
            </div>
          </div>

          <div class="flex justify-end gap-1.5 mt-2.5 pt-2 border-t border-gray-200">
            <button
              type="button"
              @click="closeModal"
              class="px-2.5 py-1 text-2xs font-medium text-gray-700 bg-white border border-gray-300 rounded hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="isSubmitting"
              class="px-2.5 py-1 text-2xs font-medium text-white bg-sky-700 rounded hover:bg-sky-800 disabled:bg-gray-400"
            >
              {{ isSubmitting ? 'Saving...' : (isEditing ? 'Update' : 'Create') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Subaccount Create/Edit Modal -->
    <div v-if="showSubaccountModal" class="fixed inset-0 bg-gray-500/75 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-md w-full mx-2">
        <div class="bg-sage px-3 py-2 rounded-t-lg flex justify-between items-center">
          <h3 class="text-xs font-semibold text-white">
            {{ isEditingSubaccount ? 'Edit Subaccount' : 'Create Subaccount' }}
          </h3>
          <button @click="closeSubaccountModal" class="text-white hover:text-gray-200">
            <span class="text-lg">×</span>
          </button>
        </div>

        <form @submit.prevent="submitSubaccountForm" class="p-3">
          <div v-if="modalError" class="mb-2 p-2 bg-red-50 border border-red-200 rounded">
            <p class="text-2xs text-red-600">{{ modalError }}</p>
          </div>

          <div class="space-y-2">
            <div>
              <label class="block text-2xs font-medium text-gray-700 mb-0.5">Parent Account</label>
              <input
                :value="subaccountFormData.account_code"
                disabled
                type="text"
                class="block w-full rounded border-gray-300 text-2xs py-0.5 bg-gray-100"
              />
            </div>

            <div>
              <label class="block text-2xs font-medium text-gray-700 mb-0.5">Subaccount Code*</label>
              <input
                v-model="subaccountFormData.code"
                :disabled="isEditingSubaccount"
                required
                type="text"
                class="block w-full rounded border-gray-300 text-2xs py-0.5 disabled:bg-gray-100"
              />
            </div>

            <div>
              <label class="block text-2xs font-medium text-gray-700 mb-0.5">Subaccount Name*</label>
              <input
                v-model="subaccountFormData.name"
                required
                type="text"
                class="block w-full rounded border-gray-300 text-2xs py-0.5"
              />
            </div>

            <div>
              <label class="block text-2xs font-medium text-gray-700 mb-0.5">Type</label>
              <select
                v-model="subaccountFormData.type"
                class="block w-full rounded border-gray-300 text-2xs py-0.5"
              >
                <option value="CUSTOM">Custom</option>
                <option value="CLIENT">Client</option>
                <option value="VENDOR">Vendor</option>
                <option value="EMPLOYEE">Employee</option>
                <option value="PROJECT">Project</option>
              </select>
            </div>

            <div class="flex items-center gap-1.5">
              <input
                id="is_active_subaccount"
                v-model="subaccountFormData.is_active"
                type="checkbox"
                class="h-3 w-3 rounded border-gray-300 text-sky-700"
              />
              <label for="is_active_subaccount" class="text-2xs text-gray-700">Active</label>
            </div>
          </div>

          <div class="flex justify-end gap-1.5 mt-2.5 pt-2 border-t border-gray-200">
            <button
              type="button"
              @click="closeSubaccountModal"
              class="px-2.5 py-1 text-2xs font-medium text-gray-700 bg-white border border-gray-300 rounded hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="isSubmitting"
              class="px-2.5 py-1 text-2xs font-medium text-white bg-sky-700 rounded hover:bg-sky-800 disabled:bg-gray-400"
            >
              {{ isSubmitting ? 'Saving...' : (isEditingSubaccount ? 'Update' : 'Create') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import type { ChartOfAccount, ChartOfAccountCreate, ChartOfAccountUpdate } from '../../types/ChartOfAccount';
import type { Subaccount, SubaccountCreate, SubaccountUpdate } from '../../types/Subaccount';
import {
  getChartOfAccounts,
  createChartOfAccount,
  updateChartOfAccount,
  deactivateChartOfAccount,
  seedSystemAccounts as seedAccountsAPI,
} from '../../api/chartOfAccounts';
import {
  getSubaccounts,
  createSubaccount,
  updateSubaccount,
  deactivateSubaccount as deactivateSubaccountAPI,
} from '../../api/subaccounts';

const accounts = ref<ChartOfAccount[]>([]);
const subaccounts = ref<Subaccount[]>([]);
const expandedAccounts = ref<Set<string>>(new Set());
const isLoading = ref(false);
const searchQuery = ref('');

const filters = ref({
  accountType: '',
  activeOnly: true,
});

// Account Modal state
const showModal = ref(false);
const isEditing = ref(false);
const isSubmitting = ref(false);
const modalError = ref<string | null>(null);
const editingAccountCode = ref<string | null>(null);

const formData = ref<ChartOfAccountCreate>({
  account_code: '',
  account_name: '',
  account_type: 'EXPENSE',
  description: '',
  is_active: true,
});

// Subaccount Modal state
const showSubaccountModal = ref(false);
const isEditingSubaccount = ref(false);
const editingSubaccountCode = ref<string | null>(null);

const subaccountFormData = ref<SubaccountCreate>({
  code: '',
  name: '',
  account_code: '',
  type: 'CUSTOM',
  is_active: true,
});

const accountTypes = computed(() => {
  const types = ['ASSET', 'LIABILITY', 'EQUITY', 'REVENUE', 'EXPENSE'];
  if (filters.value.accountType) {
    return [filters.value.accountType];
  }
  return types;
});

const filteredAccounts = computed(() => {
  let filtered = accounts.value;

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(
      (acc) =>
        acc.account_code.toLowerCase().includes(query) ||
        acc.account_name.toLowerCase().includes(query) ||
        (acc.description && acc.description.toLowerCase().includes(query))
    );
  }

  return filtered;
});

function getAccountsByType(type: string) {
  return filteredAccounts.value.filter((acc) => acc.account_type === type);
}

function getSubaccountsByAccount(accountCode: string) {
  return subaccounts.value.filter((sub) => sub.account_code === accountCode);
}

async function toggleAccount(accountCode: string) {
  if (expandedAccounts.value.has(accountCode)) {
    expandedAccounts.value.delete(accountCode);
  } else {
    expandedAccounts.value.add(accountCode);
    
    // Lazy load subaccounts for this account if not already loaded
    const hasSubaccounts = subaccounts.value.some(sub => sub.account_code === accountCode);
    if (!hasSubaccounts) {
      try {
        const accountSubaccounts = await getSubaccounts({
          account_code: accountCode,
          active_only: filters.value.activeOnly,
        });
        subaccounts.value = [...subaccounts.value, ...accountSubaccounts];
      } catch (error) {
        console.error(`Failed to fetch subaccounts for ${accountCode}:`, error);
      }
    }
  }
}

async function fetchData() {
  isLoading.value = true;
  try {
    accounts.value = await getChartOfAccounts({
      account_type: filters.value.accountType || undefined,
      active_only: filters.value.activeOnly,
    });
    
    // Auto-seed system accounts if none exist
    if (accounts.value.length === 0) {
      try {
        await seedAccountsAPI();
        // Reload accounts after seeding
        accounts.value = await getChartOfAccounts({
          account_type: filters.value.accountType || undefined,
          active_only: filters.value.activeOnly,
        });
      } catch (error) {
        console.error('Failed to seed system accounts:', error);
      }
    }

    // Don't fetch all subaccounts upfront - only load when accounts are expanded
    // This dramatically improves performance when there are many subaccounts
    subaccounts.value = [];
  } catch (error) {
    console.error('Failed to fetch data:', error);
  } finally {
    isLoading.value = false;
  }
}

function openCreateModal() {
  isEditing.value = false;
  editingAccountCode.value = null;
  modalError.value = null;
  formData.value = {
    account_code: '',
    account_name: '',
    account_type: 'EXPENSE',
    description: '',
    is_active: true,
  };
  showModal.value = true;
}

function openEditModal(account: ChartOfAccount) {
  isEditing.value = true;
  editingAccountCode.value = account.account_code;
  modalError.value = null;
  formData.value = {
    account_code: account.account_code,
    account_name: account.account_name,
    account_type: account.account_type,
    description: account.description || '',
    is_active: account.is_active,
  };
  showModal.value = true;
}

function closeModal() {
  showModal.value = false;
  isEditing.value = false;
  editingAccountCode.value = null;
  modalError.value = null;
}

async function submitForm() {
  isSubmitting.value = true;
  modalError.value = null;
  
  try {
    if (isEditing.value && editingAccountCode.value) {
      const updateData: ChartOfAccountUpdate = {
        account_name: formData.value.account_name,
        description: formData.value.description,
        is_active: formData.value.is_active,
      };
      await updateChartOfAccount(editingAccountCode.value, updateData);
    } else {
      await createChartOfAccount(formData.value);
    }
    await fetchData();
    closeModal();
  } catch (error: any) {
    console.error('Failed to save account:', error);
    modalError.value = error.response?.data?.error || 'Failed to save account';
  } finally {
    isSubmitting.value = false;
  }
}

function openCreateSubaccountModal(accountCode: string) {
  isEditingSubaccount.value = false;
  editingSubaccountCode.value = null;
  modalError.value = null;
  subaccountFormData.value = {
    code: '',
    name: '',
    account_code: accountCode,
    type: 'CUSTOM',
    is_active: true,
  };
  showSubaccountModal.value = true;
}

function openEditSubaccountModal(subaccount: Subaccount) {
  isEditingSubaccount.value = true;
  editingSubaccountCode.value = subaccount.code;
  modalError.value = null;
  subaccountFormData.value = {
    code: subaccount.code,
    name: subaccount.name,
    account_code: subaccount.account_code,
    type: subaccount.type,
    is_active: subaccount.is_active,
  };
  showSubaccountModal.value = true;
}

function closeSubaccountModal() {
  showSubaccountModal.value = false;
  isEditingSubaccount.value = false;
  editingSubaccountCode.value = null;
  modalError.value = null;
}

async function submitSubaccountForm() {
  isSubmitting.value = true;
  modalError.value = null;
  
  try {
    if (isEditingSubaccount.value && editingSubaccountCode.value) {
      const updateData: SubaccountUpdate = {
        name: subaccountFormData.value.name,
        type: subaccountFormData.value.type,
        is_active: subaccountFormData.value.is_active,
      };
      await updateSubaccount(editingSubaccountCode.value, updateData);
    } else {
      await createSubaccount(subaccountFormData.value);
    }
    await fetchData();
    // Auto-expand the parent account
    expandedAccounts.value.add(subaccountFormData.value.account_code);
    closeSubaccountModal();
  } catch (error: any) {
    console.error('Failed to save subaccount:', error);
    modalError.value = error.response?.data?.error || 'Failed to save subaccount';
  } finally {
    isSubmitting.value = false;
  }
}

async function deactivateAccount(accountCode: string) {
  if (!confirm(`Are you sure you want to deactivate ${accountCode}?`)) {
    return;
  }
  
  try {
    await deactivateChartOfAccount(accountCode);
    await fetchData();
  } catch (error) {
    console.error('Failed to deactivate account:', error);
    alert('Failed to deactivate account');
  }
}

async function deactivateSubaccount(code: string) {
  if (!confirm(`Are you sure you want to deactivate ${code}?`)) {
    return;
  }
  
  try {
    await deactivateSubaccountAPI(code);
    await fetchData();
  } catch (error) {
    console.error('Failed to deactivate subaccount:', error);
    alert('Failed to deactivate subaccount');
  }
}

onMounted(async () => {
  await fetchData();
});
</script>
