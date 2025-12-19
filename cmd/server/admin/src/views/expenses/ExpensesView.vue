<template>
  <div class="max-w-7xl mx-auto px-3 py-2">
    <!-- Header -->
    <div class="mb-2 flex justify-between items-center">
      <div>
        <h1 class="text-lg font-bold text-gray-900">My Expenses</h1>
      </div>
      <button
        @click="openCreateModal"
        class="px-3 py-1.5 text-xs font-medium text-white bg-blue-600 hover:bg-blue-700 rounded"
      >
        New Expense
      </button>
    </div>

    <!-- Filters -->
    <div class="bg-white shadow rounded p-2 mb-2">
      <div class="flex gap-2 items-end">
        <div class="flex-1">
          <label class="block text-xs font-medium text-gray-700 mb-0.5">Status</label>
          <select
            v-model="filters.status"
            @change="loadExpenses"
            class="block w-full rounded border-gray-300 text-xs py-1"
          >
            <option value="">All Statuses</option>
            <option value="EXPENSE_STATE_DRAFT">Draft</option>
            <option value="EXPENSE_STATE_SUBMITTED">Submitted</option>
            <option value="EXPENSE_STATE_APPROVED">Approved</option>
            <option value="EXPENSE_STATE_REJECTED">Rejected</option>
          </select>
        </div>
        <div class="flex-1">
          <label class="block text-xs font-medium text-gray-700 mb-0.5">Project</label>
          <select
            v-model="filters.projectId"
            @change="loadExpenses"
            class="block w-full rounded border-gray-300 text-xs py-1"
          >
            <option value="">All Projects</option>
            <option v-for="project in projects" :key="project.ID" :value="project.ID">
              {{ project.name }}
            </option>
          </select>
        </div>
        <div class="flex-1">
          <label class="block text-xs font-medium text-gray-700 mb-0.5">Search</label>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search expenses..."
            class="block w-full rounded border-gray-300 text-xs py-1"
          />
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="text-center py-4">
      <p class="text-xs text-gray-500">Loading expenses...</p>
    </div>

    <!-- Expenses Table -->
    <div v-else class="bg-white shadow rounded overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-2 py-1 text-left text-xs font-medium text-gray-500 uppercase">Date</th>
            <th class="px-2 py-1 text-left text-xs font-medium text-gray-500 uppercase">Description</th>
            <th class="px-2 py-1 text-left text-xs font-medium text-gray-500 uppercase">Project</th>
            <th class="px-2 py-1 text-left text-xs font-medium text-gray-500 uppercase">Category</th>
            <th class="px-2 py-1 text-left text-xs font-medium text-gray-500 uppercase">Tags</th>
            <th class="px-2 py-1 text-right text-xs font-medium text-gray-500 uppercase">Amount</th>
            <th class="px-2 py-1 text-center text-xs font-medium text-gray-500 uppercase">Status</th>
            <th class="px-2 py-1 text-center text-xs font-medium text-gray-500 uppercase">Receipt</th>
            <th class="px-2 py-1 text-center text-xs font-medium text-gray-500 uppercase">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-if="filteredExpenses.length === 0">
            <td colspan="9" class="px-2 py-4 text-center text-xs text-gray-500">
              No expenses found. Click "New Expense" to create one.
            </td>
          </tr>
          <tr 
            v-for="expense in filteredExpenses" 
            :key="expense.ID" 
            :class="[
              'transition-colors',
              expense.state === 'EXPENSE_STATE_DRAFT' 
                ? 'hover:bg-sage-pale hover:bg-opacity-30 cursor-pointer' 
                : 'opacity-60 cursor-not-allowed'
            ]"
            @click="expense.state === 'EXPENSE_STATE_DRAFT' ? openEditModal(expense) : null"
          >
            <td class="px-2 py-1 text-xs text-gray-900">{{ formatDate(expense.date) }}</td>
            <td class="px-2 py-1 text-xs text-gray-900">{{ expense.description }}</td>
            <td class="px-2 py-1 text-xs text-gray-600">{{ getProjectName(expense.project_id) }}</td>
            <td class="px-2 py-1 text-xs text-gray-600">
              <span v-if="expense.category" class="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800">
                {{ expense.category.name }}
              </span>
              <span v-else class="text-gray-400">-</span>
            </td>
            <td class="px-2 py-1 text-xs">
              <div v-if="expense.tags && expense.tags.length > 0" class="flex flex-wrap gap-1">
                <span 
                  v-for="tag in expense.tags" 
                  :key="tag.ID"
                  class="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800"
                >
                  {{ tag.name }}
                </span>
              </div>
              <span v-else class="text-gray-400">-</span>
            </td>
            <td class="px-2 py-1 text-right text-xs text-gray-900">{{ formatCurrency(expense.amount) }}</td>
            <td class="px-2 py-1 text-center">
              <span :class="getStatusColor(expense.state)" class="px-1.5 py-0.5 text-xs font-semibold rounded">
                {{ formatStatus(expense.state) }}
              </span>
            </td>
            <td class="px-2 py-1 text-center" @click.stop>
              <button
                v-if="expense.receipt_id"
                @click="viewReceipt(expense.receipt_id)"
                class="inline-flex items-center justify-center w-7 h-7 text-white bg-sage hover:bg-sage-dark rounded shadow-sm transition-colors"
                title="View receipt"
              >
                <i class="fa fa-receipt text-xs"></i>
              </button>
              <span v-else class="text-xs text-gray-400">-</span>
            </td>
            <td class="px-2 py-1 text-center" @click.stop>
              <div class="flex items-center justify-center gap-1">
                <button
                  v-if="expense.state === 'EXPENSE_STATE_DRAFT'"
                  @click="handleSubmitExpense(expense.ID)"
                  class="inline-flex items-center justify-center w-7 h-7 text-white bg-sage hover:bg-sage-dark rounded shadow-sm transition-colors"
                  title="Submit for approval"
                >
                  <i class="fa fa-paper-plane text-xs"></i>
                </button>
                <button
                  v-if="expense.state === 'EXPENSE_STATE_DRAFT'"
                  @click="handleDeleteExpense(expense.ID)"
                  class="inline-flex items-center justify-center w-7 h-7 text-white bg-red-600 hover:bg-red-500 rounded shadow-sm transition-colors"
                  title="Delete expense"
                >
                  <i class="fa fa-trash text-xs"></i>
                </button>
                <span v-if="expense.state !== 'EXPENSE_STATE_DRAFT'" class="text-xs text-gray-400">-</span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="fixed inset-0 z-10 w-screen overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4 text-center sm:p-0">
        <!-- Backdrop -->
        <div class="fixed inset-0 bg-gray-500/75 transition-opacity" @click="closeModal"></div>
        
        <!-- Modal -->
        <div class="relative transform overflow-hidden rounded-lg bg-white text-left shadow-xl transition-all sm:my-6 sm:w-full sm:max-w-md">
          <!-- Header -->
          <div class="bg-sage text-white px-4 py-2 sm:px-4 sm:py-2 flex justify-between items-center">
            <h3 class="text-xs font-semibold leading-6">
              {{ isEditing ? 'Edit Expense' : 'New Expense' }}
            </h3>
            <button
              @click="closeModal"
              type="button"
              class="text-white hover:text-gray-200 focus:outline-none"
            >
              <i class="fa fa-times h-5 w-5"></i>
            </button>
          </div>
          
          <!-- Body -->
          <form @submit.prevent="submitForm" class="bg-white px-3 py-4 sm:p-6">
            <!-- Expense Type Selector -->
            <fieldset class="py-2 border-b border-gray-100">
              <legend class="text-xs font-semibold text-gray-900 mb-2">Expense Type</legend>
              
              <div class="grid grid-cols-2 gap-3">
                <!-- Client Project Card -->
                <label 
                  class="group relative flex flex-col rounded-lg border-2 bg-white p-3 cursor-pointer transition-all
                         has-[:checked]:border-sage has-[:checked]:ring-2 has-[:checked]:ring-sage/20 has-[:checked]:bg-sage-pale
                         hover:border-sage/50"
                >
                  <input 
                    type="radio" 
                    name="expense-type"
                    :value="false"
                    v-model="isInternalExpense" 
                    class="sr-only"
                  />
                  
                  <div class="flex items-start justify-between">
                    <div class="flex-1">
                      <div class="flex items-center gap-2 mb-1">
                        <i class="fas fa-briefcase text-sage text-sm"></i>
                        <span class="text-xs font-semibold text-gray-900">Client Project</span>
                      </div>
                      <span class="text-2xs text-gray-500">Billable expense for a specific client</span>
                    </div>
                    
                    <svg 
                      viewBox="0 0 20 20" 
                      fill="currentColor" 
                      class="invisible group-has-[:checked]:visible h-5 w-5 text-sage flex-shrink-0"
                    >
                      <path d="M10 18a8 8 0 1 0 0-16 8 8 0 0 0 0 16Zm3.857-9.809a.75.75 0 0 0-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 1 0-1.06 1.061l2.5 2.5a.75.75 0 0 0 1.137-.089l4-5.5Z" clip-rule="evenodd" fill-rule="evenodd" />
                    </svg>
                  </div>
                </label>

                <!-- Internal Expense Card -->
                <label 
                  class="group relative flex flex-col rounded-lg border-2 bg-white p-3 cursor-pointer transition-all
                         has-[:checked]:border-orange-500 has-[:checked]:ring-2 has-[:checked]:ring-orange-500/20 has-[:checked]:bg-orange-50
                         hover:border-orange-400"
                >
                  <input 
                    type="radio" 
                    name="expense-type"
                    :value="true"
                    v-model="isInternalExpense" 
                    class="sr-only"
                  />
                  
                  <div class="flex items-start justify-between">
                    <div class="flex-1">
                      <div class="flex items-center gap-2 mb-1">
                        <i class="fas fa-building text-orange-600 text-sm"></i>
                        <span class="text-xs font-semibold text-gray-900">Internal Expense</span>
                      </div>
                      <span class="text-2xs text-gray-500">Company overhead, not billable</span>
                    </div>
                    
                    <svg 
                      viewBox="0 0 20 20" 
                      fill="currentColor" 
                      class="invisible group-has-[:checked]:visible h-5 w-5 text-orange-600 flex-shrink-0"
                    >
                      <path d="M10 18a8 8 0 1 0 0-16 8 8 0 0 0 0 16Zm3.857-9.809a.75.75 0 0 0-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 1 0-1.06 1.061l2.5 2.5a.75.75 0 0 0 1.137-.089l4-5.5Z" clip-rule="evenodd" fill-rule="evenodd" />
                    </svg>
                  </div>
                </label>
              </div>
            </fieldset>

            <!-- Reimbursable Selection (only if internal expense) -->
            <div v-if="isInternalExpense" class="py-1 border-b border-gray-100 mt-1.5">
              <label class="flex items-center">
                <input
                  type="checkbox"
                  v-model="formData.is_reimbursable"
                  class="h-4 w-4 rounded border-gray-300 text-sage focus:ring-sage"
                />
                <span class="ml-2 text-xs font-medium text-gray-900">Reimbursable</span>
              </label>
              <p class="mt-1 text-xs text-gray-500">
                Check if you made this payment with a personal credit card and require reimbursement. 
              </p>
            </div>

            <!-- Project Selection (only if client expense) -->
            <div v-if="!isInternalExpense" class="py-1 border-b border-gray-100 mt-1.5">
              <label class="block text-xs font-medium text-gray-700 mb-0.5">
                Project <span class="text-red-500">*</span>
              </label>
              <select
                v-model="formData.project_id"
                required
                class="bg-white border border-gray-300 text-gray-900 text-xs rounded focus:ring-sage focus:border-sage block w-full p-1"
              >
                <option value="">Select project...</option>
                <option v-for="project in projects" :key="project.ID" :value="project.ID">
                  {{ project.name }}
                </option>
              </select>
            </div>

            <!-- Category -->
            <div class="py-1 border-b border-gray-100 mt-1.5">
              <label class="block text-xs font-medium text-gray-700 mb-0.5">
                Category <span class="text-red-500">*</span>
              </label>
              <select
                v-model="formData.category_id"
                required
                class="bg-white border border-gray-300 text-gray-900 text-xs rounded focus:ring-sage focus:border-sage block w-full p-1"
              >
                <option value="">Select category...</option>
                <option v-for="category in categories" :key="category.ID" :value="category.ID">
                  {{ category.name }}
                </option>
              </select>
            </div>

            <!-- Payment Account removed - will be determined during bank reconciliation -->

            <!-- Tags (Optional) -->
            <div class="py-1 border-b border-gray-100 mt-1.5">
              <label class="block text-xs font-medium text-gray-700 mb-0.5">Tags (optional)</label>
              
              <!-- Selected Tags Display -->
              <div v-if="selectedTagsDisplay.length > 0" class="flex flex-wrap gap-1 mb-1.5">
                <span 
                  v-for="tag in selectedTagsDisplay" 
                  :key="tag.ID"
                  class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 border border-green-200"
                >
                  {{ tag.name }}
                  <button
                    type="button"
                    @click="removeTag(tag.ID)"
                    class="hover:text-green-900"
                  >
                    <i class="fa fa-times text-xs"></i>
                  </button>
                </span>
              </div>

              <!-- Add Tags Dropdown -->
              <div class="relative">
                <button
                  type="button"
                  @click="toggleTagDropdown"
                  class="inline-flex items-center gap-1 px-2 py-1 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded hover:bg-gray-50"
                >
                  <i class="fa fa-plus text-xs"></i>
                  Add Tags
                </button>

                <!-- Tags Dropdown -->
                <div
                  v-if="showTagDropdown"
                  class="absolute z-10 mt-1 w-64 bg-white border border-gray-300 rounded shadow-lg max-h-48 overflow-y-auto"
                >
                  <div v-if="availableTagsToAdd.length === 0" class="px-2 py-1.5 text-xs text-gray-500 text-center">
                    {{ tags.length === 0 ? 'No tags available' : 'All tags selected' }}
                  </div>
                  <button
                    v-for="tag in availableTagsToAdd"
                    :key="tag.ID"
                    type="button"
                    @click="addTag(tag.ID)"
                    class="w-full text-left px-2 py-1.5 text-xs hover:bg-sage-pale hover:bg-opacity-30"
                  >
                    {{ tag.name }}
                  </button>
                </div>
              </div>

              <!-- Click outside to close dropdown -->
              <div
                v-if="showTagDropdown"
                @click="showTagDropdown = false"
                class="fixed inset-0 z-0"
              ></div>
            </div>
            
            <!-- Date and Amount Row -->
            <div class="bg-sage-pale bg-opacity-30 rounded p-1.5 border border-sage border-opacity-20 mt-2">
              <div class="grid grid-cols-2 gap-2">
                <div>
                  <label class="block text-xs font-medium text-gray-700 mb-0.5">Date</label>
                  <input
                    v-model="formData.date"
                    type="date"
                    required
                    class="bg-white border border-gray-300 text-gray-900 text-xs rounded focus:ring-sage focus:border-sage block w-full p-1"
                  />
                </div>
                
                <div>
                  <label class="block text-xs font-medium text-gray-700 mb-0.5">Amount</label>
                  <div class="flex rounded border border-gray-300 bg-white focus-within:ring-sage focus-within:border-sage overflow-hidden">
                    <span class="flex items-center px-1.5 text-gray-500 text-xs bg-gray-50 border-r border-gray-300">$</span>
                    <input
                      v-model.number="formData.amount"
                      type="number"
                      step="0.01"
                      min="0"
                      required
                      placeholder="0.00"
                      class="flex-1 text-gray-900 text-xs p-1 focus:outline-none border-0"
                    />
                  </div>
                </div>
              </div>
            </div>
            
            <!-- Description -->
            <div class="mt-2">
              <label class="block text-xs font-medium text-gray-700 mb-0.5">
                Description <span class="text-gray-400">(required)</span>
              </label>
              <textarea
                v-model="formData.description"
                rows="3"
                required
                placeholder="What was this expense for?"
                class="bg-white border border-gray-300 text-gray-900 text-xs rounded focus:ring-sage focus:border-sage block w-full p-1.5"
              ></textarea>
            </div>

            <!-- Receipt -->
            <div class="mt-2">
              <label class="block text-xs font-medium text-gray-700 mb-0.5">Receipt (optional)</label>
              <div v-if="isEditing && editingExpense?.receipt_id" class="mt-1 mb-1.5 p-1.5 bg-sage-pale bg-opacity-30 rounded border border-sage border-opacity-20">
                <div class="flex items-center justify-between">
                  <span class="text-xs text-gray-600">
                    <i class="fa fa-check-circle text-sage mr-1"></i>
                    Receipt attached
                  </span>
                  <button
                    type="button"
                    @click="viewReceipt(editingExpense.receipt_id)"
                    class="text-xs text-sage hover:text-sage-dark"
                  >
                    View
                  </button>
                </div>
              </div>
              <input
                ref="receiptInput"
                type="file"
                accept="image/*,.pdf"
                @change="handleReceiptSelect"
                class="block w-full text-xs text-gray-600 
                       file:mr-2 file:py-1 file:px-2 
                       file:rounded file:border-0 
                       file:text-xs file:font-medium
                       file:bg-gray-100 file:text-gray-700 
                       hover:file:bg-gray-200"
              />
              <p v-if="isEditing && editingExpense?.receipt_id" class="mt-1 text-xs text-gray-500">
                Upload a new file to replace the existing receipt
              </p>
            </div>
            
            <!-- Error -->
            <div v-if="modalError" class="mt-2 p-1.5 bg-red-50 border border-red-200 text-red-700 text-xs rounded">
              {{ modalError }}
            </div>
            
            <!-- Footer -->
            <div class="mt-4 flex flex-row-reverse">
              <button 
                type="submit" 
                class="inline-flex justify-center rounded bg-sage px-2.5 py-1 text-xs font-semibold text-white shadow-sm hover:bg-sage-dark sm:ml-2 sm:w-auto"
                :disabled="isSubmitting"
              >
                <template v-if="isSubmitting">
                  <svg class="animate-spin -ml-1 mr-2 h-3 w-3 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Saving...
                </template>
                <template v-else>
                  {{ isEditing ? 'Update' : 'Create' }}
                </template>
              </button>
              <button 
                type="button" 
                class="mt-2 inline-flex justify-center rounded bg-white px-2.5 py-1 text-xs font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                @click="closeModal"
                :disabled="isSubmitting"
              >
                Cancel
              </button>
              <div class="sm:flex-grow">
                <span v-if="isSubmitting" class="text-2xs text-gray-500 ml-2 sm:ml-0">Processing...</span>
              </div>
              <button 
                v-if="isEditing"
                type="button" 
                class="mt-2 inline-flex justify-center rounded bg-red-600 px-2.5 py-1 text-xs font-semibold text-white shadow-sm hover:bg-red-500 sm:mt-0 sm:w-auto"
                @click="confirmDelete"
                :disabled="isSubmitting"
              >
                Delete
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { fetchProjects as getProjects } from '../../api/projects';
import { 
  getExpenses as fetchExpenses, 
  createExpense, 
  updateExpense,
  deleteExpense,
  submitExpense,
  type Expense
} from '../../api/expenses';
import { getExpenseCategories } from '../../api/expenseCategories';
import { getExpenseTags } from '../../api/expenseTags';
import type { ExpenseCategory } from '../../types/ExpenseCategory';
import type { ExpenseTag } from '../../types/ExpenseTag';

interface Project {
  ID: number;
  name: string;
}

const expenses = ref<Expense[]>([]);
const projects = ref<Project[]>([]);
const categories = ref<ExpenseCategory[]>([]);
const tags = ref<ExpenseTag[]>([]);
const isLoading = ref(false);
const searchQuery = ref('');

const filters = ref({
  status: '',
  projectId: '',
});

// Modal state
const showModal = ref(false);
const isEditing = ref(false);
const isSubmitting = ref(false);
const modalError = ref<string | null>(null);
const receiptInput = ref<HTMLInputElement | null>(null);
const selectedReceipt = ref<File | null>(null);

const formData = ref({
  id: 0,
  date: new Date().toISOString().split('T')[0],
  project_id: '',
  amount: 0,
  description: '',
  category_id: '',
  tag_ids: [] as number[],
  is_reimbursable: false,
});

const isInternalExpense = ref(false); // Track if this is an internal expense (no project)

const editingExpense = ref<Expense | null>(null);
const showTagDropdown = ref(false);

const selectedTagsDisplay = computed(() => {
  return tags.value.filter(tag => formData.value.tag_ids.includes(tag.ID));
});

const availableTagsToAdd = computed(() => {
  return tags.value.filter(tag => !formData.value.tag_ids.includes(tag.ID));
});

const filteredExpenses = computed(() => {
  let filtered = expenses.value;
  
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(e =>
      e.description.toLowerCase().includes(query)
    );
  }
  
  return filtered;
});

function getProjectName(projectId: number): string {
  const project = projects.value.find(p => p.ID === projectId);
  return project ? project.name : '-';
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString();
}

function formatCurrency(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`;
}

function formatStatus(status: string): string {
  switch (status) {
    case 'EXPENSE_STATE_DRAFT':
      return 'Draft';
    case 'EXPENSE_STATE_SUBMITTED':
      return 'Submitted';
    case 'EXPENSE_STATE_APPROVED':
      return 'Approved';
    case 'EXPENSE_STATE_REJECTED':
      return 'Rejected';
    case 'EXPENSE_STATE_INVOICED':
      return 'Invoiced';
    case 'EXPENSE_STATE_PAID':
      return 'Paid';
    default:
      return status;
  }
}

function getStatusColor(status: string): string {
  switch (status) {
    case 'EXPENSE_STATE_DRAFT':
      return 'bg-gray-100 text-gray-800';
    case 'EXPENSE_STATE_SUBMITTED':
      return 'bg-yellow-100 text-yellow-800';
    case 'EXPENSE_STATE_APPROVED':
      return 'bg-green-100 text-green-800';
    case 'EXPENSE_STATE_REJECTED':
      return 'bg-red-100 text-red-800';
    case 'EXPENSE_STATE_INVOICED':
      return 'bg-blue-100 text-blue-800';
    case 'EXPENSE_STATE_PAID':
      return 'bg-purple-100 text-purple-800';
    default:
      return 'bg-gray-100 text-gray-800';
  }
}

async function loadExpenses() {
  isLoading.value = true;
  try {
    const params: any = {};
    if (filters.value.status) params.status = filters.value.status;
    if (filters.value.projectId) params.project_id = parseInt(filters.value.projectId);
    
    expenses.value = await fetchExpenses(params);
  } catch (error) {
    console.error('Failed to fetch expenses:', error);
    expenses.value = [];
  } finally {
    isLoading.value = false;
  }
}

async function fetchProjects() {
  try {
    const response = await getProjects();
    projects.value = response;
  } catch (error) {
    console.error('Failed to fetch projects:', error);
    projects.value = [];
  }
}

function openCreateModal() {
  isEditing.value = false;
  isInternalExpense.value = false; // Reset to false for new expense
  formData.value = {
    id: 0,
    date: new Date().toISOString().split('T')[0],
    project_id: '',
    amount: 0,
    description: '',
    category_id: '',
    tag_ids: [],
    is_reimbursable: false,
  };
  selectedReceipt.value = null;
  modalError.value = null;
  showModal.value = true;
}

function openEditModal(expense: Expense) {
  isEditing.value = true;
  editingExpense.value = expense;
  
  // Set internal expense flag if project_id is null
  isInternalExpense.value = !expense.project_id;
  
  formData.value = {
    id: expense.ID,
    date: expense.date.split('T')[0], // Extract YYYY-MM-DD from ISO date
    project_id: expense.project_id ? String(expense.project_id) : '',
    amount: expense.amount / 100, // Convert cents to dollars
    description: expense.description,
    category_id: String(expense.category_id || ''),
    tag_ids: (expense as any).tags ? (expense as any).tags.map((t: any) => t.ID) : [],
    is_reimbursable: expense.is_reimbursable || false,
  };
  selectedReceipt.value = null;
  modalError.value = null;
  showModal.value = true;
}

function closeModal() {
  showModal.value = false;
  modalError.value = null;
  selectedReceipt.value = null;
  editingExpense.value = null;
  showTagDropdown.value = false;
}

function toggleTagDropdown() {
  showTagDropdown.value = !showTagDropdown.value;
}

function addTag(tagId: number) {
  if (!formData.value.tag_ids.includes(tagId)) {
    formData.value.tag_ids.push(tagId);
  }
  showTagDropdown.value = false;
}

function removeTag(tagId: number) {
  const index = formData.value.tag_ids.indexOf(tagId);
  if (index > -1) {
    formData.value.tag_ids.splice(index, 1);
  }
}

function handleReceiptSelect(event: Event) {
  const target = event.target as HTMLInputElement;
  selectedReceipt.value = target.files?.[0] || null;
}

async function submitForm() {
  isSubmitting.value = true;
  modalError.value = null;
  
  try {
    const formDataToSend = new FormData();
    
    // Only append project_id if not an internal expense
    if (!isInternalExpense.value && formData.value.project_id) {
      formDataToSend.append('project_id', formData.value.project_id);
    }
    // For internal expenses, don't append project_id (backend will treat as null)
    
    formDataToSend.append('amount', formData.value.amount.toString());
    formDataToSend.append('date', formData.value.date);
    formDataToSend.append('description', formData.value.description);
    formDataToSend.append('category_id', formData.value.category_id);
    formDataToSend.append('tag_ids', formData.value.tag_ids.join(',')); // Comma-separated tag IDs
    // Only include is_reimbursable for internal expenses
    if (isInternalExpense.value) {
      formDataToSend.append('is_reimbursable', formData.value.is_reimbursable ? 'true' : 'false');
    }
    // payment_account_code removed - determined during reconciliation with bank statement
    
    if (selectedReceipt.value) {
      formDataToSend.append('receipt', selectedReceipt.value);
    }
    
    console.log('Form data being sent:', {
      isInternal: isInternalExpense.value,
      project_id: isInternalExpense.value ? null : formData.value.project_id,
      amount: formData.value.amount,
      date: formData.value.date,
      description: formData.value.description,
      category_id: formData.value.category_id,
      tag_ids: formData.value.tag_ids,
      has_receipt: !!selectedReceipt.value,
      is_editing: isEditing.value,
      expense_id: formData.value.id
    });
    
    if (isEditing.value) {
      console.log('Updating expense ID:', formData.value.id);
      await updateExpense(formData.value.id, formDataToSend);
    } else {
      console.log('Creating new expense');
      await createExpense(formDataToSend);
    }
    
    await loadExpenses();
    closeModal();
  } catch (error: any) {
    console.error('Failed to save expense:', error);
    modalError.value = error.response?.data?.error || error.message || 'Failed to save expense';
  } finally {
    isSubmitting.value = false;
  }
}

async function handleSubmitExpense(expenseId: number) {
  try {
    await submitExpense(expenseId);
    await loadExpenses();
  } catch (error) {
    console.error('Failed to submit expense:', error);
  }
}

async function handleDeleteExpense(expenseId: number) {
  try {
    await deleteExpense(expenseId);
    await loadExpenses();
  } catch (error) {
    console.error('Failed to delete expense:', error);
  }
}

async function confirmDelete() {
  try {
    await deleteExpense(formData.value.id);
    closeModal();
    await loadExpenses();
  } catch (error) {
    console.error('Failed to delete expense:', error);
    modalError.value = 'Failed to delete expense';
  }
}

async function viewReceipt(receiptId: number) {
  try {
    const expense = expenses.value.find(e => e.receipt_id === receiptId);
    if (!expense?.receipt) {
      console.error('Receipt not found');
      return;
    }
    
    // Use download endpoint for GCS-stored receipts
    const downloadUrl = `/api/assets/${receiptId}/download`;
    
    // Fetch with authentication headers
    const token = localStorage.getItem('snowpack_token');
    const response = await fetch(downloadUrl, {
      headers: {
        'x-access-token': token || ''
      }
    });
    
    if (!response.ok) {
      throw new Error(`Download failed: ${response.statusText}`);
    }
    
    // Create blob and download
    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = expense.receipt.name || `receipt_${receiptId}`;
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);
  } catch (error) {
    console.error('Failed to view receipt:', error);
  }
}

onMounted(async () => {
  await fetchProjects();
  await fetchCategories();
  await fetchTags();
  await loadExpenses();
});

async function fetchCategories() {
  try {
    categories.value = await getExpenseCategories(true); // Only active categories
    console.log('Fetched expense categories:', categories.value);
    if (categories.value.length === 0) {
      console.warn('No expense categories found - you may need to create some in the Expense Config page');
    }
  } catch (error) {
    console.error('Failed to fetch expense categories:', error);
  }
}

async function fetchTags() {
  try {
    tags.value = await getExpenseTags(true); // Only active tags
  } catch (error) {
    console.error('Failed to fetch expense tags:', error);
  }
}

// Payment account removed - actual payment account will be determined during bank reconciliation
</script>

