<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
    <!-- Header -->
    <div class="mb-4 flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-gray-900">Recurring Compensation</h1>
        <p class="mt-1 text-xs text-gray-500">
          Manage fixed monthly compensation (base salary, stipends) for staff
        </p>
      </div>
      <div class="flex gap-2">
        <button
          @click="openCreateModal"
          class="px-3 py-1.5 bg-sage text-white text-xs rounded hover:bg-sage-dark"
        >
          <i class="fas fa-plus mr-1"></i>
          Add Entry
        </button>
        <button
          @click="syncEmployees"
          :disabled="isSyncing"
          class="px-3 py-1.5 bg-sky-600 text-white text-xs rounded hover:bg-sky-700 disabled:bg-gray-300"
        >
          <i class="fas fa-sync mr-1" :class="{ 'fa-spin': isSyncing }"></i>
          {{ isSyncing ? 'Syncing...' : 'Sync Staff' }}
        </button>
        <button
          @click="generateNow"
          :disabled="isGenerating"
          class="px-3 py-1.5 bg-sage text-white text-xs rounded hover:bg-sage-dark disabled:bg-gray-300"
        >
          <i class="fas fa-play mr-1"></i>
          {{ isGenerating ? 'Generating...' : 'Generate Payroll' }}
        </button>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-3 mb-4">
      <div class="bg-white rounded-lg shadow p-3">
        <div class="flex items-center">
          <div class="flex-shrink-0 bg-sage-pale rounded-md p-2">
            <i class="fas fa-users text-sage text-sm"></i>
          </div>
          <div class="ml-3">
            <p class="text-2xs font-medium text-gray-500">Active Entries</p>
            <p class="text-lg font-semibold text-gray-900">{{ activeCount }}</p>
          </div>
        </div>
      </div>
      
      <div class="bg-white rounded-lg shadow p-3">
        <div class="flex items-center">
          <div class="flex-shrink-0 bg-blue-50 rounded-md p-2">
            <i class="fas fa-dollar-sign text-blue-600 text-sm"></i>
          </div>
          <div class="ml-3">
            <p class="text-2xs font-medium text-gray-500">Monthly Fixed Costs</p>
            <p class="text-lg font-semibold text-gray-900">{{ formatCurrency(monthlyTotal) }}</p>
          </div>
        </div>
      </div>

      <div class="bg-white rounded-lg shadow p-3">
        <div class="flex items-center">
          <div class="flex-shrink-0 bg-sage-pale rounded-md p-2">
            <i class="fas fa-calendar text-sage text-sm"></i>
          </div>
          <div class="ml-3">
            <p class="text-2xs font-medium text-gray-500">Last Generated</p>
            <p class="text-xs font-semibold text-gray-900">{{ lastGenerated || 'Never' }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white shadow rounded-lg overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-4 py-2 text-left text-2xs font-medium text-gray-500 uppercase tracking-wider">
              Employee
            </th>
            <th class="px-4 py-2 text-left text-2xs font-medium text-gray-500 uppercase tracking-wider">
              Type
            </th>
            <th class="px-4 py-2 text-right text-2xs font-medium text-gray-500 uppercase tracking-wider">
              Amount
            </th>
            <th class="px-4 py-2 text-left text-2xs font-medium text-gray-500 uppercase tracking-wider">
              Frequency
            </th>
            <th class="px-4 py-2 text-center text-2xs font-medium text-gray-500 uppercase tracking-wider">
              Status
            </th>
            <th class="px-4 py-2 text-center text-2xs font-medium text-gray-500 uppercase tracking-wider">
              Last Generated
            </th>
            <th class="px-4 py-2 text-center text-2xs font-medium text-gray-500 uppercase tracking-wider">
              Actions
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-if="entries.length === 0">
            <td colspan="7" class="px-4 py-6 text-center text-xs text-gray-500">
              <i class="fas fa-inbox text-2xl text-gray-300 mb-2"></i>
              <p>No recurring entries found</p>
              <p class="text-2xs mt-1">Click "Sync Staff" to create entries for salaried employees</p>
            </td>
          </tr>
          <tr v-for="entry in entries" :key="entry.ID" class="hover:bg-gray-50">
            <td class="px-4 py-2 whitespace-nowrap">
              <div class="flex items-center">
                <StaffAvatar :employee="entry.employee" size="sm" />
                <div class="ml-3">
                  <div class="text-xs font-medium text-gray-900">
                    {{ entry.employee?.first_name }} {{ entry.employee?.last_name }}
                  </div>
                  <div class="text-2xs text-gray-500">
                    {{ entry.employee?.title }}
                  </div>
                </div>
              </div>
            </td>
            <td class="px-4 py-2 whitespace-nowrap text-xs text-gray-900">
              {{ formatType(entry.type) }}
            </td>
            <td class="px-4 py-2 whitespace-nowrap text-xs text-right font-mono text-gray-900">
              {{ formatCurrency(entry.amount) }}
            </td>
            <td class="px-4 py-2 whitespace-nowrap text-xs text-gray-700">
              {{ entry.frequency }}
            </td>
            <td class="px-4 py-2 whitespace-nowrap text-center">
              <span 
                :class="[
                  'px-2 py-0.5 inline-flex text-2xs leading-4 font-semibold rounded-full',
                  entry.is_active ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                ]"
              >
                {{ entry.is_active ? 'Active' : 'Inactive' }}
              </span>
            </td>
            <td class="px-4 py-2 whitespace-nowrap text-xs text-center text-gray-500">
              {{ entry.last_generated_for ? formatDate(entry.last_generated_for) : '-' }}
            </td>
            <td class="px-4 py-2 whitespace-nowrap text-center text-xs">
              <button
                @click="toggleActive(entry)"
                class="text-sky-600 hover:text-sky-900 mr-2"
                :title="entry.is_active ? 'Deactivate' : 'Activate'"
              >
                <i :class="['fas', 'text-2xs', entry.is_active ? 'fa-pause' : 'fa-play']"></i>
              </button>
              <button
                @click="deleteEntry(entry)"
                class="text-red-600 hover:text-red-900"
                title="Delete"
              >
                <i class="fas fa-trash-alt text-2xs"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create/Edit Modal -->
    <div
      v-if="showModal"
      @click="closeModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-gray-500/75 backdrop-blur-sm"
    >
      <div
        @click.stop
        class="bg-white rounded-lg shadow-xl w-full max-w-lg mx-4 max-h-[90vh] flex flex-col"
      >
        <div class="bg-gradient-to-r from-sage to-sage-dark px-4 py-2 rounded-t-lg flex items-center justify-between">
          <h3 class="text-xs font-semibold text-white">
            {{ isEditing ? 'Edit' : 'Add' }} Recurring Entry
          </h3>
          <button @click="closeModal" class="text-white hover:text-gray-200">
            <i class="fas fa-times text-xs"></i>
          </button>
        </div>
        <div class="p-3 overflow-y-auto flex-1">
          <div v-if="modalError" class="mb-2 bg-red-50 border border-red-200 text-red-700 px-2 py-1.5 rounded text-2xs">
            {{ modalError }}
          </div>

          <div class="space-y-2">
            <!-- Employee -->
            <div>
              <label class="block text-2xs font-medium text-gray-700 mb-1">
                Employee <span class="text-red-500">*</span>
              </label>
              <select
                v-model="formData.employee_id"
                required
                class="w-full px-3 py-1.5 text-xs border border-gray-300 rounded focus:ring-sage focus:border-sage"
              >
                <option value="">Select employee...</option>
                <option v-for="emp in allEmployees" :key="emp.ID" :value="emp.ID">
                  {{ emp.first_name }} {{ emp.last_name }}
                </option>
              </select>
            </div>

            <!-- Type -->
            <div>
              <label class="block text-2xs font-medium text-gray-700 mb-1">
                Type <span class="text-red-500">*</span>
              </label>
              <select
                v-model="formData.type"
                required
                class="w-full px-3 py-1.5 text-xs border border-gray-300 rounded focus:ring-sage focus:border-sage"
              >
                <option value="base_salary">Base Salary</option>
                <option value="bonus">Bonus</option>
                <option value="stipend">Stipend</option>
                <option value="other">Other</option>
              </select>
            </div>

            <!-- Description -->
            <div>
              <label class="block text-2xs font-medium text-gray-700 mb-1">
                Description <span class="text-red-500">*</span>
              </label>
              <input
                v-model="formData.description"
                type="text"
                required
                placeholder="e.g., Monthly Base Salary"
                class="w-full px-3 py-1.5 text-xs border border-gray-300 rounded focus:ring-sage focus:border-sage"
              />
            </div>

            <!-- Amount -->
            <div>
              <label class="block text-2xs font-medium text-gray-700 mb-1">
                Monthly Amount <span class="text-red-500">*</span>
              </label>
              <div class="relative">
                <span class="absolute left-3 top-1.5 text-gray-500 text-xs">$</span>
                <input
                  v-model.number="formData.amount"
                  type="number"
                  step="0.01"
                  required
                  class="w-full pl-8 pr-3 py-1.5 text-xs border border-gray-300 rounded focus:ring-sage focus:border-sage"
                />
              </div>
            </div>

            <!-- Start Date -->
            <div>
              <label class="block text-2xs font-medium text-gray-700 mb-1">
                Start Date <span class="text-red-500">*</span>
              </label>
              <input
                v-model="formData.start_date"
                type="date"
                required
                class="w-full px-3 py-1.5 text-xs border border-gray-300 rounded focus:ring-sage focus:border-sage"
              />
            </div>

            <!-- End Date -->
            <div>
              <label class="block text-2xs font-medium text-gray-700 mb-1">
                End Date (Optional)
              </label>
              <input
                v-model="formData.end_date"
                type="date"
                class="w-full px-3 py-1.5 text-xs border border-gray-300 rounded focus:ring-sage focus:border-sage"
              />
            </div>
          </div>
        </div>

        <!-- Fixed Footer -->
        <div class="border-t border-gray-200 px-3 py-2 bg-gray-50 rounded-b-lg flex justify-end gap-2">
          <button
            @click="closeModal"
            class="px-3 py-1.5 text-xs text-gray-700 border border-gray-300 rounded hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            @click="saveEntry"
            :disabled="isSubmitting"
            class="px-3 py-1.5 text-xs bg-sage text-white rounded hover:bg-sage-dark disabled:bg-gray-300"
          >
            {{ isSubmitting ? 'Saving...' : (isEditing ? 'Update' : 'Create') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import {
  getRecurringEntries,
  createRecurringEntry,
  updateRecurringEntry,
  deleteRecurringEntry,
  syncAllEmployees,
  generateRecurringEntries,
  type RecurringEntry
} from '../../api/recurringEntries';
import { fetchStaff } from '../../api/staff';
import StaffAvatar from '../../components/StaffAvatar.vue';

const entries = ref<RecurringEntry[]>([]);
const allEmployees = ref<any[]>([]);
const isSyncing = ref(false);
const isGenerating = ref(false);
const showModal = ref(false);
const isEditing = ref(false);
const isSubmitting = ref(false);
const modalError = ref<string | null>(null);

const formData = ref({
  id: 0,
  employee_id: '',
  type: 'base_salary',
  description: '',
  amount: 0,
  start_date: new Date().toISOString().split('T')[0],
  end_date: '',
});

const activeCount = computed(() => entries.value.filter(e => e.is_active).length);
const monthlyTotal = computed(() => 
  entries.value
    .filter(e => e.is_active)
    .reduce((sum, e) => sum + e.amount, 0)
);
const lastGenerated = computed(() => {
  const dates = entries.value
    .map(e => e.last_generated_for)
    .filter(d => d)
    .sort()
    .reverse();
  return dates.length > 0 ? formatDate(dates[0]!) : null;
});

async function loadEntries() {
  try {
    entries.value = await getRecurringEntries();
  } catch (error) {
    console.error('Failed to load recurring entries:', error);
  }
}

async function loadEmployees() {
  try {
    allEmployees.value = await fetchStaff();
  } catch (error) {
    console.error('Failed to load employees:', error);
  }
}

async function syncEmployees() {
  isSyncing.value = true;
  try {
    const result = await syncAllEmployees();
    console.log('Sync result:', result);
    await loadEntries();
  } catch (error) {
    console.error('Failed to sync employees:', error);
  } finally {
    isSyncing.value = false;
  }
}

async function generateNow() {
  isGenerating.value = true;
  try {
    const result = await generateRecurringEntries();
    console.log('Generate result:', result);
    console.log(`Period: ${result.period}`);
    console.log(`Line items added: ${result.line_items_added} (${result.line_items_before} → ${result.line_items_after})`);
    console.log(`GL entries: ${result.gl_entries}`);
    await loadEntries();
  } catch (error) {
    console.error('Failed to generate entries:', error);
  } finally {
    isGenerating.value = false;
  }
}

async function toggleActive(entry: RecurringEntry) {
  try {
    await updateRecurringEntry(entry.ID, { is_active: !entry.is_active });
    await loadEntries();
  } catch (error) {
    console.error('Failed to toggle entry:', error);
  }
}

async function deleteEntry(entry: RecurringEntry) {
  try {
    await deleteRecurringEntry(entry.ID);
    await loadEntries();
  } catch (error) {
    console.error('Failed to delete entry:', error);
  }
}

function formatType(type: string): string {
  return type
    .split('_')
    .map(w => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ');
}

function formatCurrency(cents: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(cents / 100);
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short' });
}

function openCreateModal() {
  isEditing.value = false;
  formData.value = {
    id: 0,
    employee_id: '',
    type: 'base_salary',
    description: '',
    amount: 0,
    start_date: new Date().toISOString().split('T')[0],
    end_date: '',
  };
  modalError.value = null;
  showModal.value = true;
}

function closeModal() {
  showModal.value = false;
  modalError.value = null;
}

async function saveEntry() {
  isSubmitting.value = true;
  modalError.value = null;

  try {
    const data = {
      employee_id: parseInt(formData.value.employee_id),
      type: formData.value.type,
      description: formData.value.description,
      amount: Math.round(formData.value.amount * 100), // Convert dollars to cents
      frequency: 'monthly',
      start_date: formData.value.start_date,
      end_date: formData.value.end_date || undefined,
      is_active: true,
    };

    if (isEditing.value) {
      await updateRecurringEntry(formData.value.id, data);
    } else {
      await createRecurringEntry(data);
    }

    await loadEntries();
    closeModal();
  } catch (error: any) {
    console.error('Failed to save entry:', error);
    modalError.value = error.response?.data?.error || error.message || 'Failed to save entry';
  } finally {
    isSubmitting.value = false;
  }
}

onMounted(() => {
  loadEntries();
  loadEmployees();
});
</script>

