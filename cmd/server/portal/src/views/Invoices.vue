<template>
  <div class="px-4 sm:px-6 lg:px-8 py-4">
    <div class="sm:flex sm:items-center mb-4">
      <div class="sm:flex-auto">
        <h1 class="text-base font-semibold leading-6 text-gray-900">Invoices</h1>
        <p class="mt-1 text-sm text-gray-700">Outstanding and paid invoices.</p>
      </div>
    </div>

    <div v-if="isLoading" class="mt-6 text-center">
      <p class="text-gray-500 text-sm">Loading invoices...</p>
    </div>
    <div v-else-if="apiError" class="mt-6 rounded-md bg-red-50 p-3">
    <div class="flex">
        <div class="flex-shrink-0">
          <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="ml-2">
          <h3 class="text-sm font-medium text-red-800">Error loading invoices</h3>
          <p class="mt-1 text-xs text-red-700">{{ apiError }}</p>
        </div>
      </div>
    </div>

    <div v-else class="mt-6 flow-root">
      <div v-if="filteredInvoices.length === 0" class="text-center py-8">
        <div class="flex flex-col items-center justify-center p-6 bg-white rounded-lg shadow-sm">
            <svg class="mx-auto h-10 w-10 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0_12.75h.008v.008H8.25V15m0_2.25h.008v.008H8.25V17.25m0_2.25h.008V19.5H8.25V19.5m0_2.25h.008V21.75H8.25V21.75M12_15.375h.008v.008H12V15.375M12_17.625h.008V17.63H12v-.008m0_2.25h.008v.008H12v-.008m0_2.25h.008v.008H12v-.008m-2.25-4.5h.008v.008H9.75V15.375M9.75_17.625h.008V17.63H9.75v-.008m0_2.25h.008v.008H9.75v-.008m0_2.25h.008v.008H9.75v-.008M7.5_19.5h3M7.5_21.75h3m4.5_0h3M12_21.75a.75.75_0_00-.75.75v.008c0 .414.336.75.75.75h.008a.75.75_0_00.75-.75v-.008a.75.75_0_00-.75-.75H12zm0_0H4.5m4.5_0a.75.75_0_00-.75-.75H8.25a.75.75_0_000_1.5h2.25a.75.75_0_00.75-.75V21.75zm0_0h6.75A2.25 2.25 0 0021 19.5V7.5l-6-6H6A2.25 2.25 0 003.75 3.75v15.75c0 .534.213 1.023.568 1.383" />
            </svg>
            <h3 class="mt-2 text-sm font-medium text-gray-900">No invoices found</h3>
            <p class="mt-1 text-xs text-gray-500">You currently have no invoices.</p>
        </div>
      </div>
      <ul v-else role="list" class="space-y-3">
        <li v-for="invoice in filteredInvoices" :key="invoice.ID" class="bg-white shadow-sm overflow-hidden rounded-md border border-gray-200">
          <div class="px-3 py-2.5 sm:px-4">
            
            <div class="flex items-center justify-between mb-1.5">
              <h3 class="text-sm font-semibold leading-snug text-gray-800">Invoice {{ formatInvoiceNumber(invoice) }}</h3>
              <span :class="determineInvoiceStatus(invoice).className" class="ml-2 whitespace-nowrap text-xs px-2 py-0.5">
                {{ determineInvoiceStatus(invoice).text }}
              </span>
            </div>

            <div class="grid grid-cols-5 gap-x-3 mb-2 text-xs">
              <div class="col-span-3">
                <p class="font-medium text-gray-500 truncate">{{ invoice.name || 'N/A' }}</p>
              </div>
              <div class="col-span-2 text-right">
                <p class="text-base font-semibold text-gray-800">{{ formatCurrency(invoice.total_amount) }}</p>
              </div>
            </div>

            <dl class="grid grid-cols-3 gap-x-3 gap-y-0.5 text-xs mb-2 pb-2 border-b border-gray-100">
              <div>
                <dt class="text-gray-500">Issued:</dt>
                <dd class="text-gray-700 font-medium">{{ formatDate(invoice.sent_at) }}</dd>
              </div>
              <div>
                <dt class="text-gray-500">Due:</dt>
                <dd class="text-gray-700 font-medium">{{ formatDate(invoice.due_at) }}</dd>
              </div>
              <div v-if="invoice.state === 'INVOICE_STATE_PAID' && invoice.closed_at">
                <dt class="text-gray-500">Paid:</dt>
                <dd class="text-green-600 font-medium">{{ formatDate(invoice.closed_at) }}</dd>
              </div>
            </dl>

            <dl class="grid grid-cols-3 gap-x-3 gap-y-0.5 text-xs mb-2">
              <div v-if="invoice.total_hours != null">
                <dt class="text-gray-500">Hours:</dt>
                <dd class="text-gray-700">{{ invoice.total_hours.toFixed(1) }}</dd>
              </div>
              <div>
                <dt class="text-gray-500">Fees:</dt>
                <dd class="text-gray-700">{{ formatCurrency(invoice.total_fees) }}</dd>
              </div>
              <div>
                <dt class="text-gray-500">Adjustments:</dt>
                <dd class="text-gray-700">{{ formatCurrency(invoice.total_adjustments) }}</dd>
              </div>
            </dl>

            <!-- Expandable Entries Section -->
            <div v-if="expandedInvoices[invoice.ID]" class="mt-2 pt-2 border-t border-gray-200">
              <h4 class="text-xs font-semibold text-gray-600 mb-1">Line Items:</h4>
              <ul v-if="invoice.entries && invoice.entries.length > 0" class="space-y-1">
                <li v-for="entry in invoice.entries" :key="entry.ID" class="p-2 bg-gray-50 rounded-md">
                  <div class="flex justify-between items-start gap-x-3">
                    <div class="flex-1 flex items-start gap-x-2">
                        <span class="mt-0.5 text-xs text-gray-600 whitespace-nowrap">{{ formatDate(entry.start, true) }}</span>
                        <span class="text-xs text-gray-800 break-words" :title="entry.notes">{{ entry.notes || 'N/A' }}</span>
                    </div>
                    <div class="flex-shrink-0 text-right">
                        <span class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold bg-indigo-100 text-indigo-800 whitespace-nowrap">
                          {{ formatEntryDuration(entry.start, entry.end) }}
                        </span>
                    </div>
                  </div>
                </li>
              </ul>
              <p v-else class="text-xs text-gray-500 italic">No line items for this invoice.</p>
            </div>

            <!-- Toggle and PDF Button Section -->
            <div class="mt-3 space-y-2">
                <button @click="toggleEntries(invoice.ID)" 
                        class="w-full flex items-center justify-center px-2.5 py-1 border border-gray-300 rounded text-xs font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-1 focus:ring-indigo-500">
                  <span v-if="expandedInvoices[invoice.ID]">Hide Details</span>
                  <span v-else>View Details</span>
                  <svg v-if="expandedInvoices[invoice.ID]" xmlns="http://www.w3.org/2000/svg" class="ml-1 h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                  </svg>
                  <svg v-else xmlns="http://www.w3.org/2000/svg" class="ml-1 h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
                  </svg>
                </button>
                <a v-if="invoice.file" :href="invoice.file" target="_blank" download
                    class="w-full flex items-center justify-center px-2.5 py-1 border border-transparent rounded text-xs font-medium text-white bg-indigo-500 hover:bg-indigo-600 focus:outline-none focus:ring-2 focus:ring-offset-1 focus:ring-indigo-400">
                    Download PDF
                    <svg class="ml-1 h-3.5 w-3.5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                    <path fill-rule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" clip-rule="evenodd" />
                    </svg>
                </a>
            </div>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { portalAPI } from '../api'; 

interface InvoiceProject {
  ID: number;
  name?: string;
}

interface PortalEntry {
  ID: number;
  notes?: string;
  start: string | Date;
  end?: string | Date;
  fee_cents: number;
}

interface PortalInvoice {
  ID: number;
  name?: string;
  state: string;
  created_at?: string | Date;
  sent_at?: string | Date; 
  due_at?: string | Date;
  closed_at?: string | Date;
  accepted_at?: string | Date; 
  total_hours?: number;
  total_fees: number;
  total_adjustments: number;
  total_amount: number;
  Project?: InvoiceProject;
  file?: string;
  entries?: PortalEntry[];
}

const allInvoices = ref<PortalInvoice[]>([]);
const isLoading = ref(true);
const apiError = ref<string | null>(null);
const expandedInvoices = ref<Record<number, boolean>>({});

const filteredInvoices = computed(() => {
  return allInvoices.value.filter(invoice => invoice.sent_at && invoice.sent_at !== '0001-01-01T00:00:00Z');
});

const toggleEntries = (invoiceId: number) => {
  expandedInvoices.value[invoiceId] = !expandedInvoices.value[invoiceId];
};

onMounted(async () => {
  isLoading.value = true;
  apiError.value = null;
  try {
    const data = await portalAPI.fetchPortalAcceptedInvoices();
    allInvoices.value = (data || []).map((invoice: PortalInvoice) => ({
      ...invoice,
      entries: invoice.entries?.sort((a: PortalEntry, b: PortalEntry) => {
        const dateA = new Date(a.start).getTime();
        const dateB = new Date(b.start).getTime();
        return dateB - dateA; // Sort descending
      }).map((entry: PortalEntry) => ({
        ...entry,
      }))
    }));
  } catch (error: any) {
    console.error('Error fetching accepted invoices:', error);
    apiError.value = error?.response?.data?.error || error?.message || 'An unknown error occurred.';
    allInvoices.value = [];
  } finally {
    isLoading.value = false;
  }
});

const formatDate = (dateString?: string | Date | null, shortMonth: boolean = false): string => {
  if (!dateString) return '-';
  if (typeof dateString === 'string' && (dateString.startsWith('0001-01-01') || dateString.startsWith('0000-00-00'))) return '-';
  const date = new Date(dateString);
  if (isNaN(date.getTime())) return '-';
  const options: Intl.DateTimeFormatOptions = { year: 'numeric', month: shortMonth ? 'short' : 'short', day: 'numeric' };
  return date.toLocaleDateString(undefined, options);
};

const formatEntryDuration = (start?: string | Date, end?: string | Date): string => {
  if (!start || !end) return '- hrs';

  const startDate = new Date(start);
  const endDate = new Date(end);

  if (isNaN(startDate.getTime()) || isNaN(endDate.getTime())) return '- hrs';

  let diffMs = endDate.getTime() - startDate.getTime();

  if (diffMs < 0) diffMs = 0;

  const hours = diffMs / (1000 * 60 * 60);
  
  if (hours < 0.01 && hours > 0) {
      return '0.01 hrs';
  } 
  if (hours === 0) {
      return '0.00 hrs';
  }
  
  return hours.toFixed(2) + ' hrs';
};

const formatCurrency = (amount?: number): string => {
  if (typeof amount !== 'number' || isNaN(amount)) return '-';
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(amount);
};

const formatInvoiceNumber = (invoice: PortalInvoice): string => {
  let year = new Date().getFullYear(); 
  const dateToUse = invoice.accepted_at || invoice.sent_at || invoice.created_at;
  if (dateToUse && typeof dateToUse === 'string' && !dateToUse.startsWith('0001-01-01')) {
    year = new Date(dateToUse).getFullYear();
  }
  const paddedId = invoice.ID.toString().padStart(4, '0');
  return `${year}${paddedId}`;
};

interface InvoiceStatusInfo {
  text: string;
  className: string[];
}

const determineInvoiceStatus = (invoice: PortalInvoice): InvoiceStatusInfo => {
  const baseClasses = ['inline-flex', 'items-center', 'rounded-full', 'px-2', 'py-0.5', 'text-xs', 'font-medium'];
  
  if (invoice.state === 'INVOICE_STATE_PAID') {
    return { text: 'Paid', className: [...baseClasses, 'bg-green-100', 'text-green-700'] };
  }

  if (invoice.due_at) {
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const dueDate = new Date(invoice.due_at);
    dueDate.setHours(0, 0, 0, 0);

    if (dueDate < today) {
      return { text: 'Overdue', className: [...baseClasses, 'bg-red-100', 'text-red-700'] };
    }
  }
  return { text: 'Outstanding', className: [...baseClasses, 'bg-yellow-100', 'text-yellow-700'] };
};

</script>

<style scoped>
/* Scoped styles for the Invoices view if needed */
</style> 