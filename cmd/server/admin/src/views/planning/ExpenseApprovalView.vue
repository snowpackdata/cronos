<template>
  <div class="max-w-7xl mx-auto px-3 py-2">
    <!-- Header -->
    <div class="mb-2 flex justify-between items-center">
      <div>
        <h1 class="text-lg font-bold text-gray-900">Expense Approvals</h1>
        <p class="text-xs text-gray-500">
          Review and approve submitted expenses
        </p>
      </div>
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
            <option value="EXPENSE_STATE_SUBMITTED">Pending Approval</option>
            <option value="EXPENSE_STATE_APPROVED">Approved</option>
            <option value="EXPENSE_STATE_REJECTED">Rejected</option>
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

    <!-- Expenses Table -->
    <div class="bg-white shadow rounded overflow-hidden">
      <div v-if="isLoading" class="p-4 text-center text-xs text-gray-500">
        Loading...
      </div>
      <div v-else-if="filteredExpenses.length === 0" class="p-4 text-center text-xs text-gray-500">
        No expenses found
      </div>
      <table v-else class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-2 py-1 text-left text-xs font-medium text-gray-500 uppercase">Date</th>
            <th class="px-2 py-1 text-left text-xs font-medium text-gray-500 uppercase">Submitter</th>
            <th class="px-2 py-1 text-left text-xs font-medium text-gray-500 uppercase">Project</th>
            <th class="px-2 py-1 text-left text-xs font-medium text-gray-500 uppercase">Category</th>
            <th class="px-2 py-1 text-left text-xs font-medium text-gray-500 uppercase">Tags</th>
            <th class="px-2 py-1 text-left text-xs font-medium text-gray-500 uppercase">Description</th>
            <th class="px-2 py-1 text-right text-xs font-medium text-gray-500 uppercase">Amount</th>
            <th class="px-2 py-1 text-center text-xs font-medium text-gray-500 uppercase">Receipt</th>
            <th class="px-2 py-1 text-center text-xs font-medium text-gray-500 uppercase">Status</th>
            <th class="px-2 py-1 text-center text-xs font-medium text-gray-500 uppercase">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-for="expense in filteredExpenses" :key="expense.ID" class="hover:bg-sage-pale hover:bg-opacity-30 transition-colors">
            <td class="px-2 py-1 text-xs text-gray-900">
              {{ formatDate(expense.date) }}
            </td>
            <td class="px-2 py-1">
              <div class="flex items-center gap-2">
                <StaffAvatar v-if="expense.submitter" :employee="expense.submitter" size="xs" />
                <span class="text-xs text-gray-900">{{ getSubmitterName(expense) }}</span>
              </div>
            </td>
            <td class="px-2 py-1 text-xs text-gray-600">
              {{ getProjectName(expense.project_id) }}
            </td>
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
            <td class="px-2 py-1 text-xs text-gray-900">
              {{ expense.description }}
            </td>
            <td class="px-2 py-1 text-right text-xs text-gray-900">
              {{ formatCurrency(expense.amount) }}
            </td>
            <td class="px-2 py-1 text-center">
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
            <td class="px-2 py-1 text-center">
              <span :class="getStatusColor(expense.state)" class="px-1.5 py-0.5 text-xs font-semibold rounded">
                {{ formatStatus(expense.state) }}
              </span>
            </td>
            <td class="px-2 py-1 text-center">
              <div class="flex items-center justify-center gap-1">
                <button
                  v-if="expense.state === 'EXPENSE_STATE_SUBMITTED'"
                  @click="handleApprove(expense.ID)"
                  class="inline-flex items-center justify-center w-7 h-7 text-white bg-sage hover:bg-sage-dark rounded shadow-sm transition-colors"
                  title="Approve expense"
                >
                  <i class="fa fa-check text-xs"></i>
                </button>
                <button
                  v-if="expense.state === 'EXPENSE_STATE_SUBMITTED'"
                  @click="handleReject(expense.ID)"
                  class="inline-flex items-center justify-center w-7 h-7 text-white bg-red-600 hover:bg-red-500 rounded shadow-sm transition-colors"
                  title="Reject expense"
                >
                  <i class="fa fa-times text-xs"></i>
                </button>
                <span v-if="expense.state === 'EXPENSE_STATE_APPROVED'" class="text-xs text-gray-600">
                  <i class="fa fa-check-circle text-sage mr-1"></i>
                  {{ getApproverName(expense) }}
                </span>
                <span v-if="expense.state === 'EXPENSE_STATE_REJECTED'" class="text-xs text-gray-600 cursor-help" :title="expense.rejection_reason">
                  <i class="fa fa-times-circle text-red-600 mr-1"></i>
                  Rejected
                </span>
                <span v-if="expense.state !== 'EXPENSE_STATE_SUBMITTED' && expense.state !== 'EXPENSE_STATE_APPROVED' && expense.state !== 'EXPENSE_STATE_REJECTED'" class="text-xs text-gray-400">-</span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Reject Modal -->
    <div v-if="showRejectModal" class="fixed inset-0 z-10 w-screen overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4 text-center sm:p-0">
        <!-- Backdrop -->
        <div class="fixed inset-0 bg-gray-500/75 transition-opacity" @click="closeRejectModal"></div>
        
        <!-- Modal -->
        <div class="relative transform overflow-hidden rounded-lg bg-white text-left shadow-xl transition-all sm:my-6 sm:w-full sm:max-w-md">
          <!-- Header -->
          <div class="bg-red-600 text-white px-4 py-2 sm:px-4 sm:py-2 flex justify-between items-center">
            <h3 class="text-base font-semibold leading-6">
              Reject Expense
            </h3>
            <button
              @click="closeRejectModal"
              type="button"
              class="text-white hover:text-gray-200 focus:outline-none"
            >
              <i class="fa fa-times h-5 w-5"></i>
            </button>
          </div>
          
          <!-- Body -->
          <div class="bg-white px-3 py-4 sm:p-6">
            <div class="mb-3">
              <label class="block text-2xs font-medium text-gray-700">
                Reason for Rejection <span class="text-gray-400">(required)</span>
              </label>
              <textarea
                v-model="rejectionReason"
                rows="4"
                class="mt-0.5 bg-white border border-gray-300 text-gray-900 text-2xs rounded focus:ring-red-500 focus:border-red-500 block w-full p-1"
                placeholder="Explain why this expense is being rejected..."
                autofocus
              ></textarea>
            </div>

            <div v-if="modalError" class="mb-3 p-2 bg-red-50 border border-red-200 text-red-700 text-2xs rounded">
              {{ modalError }}
            </div>

            <div class="flex flex-row-reverse">
              <button
                @click="confirmReject"
                type="button"
                :disabled="!rejectionReason || isSubmitting"
                class="inline-flex justify-center rounded bg-red-600 px-2.5 py-1 text-xs font-semibold text-white shadow-sm hover:bg-red-500 sm:ml-2 sm:w-auto"
              >
                <template v-if="isSubmitting">
                  <svg class="animate-spin -ml-1 mr-2 h-3 w-3 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Rejecting...
                </template>
                <template v-else>
                  Reject
                </template>
              </button>
              <button
                @click="closeRejectModal"
                type="button"
                class="mt-2 inline-flex justify-center rounded bg-white px-2.5 py-1 text-xs font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { fetchProjects as getProjects } from '../../api/projects';
import { 
  getExpensesForReview as fetchExpenses, 
  approveExpense, 
  rejectExpense,
  type Expense
} from '../../api/expenses';
import StaffAvatar from '../../components/StaffAvatar.vue';

interface Project {
  ID: number;
  name: string;
}

const expenses = ref<Expense[]>([]);
const projects = ref<Project[]>([]);
const isLoading = ref(false);
const searchQuery = ref('');

const filters = ref({
  status: '', // Default to all expenses
});

// Reject modal state
const showRejectModal = ref(false);
const expenseToReject = ref<number | null>(null);
const rejectionReason = ref('');
const isSubmitting = ref(false);
const modalError = ref<string | null>(null);

const filteredExpenses = computed(() => {
  let filtered = expenses.value;
  
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(e =>
      getSubmitterName(e).toLowerCase().includes(query)
    );
  }
  
  return filtered;
});

function getProjectName(projectId: number): string {
  const project = projects.value.find(p => p.ID === projectId);
  return project ? project.name : '-';
}

function getSubmitterName(expense: Expense): string {
  if (!expense.submitter) return '-';
  return `${expense.submitter.first_name} ${expense.submitter.last_name}`;
}

function getApproverName(expense: Expense): string {
  if (!expense.approver) return '-';
  return `${expense.approver.first_name} ${expense.approver.last_name}`;
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
    
    expenses.value = await fetchExpenses(params);
  } catch (error) {
    console.error('Failed to fetch expenses:', error);
    expenses.value = [];
  } finally {
    isLoading.value = false;
  }
}

async function fetchProjectList() {
  try {
    const response = await getProjects();
    projects.value = response;
  } catch (error) {
    console.error('Failed to fetch projects:', error);
    projects.value = [];
  }
}

async function handleApprove(expenseId: number) {
  try {
    await approveExpense(expenseId);
    await loadExpenses();
  } catch (error) {
    console.error('Failed to approve expense:', error);
  }
}

function handleReject(expenseId: number) {
  expenseToReject.value = expenseId;
  rejectionReason.value = '';
  modalError.value = null;
  showRejectModal.value = true;
}

function closeRejectModal() {
  showRejectModal.value = false;
  expenseToReject.value = null;
  rejectionReason.value = '';
  modalError.value = null;
}

async function confirmReject() {
  if (!expenseToReject.value || !rejectionReason.value) return;
  
  isSubmitting.value = true;
  modalError.value = null;
  
  try {
    await rejectExpense(expenseToReject.value, rejectionReason.value);
    await loadExpenses();
    closeRejectModal();
  } catch (error: any) {
    modalError.value = error.message || 'Failed to reject expense';
  } finally {
    isSubmitting.value = false;
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
  await fetchProjectList();
  await loadExpenses();
});
</script>

