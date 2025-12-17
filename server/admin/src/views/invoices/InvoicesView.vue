<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { type Invoice } from '../../types/Invoice';
import { getInvoices } from '../../api';

// State
const invoices = ref<Invoice[]>([]);
const isLoading = ref(true);
const error = ref<string | null>(null);
const expandedInvoiceIds = ref<Set<number>>(new Set());

// Filters
const selectedState = ref<string>('all');
const startDate = ref('');
const endDate = ref('');
const searchTerm = ref('');

// Fetch invoices on component mount
onMounted(async () => {
  await fetchInvoices();
});

// Fetch all invoices
const fetchInvoices = async () => {
  isLoading.value = true;
  error.value = null;
  
  try {
    invoices.value = await getInvoices();
  } catch (err) {
    console.error('Error fetching invoices:', err);
    error.value = 'Failed to load invoices. Please try again.';
  } finally {
    isLoading.value = false;
  }
};

// Filter invoices based on selected criteria
const filteredInvoices = computed(() => {
  let filtered = invoices.value;
  
  // Filter by state
  if (selectedState.value !== 'all') {
    filtered = filtered.filter(inv => inv.state === selectedState.value);
  }
  
  // Filter by date range
  if (startDate.value) {
    const start = new Date(startDate.value);
    filtered = filtered.filter(inv => new Date(inv.date_created) >= start);
  }
  if (endDate.value) {
    const end = new Date(endDate.value);
    filtered = filtered.filter(inv => new Date(inv.date_created) <= end);
  }
  
  // Filter by search term
  if (searchTerm.value) {
    const term = searchTerm.value.toLowerCase();
    filtered = filtered.filter(inv => 
      inv.invoice_number.toLowerCase().includes(term) ||
      inv.account_name.toLowerCase().includes(term)
    );
  }
  
  return filtered;
});

// Toggle invoice expansion
const toggleInvoiceExpansion = (invoiceId: number) => {
  if (expandedInvoiceIds.value.has(invoiceId)) {
    expandedInvoiceIds.value.delete(invoiceId);
  } else {
    expandedInvoiceIds.value.add(invoiceId);
  }
};

// Check if invoice is expanded
const isInvoiceExpanded = (invoiceId: number) => {
  return expandedInvoiceIds.value.has(invoiceId);
};

// Format date for display
const formatDate = (dateString: string): string => {
  if (!dateString) return 'N/A';
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
};

// Format currency for display
const formatCurrency = (amount: number) => {
  // Amount is already in dollars from the backend
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
};

// Get state color class
const getStateColorClass = (state: string) => {
  switch (state) {
    case 'INVOICE_STATE_DRAFT':
      return 'bg-gray-100 text-gray-700';
    case 'INVOICE_STATE_APPROVED':
      return 'bg-blue-100 text-blue-700';
    case 'INVOICE_STATE_SENT':
      return 'bg-yellow-100 text-yellow-700';
    case 'INVOICE_STATE_PAID':
      return 'bg-green-100 text-green-700';
    case 'INVOICE_STATE_VOID':
      return 'bg-red-100 text-red-700';
    default:
      return 'bg-gray-100 text-gray-700';
  }
};

// Format state for display
const formatState = (state: string) => {
  return state.replace('INVOICE_STATE_', '').replace(/_/g, ' ');
};

// Format line item type
const formatLineItemType = (type: string) => {
  return type.replace('LINE_ITEM_TYPE_', '').replace(/_/g, ' ');
};

// Clear all filters
const clearFilters = () => {
  selectedState.value = 'all';
  startDate.value = '';
  endDate.value = '';
  searchTerm.value = '';
};
</script>

<template>
  <div class="px-4 sm:px-6 lg:px-8">
    <div class="sm:flex sm:items-center">
      <div class="sm:flex-auto">
        <h1 class="text-xl font-semibold text-gray-900">Invoices</h1>
        <p class="mt-2 text-sm text-gray-700">A list of all invoices including their status, amount, and line item details.</p>
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
            placeholder="Invoice # or Client..."
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
            <option value="all">All Statuses</option>
            <option value="INVOICE_STATE_DRAFT">Draft</option>
            <option value="INVOICE_STATE_APPROVED">Approved</option>
            <option value="INVOICE_STATE_SENT">Sent</option>
            <option value="INVOICE_STATE_PAID">Paid</option>
            <option value="INVOICE_STATE_VOID">Void</option>
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
      <span class="text-gray-600">Loading invoices...</span>
    </div>
    
    <!-- Error state -->
    <div v-else-if="error" class="flex flex-col items-center justify-center p-10 bg-white rounded-lg shadow mt-6">
      <i class="fas fa-exclamation-circle text-4xl text-red-600 mb-4"></i>
      <span class="text-gray-600 mb-2">{{ error }}</span>
      <button @click="fetchInvoices" class="btn-secondary mt-4">
        <i class="fas fa-sync mr-2"></i> Retry
      </button>
    </div>
    
    <!-- Empty state -->
    <div v-else-if="filteredInvoices.length === 0" class="flex flex-col items-center justify-center p-10 bg-white rounded-lg shadow mt-6">
      <i class="fas fa-file-invoice-dollar text-5xl text-blue-600 mb-4"></i>
      <p class="text-lg font-medium text-gray-700">No invoices found</p>
      <p class="text-gray-600 mb-4">{{ invoices.length === 0 ? 'Invoices will appear here once they are created' : 'Try adjusting your filters' }}</p>
      <button v-if="invoices.length > 0" @click="clearFilters" class="btn-secondary">
        Clear Filters
      </button>
    </div>
    
    <!-- Invoices Table -->
    <div v-else class="mt-6 bg-white shadow overflow-hidden rounded-lg border border-gray-200">
      <table class="min-w-full divide-y divide-gray-300">
        <thead class="bg-gray-50">
          <tr>
            <th scope="col" class="w-12 py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900"></th>
            <th scope="col" class="py-3.5 px-3 text-left text-sm font-semibold text-gray-900">Invoice #</th>
            <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Client</th>
            <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Date</th>
            <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Amount</th>
            <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Status</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 bg-white">
          <template v-for="invoice in filteredInvoices" :key="invoice.ID">
            <!-- Main row -->
            <tr 
              @click="toggleInvoiceExpansion(invoice.ID)"
              class="hover:bg-gray-50 cursor-pointer transition-colors"
            >
              <td class="py-4 pl-4 pr-3 text-sm">
                <i :class="isInvoiceExpanded(invoice.ID) ? 'fas fa-chevron-down' : 'fas fa-chevron-right'" class="text-gray-400"></i>
              </td>
              <td class="py-4 px-3 text-sm font-medium text-gray-900">{{ invoice.invoice_number }}</td>
              <td class="px-3 py-4 text-sm text-gray-700">{{ invoice.account_name }}</td>
              <td class="px-3 py-4 text-sm text-gray-700">{{ formatDate(invoice.date_created) }}</td>
              <td class="px-3 py-4 text-sm font-medium text-gray-900">{{ formatCurrency(invoice.total_amount) }}</td>
              <td class="px-3 py-4 text-sm">
                <span :class="getStateColorClass(invoice.state)" class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium">
                  {{ formatState(invoice.state) }}
                </span>
              </td>
            </tr>

            <!-- Expanded row with line items -->
            <tr v-if="isInvoiceExpanded(invoice.ID)" class="bg-gray-50">
              <td colspan="6" class="px-4 py-4">
                <div class="bg-white rounded-lg border border-gray-200 p-4">
                  <h4 class="text-sm font-semibold text-gray-900 mb-3">Invoice Details</h4>
                  
                  <!-- Invoice metadata -->
                  <div class="grid grid-cols-3 gap-4 mb-4 text-sm">
                    <div>
                      <span class="text-gray-600">Period:</span>
                      <span class="ml-2 text-gray-900">{{ formatDate(invoice.period_start) }} - {{ formatDate(invoice.period_end) }}</span>
                    </div>
                    <div>
                      <span class="text-gray-600">Hours:</span>
                      <span class="ml-2 text-gray-900">{{ invoice.total_hours.toFixed(2) }}</span>
                    </div>
                    <div>
                      <span class="text-gray-600">Due Date:</span>
                      <span class="ml-2 text-gray-900">{{ formatDate(invoice.date_due) }}</span>
                    </div>
                  </div>

                  <!-- Line items -->
                  <div v-if="invoice.line_items && invoice.line_items.length > 0">
                    <h5 class="text-sm font-medium text-gray-900 mb-2">Line Items</h5>
                    <table class="min-w-full text-sm">
                      <thead class="bg-gray-100">
                        <tr>
                          <th class="px-3 py-2 text-left text-xs font-medium text-gray-600">Type</th>
                          <th class="px-3 py-2 text-left text-xs font-medium text-gray-600">Description</th>
                          <th class="px-3 py-2 text-right text-xs font-medium text-gray-600">Quantity</th>
                          <th class="px-3 py-2 text-right text-xs font-medium text-gray-600">Rate</th>
                          <th class="px-3 py-2 text-right text-xs font-medium text-gray-600">Amount</th>
                        </tr>
                      </thead>
                      <tbody class="divide-y divide-gray-200">
                        <tr v-for="item in invoice.line_items" :key="item.ID">
                          <td class="px-3 py-2 text-gray-700">{{ formatLineItemType(item.type) }}</td>
                          <td class="px-3 py-2 text-gray-700">{{ item.description }}</td>
                          <td class="px-3 py-2 text-right text-gray-700">{{ item.quantity > 0 ? item.quantity.toFixed(2) : '-' }}</td>
                          <td class="px-3 py-2 text-right text-gray-700">{{ item.rate > 0 ? formatCurrency(item.rate) : '-' }}</td>
                          <td class="px-3 py-2 text-right font-medium text-gray-900">{{ formatCurrency(item.amount / 100) }}</td>
                        </tr>
                      </tbody>
                      <tfoot class="bg-gray-50 font-semibold">
                        <tr>
                          <td colspan="4" class="px-3 py-2 text-right text-gray-900">Total</td>
                          <td class="px-3 py-2 text-right text-gray-900">{{ formatCurrency(invoice.total_amount) }}</td>
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
