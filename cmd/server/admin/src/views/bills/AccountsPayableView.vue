<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import type { Bill } from '../../types/Bill';
import { getBills } from '../../api';
import StaffAvatar from '../../components/StaffAvatar.vue';

// State
const bills = ref<Bill[]>([]);
const isLoading = ref(true);
const error = ref<string | null>(null);
const expandedBillIds = ref<Set<number>>(new Set());
const successMessage = ref<string | null>(null);
const billError = ref<{ [key: number]: string | null }>({});

// Payment date modal
const showPaymentDateModal = ref(false);
const selectedBillForPayment = ref<Bill | null>(null);
const paymentDate = ref<string>(new Date().toISOString().split('T')[0]);

// PDF generation state
const generatingPDF = ref<{ [key: number]: boolean }>({});

// Filters
const selectedState = ref<string>('all');
const startDate = ref('');
const endDate = ref('');
const searchTerm = ref('');

// Fetch bills on component mount
onMounted(async () => {
  await fetchBills();
});

// Fetch all bills
const fetchBills = async () => {
  isLoading.value = true;
  error.value = null;
  
  try {
    bills.value = await getBills();
  } catch (err) {
    console.error('Error fetching bills:', err);
    error.value = 'Failed to load bills. Please try again.';
  } finally {
    isLoading.value = false;
  }
};

// Filter bills based on selected criteria
const filteredBills = computed(() => {
  let filtered = bills.value;
  
  // Filter by state
  if (selectedState.value !== 'all') {
    filtered = filtered.filter(bill => bill.state === selectedState.value);
  }
  
  // Filter by date range
  if (startDate.value) {
    const start = new Date(startDate.value);
    filtered = filtered.filter(bill => {
      const billDate = getBillDate(bill);
      return billDate && new Date(billDate) >= start;
    });
  }
  if (endDate.value) {
    const end = new Date(endDate.value);
    end.setHours(23, 59, 59, 999); // Include the entire end date
    filtered = filtered.filter(bill => {
      const billDate = getBillDate(bill);
      return billDate && new Date(billDate) <= end;
    });
  }
  
  // Filter by search term
  if (searchTerm.value) {
    const term = searchTerm.value.toLowerCase();
    filtered = filtered.filter(bill => 
      (bill.bill_number && bill.bill_number.toLowerCase().includes(term)) ||
      (bill.vendor_name && bill.vendor_name.toLowerCase().includes(term)) ||
      (bill.user && `${bill.user.first_name} ${bill.user.last_name}`.toLowerCase().includes(term))
    );
  }
  
  // Sort by date (most recent first)
  filtered.sort((a, b) => {
    const dateA = new Date(getBillDate(a) || 0);
    const dateB = new Date(getBillDate(b) || 0);
    return dateB.getTime() - dateA.getTime();
  });
  
  return filtered;
});

// Toggle bill expansion
const toggleBillExpansion = (billId: number) => {
  if (expandedBillIds.value.has(billId)) {
    expandedBillIds.value.delete(billId);
  } else {
    expandedBillIds.value.add(billId);
  }
};

// Check if bill is expanded
const isBillExpanded = (billId: number) => {
  return expandedBillIds.value.has(billId);
};

// Format date for display (YYYY-MM-DD format)
const formatDate = (dateString: string | undefined | null): string => {
  if (!dateString) return '-';
  if (dateString.startsWith('00') || dateString === '0001-01-01T00:00:00Z') return '-';
  const date = new Date(dateString);
  const year = date.getUTCFullYear();
  const month = String(date.getUTCMonth() + 1).padStart(2, '0');
  const day = String(date.getUTCDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
};

// Get the most relevant date to display for a bill
const getBillDate = (bill: Bill): string => {
  // Priority: paid date > accepted date > created date > period start
  if (bill.date_paid) return bill.date_paid;
  if (bill.accepted_at) return bill.accepted_at;
  if (bill.period_start) return bill.period_start;
  return bill.date_created || '';
};

// Format bill number as YYYYNNNN
const formatBillNumber = (bill: Bill): string => {
  let year = new Date().getFullYear();
  if (bill.accepted_at) {
    year = new Date(bill.accepted_at).getFullYear();
  } else if (bill.date_created) {
    year = new Date(bill.date_created).getFullYear();
  }
  const paddedId = bill.ID.toString().padStart(4, '0');
  return `${year}${paddedId}`;
};

// Format currency for display
const formatCurrency = (amount: number) => {
  // Amount is in cents, convert to dollars
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount / 100);
};

// Get bill name
const getBillName = (bill: Bill): string => {
  if (bill.description) {
    return bill.description;
  }
  return 'Payroll';
};

// Get state color class
const getStateColorClass = (state: string) => {
  switch (state) {
    case 'BILL_STATE_DRAFT':
      return 'bg-gray-100 text-gray-700';
    case 'BILL_STATE_ACCEPTED':
      return 'bg-blue-100 text-blue-700';
    case 'BILL_STATE_PAID':
      return 'bg-green-100 text-green-700';
    case 'BILL_STATE_VOID':
      return 'bg-red-100 text-red-700';
    default:
      return 'bg-gray-100 text-gray-700';
  }
};

// Format state for display
const formatState = (state: string | undefined) => {
  if (!state) return 'Unknown';
  return state.replace('BILL_STATE_', '').replace(/_/g, ' ');
};

// Format line item type
const formatLineItemType = (type: string | undefined) => {
  if (!type) return 'Unknown';
  return type.replace('LINE_ITEM_TYPE_', '').replace(/_/g, ' ');
};

// Check if bill is overdue
const isOverdue = (bill: Bill): boolean => {
  // Only unpaid/accepted bills can be overdue
  if (bill.state === 'BILL_STATE_PAID' || bill.state === 'BILL_STATE_VOID' || bill.state === 'BILL_STATE_DRAFT') {
    return false;
  }
  
  const dueDate = bill.date_due;
  if (!dueDate || dueDate.startsWith('00') || dueDate === '0001-01-01T00:00:00Z') {
    return false;
  }
  
  const due = new Date(dueDate);
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  
  return due < today;
};

// Get days overdue (negative if not yet due)
const getDaysOverdue = (bill: Bill): number => {
  const dueDate = bill.date_due;
  if (!dueDate || dueDate.startsWith('00') || dueDate === '0001-01-01T00:00:00Z') {
    return 0;
  }
  
  const due = new Date(dueDate);
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  
  const diffTime = today.getTime() - due.getTime();
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  
  return diffDays;
};

// Get employee name
const getEmployeeName = (bill: Bill) => {
  if (bill.user) {
    return `${bill.user.first_name} ${bill.user.last_name}`;
  }
  return bill.vendor_name || 'N/A';
};

// Clear all filters
const clearFilters = () => {
  selectedState.value = 'all';
  startDate.value = '';
  endDate.value = '';
  searchTerm.value = '';
};

// Action handlers
const acceptBill = async (bill: Bill) => {
  billError.value[bill.ID] = null;
  try {
    const token = localStorage.getItem('snowpack_token');
    const response = await fetch(`/api/bills/${bill.ID}/accept`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-access-token': token || ''
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to accept bill: ${response.statusText}`);
    }
    
    successMessage.value = `Bill #${formatBillNumber(bill)} accepted`;
    await fetchBills();
    setTimeout(() => { successMessage.value = null; }, 3000);
  } catch (err) {
    console.error('Error accepting bill:', err);
    billError.value[bill.ID] = err instanceof Error ? err.message : 'Failed to accept bill';
  }
};

// Show payment date modal
const showPaymentModal = (bill: Bill) => {
  selectedBillForPayment.value = bill;
  paymentDate.value = new Date().toISOString().split('T')[0]; // Default to today
  showPaymentDateModal.value = true;
};

// Cancel payment modal
const cancelPaymentModal = () => {
  showPaymentDateModal.value = false;
  selectedBillForPayment.value = null;
};

// Mark bill as paid with specified payment date
const markBillPaid = async () => {
  if (!selectedBillForPayment.value) return;
  
  const bill = selectedBillForPayment.value;
  billError.value[bill.ID] = null;
  
  try {
    const token = localStorage.getItem('snowpack_token');
    const response = await fetch(`/api/bills/${bill.ID}/paid`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-access-token': token || ''
      },
      body: JSON.stringify({ payment_date: paymentDate.value })
    });
    
    if (!response.ok) {
      throw new Error(`Failed to mark bill as paid: ${response.statusText}`);
    }
    
    successMessage.value = `Bill #${formatBillNumber(bill)} marked as paid on ${paymentDate.value}`;
    showPaymentDateModal.value = false;
    selectedBillForPayment.value = null;
    await fetchBills();
    setTimeout(() => { successMessage.value = null; }, 3000);
  } catch (err) {
    console.error('Error marking bill as paid:', err);
    billError.value[bill.ID] = err instanceof Error ? err.message : 'Failed to mark bill as paid';
    showPaymentDateModal.value = false;
  }
};

const voidBill = async (bill: Bill) => {
  if (!confirm(`Are you sure you want to void bill #${formatBillNumber(bill)}? This action cannot be undone.`)) {
    return;
  }
  
  billError.value[bill.ID] = null;
  try {
    const token = localStorage.getItem('snowpack_token');
    const response = await fetch(`/api/bills/${bill.ID}/void`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-access-token': token || ''
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to void bill: ${response.statusText}`);
    }
    
    successMessage.value = `Bill #${formatBillNumber(bill)} voided`;
    await fetchBills();
    setTimeout(() => { successMessage.value = null; }, 3000);
  } catch (err) {
    console.error('Error voiding bill:', err);
    billError.value[bill.ID] = err instanceof Error ? err.message : 'Failed to void bill';
  }
};

// Regenerate bill PDF
const regenerateBillPDF = async (bill: Bill) => {
  billError.value[bill.ID] = null;
  generatingPDF.value[bill.ID] = true;
  
  try {
    const token = localStorage.getItem('snowpack_token');
    const response = await fetch(`/api/bills/${bill.ID}/regenerate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-access-token': token || ''
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to regenerate PDF: ${response.statusText}`);
    }
    
    successMessage.value = `PDF regenerated for bill #${formatBillNumber(bill)}`;
    await fetchBills();
    setTimeout(() => { successMessage.value = null; }, 3000);
  } catch (err) {
    console.error('Error regenerating PDF:', err);
    billError.value[bill.ID] = err instanceof Error ? err.message : 'Failed to regenerate PDF';
  } finally {
    generatingPDF.value[bill.ID] = false;
  }
};
</script>

<template>
  <div class="px-4 sm:px-6 lg:px-8">
    <div class="sm:flex sm:items-center">
      <div class="sm:flex-auto">
        <h1 class="text-xl font-semibold text-gray-900">Accounts Payable</h1>
      </div>
    </div>

    <!-- Filters -->
    <div class="mt-4 bg-white p-4 rounded-lg shadow border border-gray-200">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <!-- Search -->
        <div>
          <label for="search" class="block text-sm font-medium text-gray-700 mb-1">Search</label>
          <input
            id="search"
            v-model="searchTerm"
            type="text"
            placeholder="Bill # or Employee..."
            class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        <!-- State Filter -->
        <div>
          <label for="state" class="block text-sm font-medium text-gray-700 mb-1">Status</label>
          <select
            id="state"
            v-model="selectedState"
            class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="all">All</option>
            <option value="BILL_STATE_DRAFT">Draft</option>
            <option value="BILL_STATE_PAID">Paid</option>
            <option value="BILL_STATE_VOID">Void</option>
          </select>
        </div>

        <!-- Start Date -->
        <div>
          <label for="start-date" class="block text-sm font-medium text-gray-700 mb-1">Start Date</label>
          <input
            id="start-date"
            v-model="startDate"
            type="date"
            class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        <!-- End Date -->
        <div>
          <label for="end-date" class="block text-sm font-medium text-gray-700 mb-1">End Date</label>
          <input
            id="end-date"
            v-model="endDate"
            type="date"
            class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
      </div>
      <div class="mt-3 flex justify-end">
        <button
          @click="clearFilters"
          class="text-sm text-gray-600 hover:text-gray-900"
        >
          Clear Filters
        </button>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="isLoading" class="flex flex-col items-center justify-center p-10 bg-white rounded-lg shadow mt-6">
      <i class="fas fa-spinner fa-spin text-4xl text-blue-600 mb-4"></i>
      <span class="text-gray-600">Loading bills...</span>
    </div>
    
    <!-- Error state -->
    <div v-else-if="error" class="flex flex-col items-center justify-center p-10 bg-white rounded-lg shadow mt-6">
      <i class="fas fa-exclamation-circle text-4xl text-red-600 mb-4"></i>
      <span class="text-gray-600 mb-2">{{ error }}</span>
      <button @click="fetchBills" class="btn-secondary mt-4">
        <i class="fas fa-sync mr-2"></i> Retry
      </button>
    </div>
    
    <!-- Empty state -->
    <div v-else-if="filteredBills.length === 0" class="flex flex-col items-center justify-center p-10 bg-white rounded-lg shadow mt-6">
      <i class="fas fa-file-invoice text-5xl text-blue-600 mb-4"></i>
      <p class="text-lg font-medium text-gray-700">No bills found</p>
      <p class="text-gray-600 mb-4">{{ bills.length === 0 ? 'Bills will appear here once they are created' : 'Try adjusting your filters' }}</p>
      <button v-if="bills.length > 0" @click="clearFilters" class="btn-secondary">
        Clear Filters
      </button>
    </div>
    
    <!-- Success Message -->
    <div v-if="successMessage" class="mt-4 bg-green-50 border border-green-200 text-green-800 px-4 py-3 rounded-lg">
      <i class="fas fa-check-circle mr-2"></i>{{ successMessage }}
    </div>

    <!-- Bills Table -->
    <div v-else class="mt-6 bg-white shadow overflow-hidden rounded-lg border border-gray-200">
      <table class="min-w-full divide-y divide-gray-300">
        <thead class="bg-gray-50">
          <tr>
            <th scope="col" class="w-10 py-2 pl-3 pr-2 text-left text-sm font-semibold text-gray-900"></th>
            <th scope="col" class="py-2 px-2 text-left text-sm font-semibold text-gray-900">Bill</th>
            <th scope="col" class="px-2 py-2 text-left text-sm font-semibold text-gray-900">Employee</th>
            <th scope="col" class="px-2 py-2 text-left text-sm font-semibold text-gray-900">Date</th>
            <th scope="col" class="px-2 py-2 text-left text-sm font-semibold text-gray-900">Amount</th>
            <th scope="col" class="px-2 py-2 text-left text-sm font-semibold text-gray-900">Status</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 bg-white">
          <template v-for="bill in filteredBills" :key="bill.ID">
            <!-- Main row -->
            <tr 
              @click="toggleBillExpansion(bill.ID)"
              class="hover:bg-gray-50 cursor-pointer transition-colors"
            >
              <td class="py-2 pl-3 pr-2 text-sm">
                <i :class="isBillExpanded(bill.ID) ? 'fas fa-chevron-down' : 'fas fa-chevron-right'" class="text-gray-400"></i>
              </td>
              <td class="py-2 px-2 text-sm">
                <div class="font-medium text-gray-900">#{{ formatBillNumber(bill) }}</div>
                <div class="text-xs text-gray-600">{{ getBillName(bill) }}</div>
              </td>
              <td class="px-2 py-2 text-sm">
                <div v-if="bill.user" class="flex items-center gap-2">
                  <StaffAvatar :employee="bill.user" size="xs" />
                  <span class="text-gray-700">{{ getEmployeeName(bill) }}</span>
                </div>
                <span v-else class="text-gray-700">{{ getEmployeeName(bill) }}</span>
              </td>
              <td class="px-2 py-2 text-sm text-gray-700">{{ formatDate(getBillDate(bill)) }}</td>
              <td class="px-2 py-2 text-sm font-medium text-gray-900">{{ formatCurrency(bill.total_amount) }}</td>
              <td class="px-2 py-2 text-sm">
                <div class="flex items-center gap-2">
                  <span :class="getStateColorClass(bill.state)" class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium">
                    {{ formatState(bill.state) }}
                  </span>
                  <span v-if="isOverdue(bill)" class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-bold bg-red-100 text-red-800">
                    <i class="fas fa-exclamation-triangle mr-1"></i> {{ getDaysOverdue(bill) }}d overdue
                  </span>
                </div>
              </td>
            </tr>

            <!-- Expanded row with line items -->
            <tr v-if="isBillExpanded(bill.ID)" class="bg-gray-50">
              <td colspan="6" class="px-3 py-3">
                <div class="bg-white rounded-lg border border-gray-200 p-3">
                  <!-- Error Message -->
                  <div v-if="billError[bill.ID]" class="mb-3 bg-red-50 border border-red-200 text-red-800 px-3 py-2 rounded text-sm">
                    <i class="fas fa-exclamation-circle mr-2"></i>{{ billError[bill.ID] }}
                  </div>

                  <div class="flex justify-between items-start mb-3">
                    <div class="flex items-center gap-3">
                      <h4 class="text-sm font-semibold text-gray-900">Bill Details</h4>
                      <a 
                        v-if="bill.file"
                        :href="bill.file"
                        target="_blank"
                        class="text-xs text-blue-600 hover:text-blue-800 flex items-center gap-1"
                      >
                        <i class="fas fa-file-pdf"></i> View PDF
                      </a>
                    </div>
                    
                    <!-- Action Buttons -->
                    <div class="flex gap-2">
                      <!-- Accept button - only for DRAFT bills -->
                      <button 
                        v-if="bill.state === 'BILL_STATE_DRAFT'"
                        @click.stop="acceptBill(bill)"
                        class="px-3 py-1 text-xs font-medium text-white bg-blue-600 hover:bg-blue-700 rounded transition-colors"
                      >
                        <i class="fas fa-check-circle mr-1"></i> Accept
                      </button>
                      
                      <!-- Mark Paid button - for ACCEPTED bills -->
                      <button 
                        v-if="bill.state === 'BILL_STATE_ACCEPTED'"
                        @click.stop="showPaymentModal(bill)"
                        class="px-3 py-1 text-xs font-medium text-white bg-green-600 hover:bg-green-700 rounded transition-colors"
                      >
                        <i class="fas fa-check mr-1"></i> Mark Paid
                      </button>
                      
                      <!-- Generate PDF button - for ACCEPTED bills -->
                      <button 
                        v-if="bill.state === 'BILL_STATE_ACCEPTED'"
                        @click.stop="regenerateBillPDF(bill)"
                        :disabled="generatingPDF[bill.ID]"
                        class="px-3 py-1 text-xs font-medium text-white bg-purple-600 hover:bg-purple-700 rounded transition-colors disabled:opacity-50"
                        title="Regenerate bill PDF"
                      >
                        <i class="fas" :class="generatingPDF[bill.ID] ? 'fa-spinner fa-spin' : 'fa-file-pdf'"></i> 
                        {{ generatingPDF[bill.ID] ? 'Generating...' : 'Generate PDF' }}
                      </button>
                      
                      <!-- Void button - for DRAFT and ACCEPTED bills -->
                      <button 
                        v-if="bill.state === 'BILL_STATE_DRAFT' || bill.state === 'BILL_STATE_ACCEPTED'"
                        @click.stop="voidBill(bill)"
                        class="px-3 py-1 text-xs font-medium text-white bg-red-600 hover:bg-red-700 rounded transition-colors"
                      >
                        <i class="fas fa-times mr-1"></i> Void
                      </button>
                    </div>
                  </div>
                  
                  <!-- Bill metadata -->
                  <div class="grid grid-cols-3 gap-4 mb-3 text-sm">
                    <div>
                      <span class="text-gray-600">Period:</span>
                      <span class="ml-2 text-gray-900">{{ formatDate(bill.period_start) }} - {{ formatDate(bill.period_end) }}</span>
                    </div>
                    <div>
                      <span class="text-gray-600">Hours:</span>
                      <span class="ml-2 text-gray-900">{{ bill.total_hours.toFixed(2) }}</span>
                    </div>
                    <div>
                      <span class="text-gray-600">Due Date:</span>
                      <span class="ml-2" :class="isOverdue(bill) ? 'text-red-700 font-bold' : 'text-gray-900'">
                        {{ formatDate(bill.date_due) }}
                      </span>
                      <span v-if="isOverdue(bill)" class="ml-2 text-xs font-bold text-red-700">
                        ({{ getDaysOverdue(bill) }} days overdue)
                      </span>
                    </div>
                  </div>

                  <!-- Line items -->
                  <div v-if="(bill.line_items && bill.line_items.length > 0) || (bill.recurring_bill_line_items && bill.recurring_bill_line_items.length > 0)">
                    <h5 class="text-sm font-medium text-gray-900 mb-2">Line Items</h5>
                    <table class="min-w-full text-sm">
                      <thead class="bg-gray-100">
                        <tr>
                          <th class="px-2 py-1.5 text-left text-xs font-medium text-gray-600">Type</th>
                          <th class="px-2 py-1.5 text-left text-xs font-medium text-gray-600">Description</th>
                          <th class="px-2 py-1.5 text-right text-xs font-medium text-gray-600">Quantity</th>
                          <th class="px-2 py-1.5 text-right text-xs font-medium text-gray-600">Rate</th>
                          <th class="px-2 py-1.5 text-right text-xs font-medium text-gray-600">Amount</th>
                        </tr>
                      </thead>
                      <tbody class="divide-y divide-gray-200">
                        <tr v-for="item in bill.line_items" :key="`line-${item.ID}`">
                          <td class="px-2 py-1.5 text-gray-700">{{ formatLineItemType(item.type) }}</td>
                          <td class="px-2 py-1.5 text-gray-700">{{ item.description }}</td>
                          <td class="px-2 py-1.5 text-right text-gray-700">{{ item.quantity > 0 ? item.quantity.toFixed(2) : '-' }}</td>
                          <td class="px-2 py-1.5 text-right text-gray-700">{{ item.rate > 0 ? formatCurrency(item.rate * 100) : '-' }}</td>
                          <td class="px-2 py-1.5 text-right font-medium text-gray-900">{{ formatCurrency(item.amount) }}</td>
                        </tr>
                        <tr v-for="item in bill.recurring_bill_line_items" :key="`recurring-${item.ID}`">
                          <td class="px-2 py-1.5 text-gray-700">SALARY</td>
                          <td class="px-2 py-1.5 text-gray-700">{{ item.description }}</td>
                          <td class="px-2 py-1.5 text-right text-gray-700">-</td>
                          <td class="px-2 py-1.5 text-right text-gray-700">-</td>
                          <td class="px-2 py-1.5 text-right font-medium text-gray-900">{{ formatCurrency(item.amount) }}</td>
                        </tr>
                      </tbody>
                      <tfoot class="bg-gray-50 font-semibold">
                        <tr>
                          <td colspan="4" class="px-2 py-1.5 text-right text-gray-900">Total</td>
                          <td class="px-2 py-1.5 text-right text-gray-900">{{ formatCurrency(bill.total_amount) }}</td>
                        </tr>
                      </tfoot>
                    </table>
                  </div>
                  <div v-else class="text-sm text-gray-500 italic">
                    No line items available
                  </div>
                </div>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>
    
    <!-- Payment Date Modal -->
    <div v-if="showPaymentDateModal" class="fixed inset-0 bg-gray-500/75 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        <!-- Header -->
        <div class="bg-sage px-6 py-4 rounded-t-lg">
          <h3 class="text-sm font-semibold text-white">Confirm Payment Date</h3>
        </div>
        
        <!-- Body -->
        <div class="px-6 py-4">
          <p class="text-xs text-gray-700 mb-4">
            Enter the date the payment was made for Bill #{{ selectedBillForPayment ? formatBillNumber(selectedBillForPayment) : '' }}
          </p>
          
          <div class="mb-4">
            <label class="block text-xs font-medium text-gray-700 mb-2">Payment Date</label>
            <input 
              v-model="paymentDate" 
              type="date" 
              class="w-full px-3 py-2 text-xs border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-sage focus:border-transparent"
              :max="new Date().toISOString().split('T')[0]"
            />
          </div>
        </div>
        
        <!-- Footer -->
        <div class="px-6 py-4 bg-gray-50 rounded-b-lg flex justify-end gap-3">
          <button 
            @click="cancelPaymentModal"
            class="px-4 py-2 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
          >
            Cancel
          </button>
          <button 
            @click="markBillPaid"
            class="px-4 py-2 text-xs font-medium text-white bg-green-600 rounded-md hover:bg-green-700 transition-colors"
          >
            <i class="fas fa-check mr-2"></i>Mark as Paid
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
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

.btn-secondary:focus {
  outline: none;
}
</style>
