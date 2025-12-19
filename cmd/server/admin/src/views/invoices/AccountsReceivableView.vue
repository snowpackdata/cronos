<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { type Invoice } from '../../types/Invoice';
import { getInvoices } from '../../api';

// State
const invoices = ref<Invoice[]>([]);
const isLoading = ref(true);
const error = ref<string | null>(null);
const expandedInvoiceIds = ref<Set<number>>(new Set());
const successMessage = ref<string | null>(null);
const invoiceError = ref<{ [key: number]: string | null }>({});

// Payment date modal
const showPaymentDateModal = ref(false);
const selectedInvoiceForPayment = ref<Invoice | null>(null);
const paymentDate = ref<string>(new Date().toISOString().split('T')[0]);

// Email send modal
const showEmailModal = ref(false);
const selectedInvoiceForEmail = ref<Invoice | null>(null);
const emailTo = ref<string[]>([]);
const emailCC = ref<string[]>([]);
const emailToInput = ref('');
const emailCCInput = ref('');
const emailSubject = ref('');
const emailBody = ref('');
const sendingEmail = ref(false);
const showPreview = ref(false);

// PDF generation state
const generatingPDF = ref<{ [key: number]: boolean }>({});

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
    filtered = filtered.filter(inv => {
      const invDate = getInvoiceDate(inv);
      return invDate && new Date(invDate) >= start;
    });
  }
  if (endDate.value) {
    const end = new Date(endDate.value);
    end.setHours(23, 59, 59, 999); // Include the entire end date
    filtered = filtered.filter(inv => {
      const invDate = getInvoiceDate(inv);
      return invDate && new Date(invDate) <= end;
    });
  }
  
  // Filter by search term
  if (searchTerm.value) {
    const term = searchTerm.value.toLowerCase();
    filtered = filtered.filter(inv => 
      (inv.invoice_number && inv.invoice_number.toLowerCase().includes(term)) ||
      (inv.account_name && inv.account_name.toLowerCase().includes(term)) ||
      (inv.account?.name && inv.account.name.toLowerCase().includes(term))
    );
  }
  
  // Sort by date (most recent first)
  filtered.sort((a, b) => {
    const dateA = new Date(getInvoiceDate(a) || 0);
    const dateB = new Date(getInvoiceDate(b) || 0);
    return dateB.getTime() - dateA.getTime();
  });
  
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

// Get the most relevant date to display for an invoice
const getInvoiceDate = (invoice: Invoice): string => {
  // Priority: sent date > approved date > created date > period start
  if (invoice.date_sent) return invoice.date_sent;
  if (invoice.date_approved) return invoice.date_approved;
  if (invoice.accepted_at) return invoice.accepted_at;
  if (invoice.period_start) return invoice.period_start;
  return invoice.date_created || '';
};

// Format invoice number as YYYYNNNN
const formatInvoiceNumber = (invoice: Invoice): string => {
  let year = new Date().getFullYear();
  if (invoice.accepted_at) {
    year = new Date(invoice.accepted_at).getFullYear();
  } else if (invoice.date_created) {
    year = new Date(invoice.date_created).getFullYear();
  }
  const paddedId = invoice.ID.toString().padStart(4, '0');
  return `${year}${paddedId}`;
};

// Format currency for display
const formatCurrency = (amount: number) => {
  // Amount is already in dollars from the backend
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
};

// Get invoice name based on project or account
const getInvoiceName = (invoice: Invoice): string => {
  // If there's a project, it's a project-based invoice
  if (invoice.project?.name) {
    return invoice.project.name;
  }
  // If there's a description, use that
  if (invoice.description) {
    return invoice.description;
  }
  // Otherwise, it's likely a multi-project account invoice
  const accountName = invoice.account?.name || invoice.account_name;
  return accountName ? `${accountName} Retainer` : 'Account Retainer';
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
const formatState = (state: string | undefined) => {
  if (!state) return 'Unknown';
  return state.replace('INVOICE_STATE_', '').replace(/_/g, ' ');
};

// Format line item type
const formatLineItemType = (type: string | undefined) => {
  if (!type) return 'Unknown';
  return type.replace('LINE_ITEM_TYPE_', '').replace(/_/g, ' ');
};

// Check if invoice is overdue
const isOverdue = (invoice: Invoice): boolean => {
  // Only unpaid/sent invoices can be overdue
  if (invoice.state === 'INVOICE_STATE_PAID' || invoice.state === 'INVOICE_STATE_VOID' || invoice.state === 'INVOICE_STATE_DRAFT') {
    return false;
  }
  
  const dueDate = invoice.date_due || invoice.due_at;
  if (!dueDate || dueDate.startsWith('00') || dueDate === '0001-01-01T00:00:00Z') {
    return false;
  }
  
  const due = new Date(dueDate);
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  
  return due < today;
};

// Get days overdue (negative if not yet due)
const getDaysOverdue = (invoice: Invoice): number => {
  const dueDate = invoice.date_due || invoice.due_at;
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

// Clear all filters
const clearFilters = () => {
  selectedState.value = 'all';
  startDate.value = '';
  endDate.value = '';
  searchTerm.value = '';
};

// Action handlers
// Show payment date modal
const showPaymentModal = (invoice: Invoice) => {
  selectedInvoiceForPayment.value = invoice;
  paymentDate.value = new Date().toISOString().split('T')[0]; // Default to today
  showPaymentDateModal.value = true;
};

// Cancel payment modal
const cancelPaymentModal = () => {
  showPaymentDateModal.value = false;
  selectedInvoiceForPayment.value = null;
};

// Mark invoice as paid with specified payment date
const markInvoicePaid = async () => {
  if (!selectedInvoiceForPayment.value) return;
  
  const invoice = selectedInvoiceForPayment.value;
  invoiceError.value[invoice.ID] = null;
  
  try {
    const token = localStorage.getItem('snowpack_token');
    const response = await fetch(`/api/invoices/${invoice.ID}/paid`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-access-token': token || ''
      },
      body: JSON.stringify({ payment_date: paymentDate.value })
    });
    
    if (!response.ok) {
      throw new Error(`Failed to mark invoice as paid: ${response.statusText}`);
    }
    
    successMessage.value = `Invoice #${formatInvoiceNumber(invoice)} marked as paid on ${paymentDate.value}`;
    showPaymentDateModal.value = false;
    selectedInvoiceForPayment.value = null;
    await fetchInvoices();
    // Success message stays visible until user dismisses it
  } catch (err) {
    console.error('Error marking invoice as paid:', err);
    invoiceError.value[invoice.ID] = err instanceof Error ? err.message : 'Failed to mark invoice as paid';
    showPaymentDateModal.value = false;
  }
};

// Open email modal for sending invoice
const sendInvoice = (invoice: Invoice) => {
  selectedInvoiceForEmail.value = invoice;
  
  // Pre-populate email fields as arrays
  emailTo.value = invoice.account?.email ? [invoice.account.email] : [];
  emailCC.value = ['accounts@snowpack-data.com'];
  emailToInput.value = '';
  emailCCInput.value = '';
  
  // Format subject with month, year, and invoice number
  const periodEndDate = new Date(invoice.period_end);
  const monthYear = periodEndDate.toLocaleDateString('en-US', { month: 'long', year: 'numeric' });
  const invoiceNumber = String(invoice.ID).padStart(6, '0');
  emailSubject.value = `Snowpack Data ${monthYear} Invoice #${invoiceNumber}`;
  
  // Determine if this is project-based or account-based
  const projectOrAccount = invoice.project?.name || invoice.account?.name || invoice.account_name;
  
  // Email template (plain text for editing, will be converted to HTML when sent)
  emailBody.value = `Please see attached invoice for support on ${projectOrAccount} for ${formatDate(invoice.period_start)} - ${formatDate(invoice.period_end)}.

During this period we [Enter accomplishments here]`;
  
  showEmailModal.value = true;
};

// Add email tag functions
const addEmailTo = () => {
  const email = emailToInput.value.trim();
  if (email && !emailTo.value.includes(email)) {
    emailTo.value.push(email);
    emailToInput.value = '';
  }
};

const removeEmailTo = (index: number) => {
  emailTo.value.splice(index, 1);
};

const addEmailCC = () => {
  const email = emailCCInput.value.trim();
  if (email && !emailCC.value.includes(email)) {
    emailCC.value.push(email);
    emailCCInput.value = '';
  }
};

const removeEmailCC = (index: number) => {
  emailCC.value.splice(index, 1);
};

const handleToKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Enter' || e.key === ',' || e.key === ' ') {
    e.preventDefault();
    addEmailTo();
  } else if (e.key === 'Backspace' && !emailToInput.value && emailTo.value.length > 0) {
    removeEmailTo(emailTo.value.length - 1);
  }
};

const handleCCKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Enter' || e.key === ',' || e.key === ' ') {
    e.preventDefault();
    addEmailCC();
  } else if (e.key === 'Backspace' && !emailCCInput.value && emailCC.value.length > 0) {
    removeEmailCC(emailCC.value.length - 1);
  }
};

// Actually send the email
const sendEmailAndInvoice = async () => {
  if (!selectedInvoiceForEmail.value) return;
  
  const invoice = selectedInvoiceForEmail.value;
  const invoiceNumber = formatInvoiceNumber(invoice);
  
  invoiceError.value[invoice.ID] = null;
  sendingEmail.value = true;
  
  // Close modal immediately and show processing message
  showEmailModal.value = false;
  successMessage.value = `Sending invoice #${invoiceNumber}... This may take a moment.`;
  
  try {
    const token = localStorage.getItem('snowpack_token');
    const response = await fetch(`/api/invoices/${invoice.ID}/send_email`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-access-token': token || ''
      },
      body: JSON.stringify({
        to: emailTo.value.join(','),
        cc: emailCC.value.join(','),
        subject: emailSubject.value,
        body: emailBody.value
      })
    });
    
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Failed to send email');
    }
    
    // Success - update message (no auto-dismiss, user can close it)
    successMessage.value = `Invoice #${invoiceNumber} sent successfully!`;
    await fetchInvoices();
  } catch (err) {
    console.error('Error sending invoice email:', err);
    const errorMsg = err instanceof Error ? err.message : 'Failed to send email';
    successMessage.value = null; // Clear the "sending" message
    invoiceError.value[invoice.ID] = `Failed to send invoice #${invoiceNumber}: ${errorMsg}`;
    // Error message stays visible until user dismisses it or page refresh
  } finally {
    sendingEmail.value = false;
  }
};

// Mark invoice as sent offline (without sending email)
const markSentOffline = async () => {
  if (!selectedInvoiceForEmail.value) return;
  
  const invoice = selectedInvoiceForEmail.value;
  invoiceError.value[invoice.ID] = null;
  
  try {
    const token = localStorage.getItem('snowpack_token');
    const response = await fetch(`/api/invoices/${invoice.ID}/send`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-access-token': token || ''
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to mark invoice as sent: ${response.statusText}`);
    }
    
    successMessage.value = `Invoice #${formatInvoiceNumber(invoice)} marked as sent`;
    showEmailModal.value = false;
    await fetchInvoices();
    setTimeout(() => { successMessage.value = null; }, 3000);
  } catch (err) {
    console.error('Error marking invoice as sent:', err);
    invoiceError.value[invoice.ID] = err instanceof Error ? err.message : 'Failed to mark as sent';
  }
};

// Cancel email modal
const closeEmailModal = () => {
  showEmailModal.value = false;
  selectedInvoiceForEmail.value = null;
  emailTo.value = [];
  emailCC.value = [];
  emailToInput.value = '';
  emailCCInput.value = '';
  emailSubject.value = '';
  emailBody.value = '';
};

const regeneratePDF = async (invoice: Invoice) => {
  invoiceError.value[invoice.ID] = null;
  generatingPDF.value[invoice.ID] = true;
  
  try {
    const token = localStorage.getItem('snowpack_token');
    const response = await fetch(`/api/invoices/${invoice.ID}/regenerate_pdf`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-access-token': token || ''
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to regenerate PDF: ${response.statusText}`);
    }
    
    successMessage.value = `Invoice #${formatInvoiceNumber(invoice)} PDF regenerated successfully`;
    await fetchInvoices();
    setTimeout(() => { successMessage.value = null; }, 3000);
  } catch (err) {
    console.error('Error regenerating PDF:', err);
    invoiceError.value[invoice.ID] = err instanceof Error ? err.message : 'Failed to regenerate PDF';
  } finally {
    generatingPDF.value[invoice.ID] = false;
  }
};

const voidInvoice = async (invoice: Invoice) => {
  if (!confirm(`Are you sure you want to void invoice #${formatInvoiceNumber(invoice)}? This action cannot be undone.`)) {
    return;
  }
  
  invoiceError.value[invoice.ID] = null;
  try {
    const token = localStorage.getItem('snowpack_token');
    const response = await fetch(`/api/invoices/${invoice.ID}/void`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-access-token': token || ''
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to void invoice: ${response.statusText}`);
    }
    
    successMessage.value = `Invoice #${formatInvoiceNumber(invoice)} voided`;
    await fetchInvoices();
    setTimeout(() => { successMessage.value = null; }, 3000);
  } catch (err) {
    console.error('Error voiding invoice:', err);
    invoiceError.value[invoice.ID] = err instanceof Error ? err.message : 'Failed to void invoice';
  }
};
</script>

<template>
  <div class="px-4 sm:px-6 lg:px-8">
    <div class="sm:flex sm:items-center">
      <div class="sm:flex-auto">
        <h1 class="text-xl font-semibold text-gray-900">Accounts Receivable</h1>
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
            <option value="all">All</option>
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
    
    <!-- Success Message -->
    <div v-if="successMessage" class="mt-4 bg-green-50 border border-green-200 text-green-800 px-4 py-3 rounded-lg flex items-center justify-between">
      <div class="flex items-center">
        <i class="fas fa-check-circle mr-2"></i>{{ successMessage }}
      </div>
      <button 
        @click="successMessage = null" 
        class="ml-4 text-green-600 hover:text-green-800"
        aria-label="Dismiss"
      >
        <i class="fas fa-times"></i>
      </button>
    </div>
    
    <!-- Invoices Table -->
    <div v-else class="mt-6 bg-white shadow overflow-hidden rounded-lg border border-gray-200">
      <table class="min-w-full divide-y divide-gray-300">
        <thead class="bg-gray-50">
          <tr>
            <th scope="col" class="w-10 py-2 pl-3 pr-2 text-left text-sm font-semibold text-gray-900"></th>
            <th scope="col" class="py-2 px-2 text-left text-sm font-semibold text-gray-900">Invoice</th>
            <th scope="col" class="px-2 py-2 text-left text-sm font-semibold text-gray-900">Client</th>
            <th scope="col" class="px-2 py-2 text-left text-sm font-semibold text-gray-900">Date</th>
            <th scope="col" class="px-2 py-2 text-left text-sm font-semibold text-gray-900">Amount</th>
            <th scope="col" class="px-2 py-2 text-left text-sm font-semibold text-gray-900">Status</th>
                </tr>
              </thead>
        <tbody class="divide-y divide-gray-200 bg-white">
          <template v-for="invoice in filteredInvoices" :key="invoice.ID">
            <!-- Main row -->
            <tr 
              @click="toggleInvoiceExpansion(invoice.ID)"
              class="hover:bg-gray-50 cursor-pointer transition-colors"
            >
              <td class="py-2 pl-3 pr-2 text-sm">
                <i :class="isInvoiceExpanded(invoice.ID) ? 'fas fa-chevron-down' : 'fas fa-chevron-right'" class="text-gray-400"></i>
                  </td>
              <td class="py-2 px-2 text-sm">
                <div class="font-medium text-gray-900">#{{ formatInvoiceNumber(invoice) }}</div>
                <div class="text-xs text-gray-600">{{ getInvoiceName(invoice) }}</div>
                  </td>
              <td class="px-2 py-2 text-sm text-gray-700">{{ invoice.account?.name || invoice.account_name || '-' }}</td>
              <td class="px-2 py-2 text-sm text-gray-700">{{ formatDate(getInvoiceDate(invoice)) }}</td>
              <td class="px-2 py-2 text-sm font-medium text-gray-900">{{ formatCurrency(invoice.total_amount) }}</td>
              <td class="px-2 py-2 text-sm">
                <div class="flex items-center gap-2">
                  <span :class="getStateColorClass(invoice.state)" class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium">
                    {{ formatState(invoice.state) }}
                  </span>
                  <span v-if="isOverdue(invoice)" class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-bold bg-red-100 text-red-800">
                    <i class="fas fa-exclamation-triangle mr-1"></i> {{ getDaysOverdue(invoice) }}d overdue
                  </span>
                </div>
                  </td>
            </tr>

            <!-- Expanded row with line items -->
            <tr v-if="isInvoiceExpanded(invoice.ID)" class="bg-gray-50">
              <td colspan="6" class="px-3 py-3">
                <div class="bg-white rounded-lg border border-gray-200 p-3">
                  <!-- Error Message -->
                  <div v-if="invoiceError[invoice.ID]" class="mb-3 bg-red-50 border border-red-200 text-red-800 px-3 py-2 rounded text-sm flex items-center justify-between">
                    <div class="flex items-center flex-1">
                      <i class="fas fa-exclamation-circle mr-2"></i>{{ invoiceError[invoice.ID] }}
                    </div>
                    <button 
                      @click="invoiceError[invoice.ID] = null" 
                      class="ml-4 text-red-600 hover:text-red-800 flex-shrink-0"
                      aria-label="Dismiss"
                    >
                      <i class="fas fa-times"></i>
                    </button>
                  </div>

                  <div class="flex justify-between items-start mb-3">
                    <div class="flex items-center gap-3">
                      <h4 class="text-sm font-semibold text-gray-900">Invoice Details</h4>
                      <a 
                        v-if="invoice.file"
                        :href="invoice.file"
                        target="_blank"
                        class="text-xs text-blue-600 hover:text-blue-800 flex items-center gap-1"
                      >
                        <i class="fas fa-file-pdf"></i> View PDF
                      </a>
                    </div>
                    
                    <!-- Action Buttons -->
                    <div class="flex gap-2">
                      <!-- Generate PDF button - for APPROVED invoices -->
                      <button 
                        v-if="invoice.state === 'INVOICE_STATE_APPROVED'"
                        @click.stop="regeneratePDF(invoice)"
                        :disabled="generatingPDF[invoice.ID]"
                        class="px-3 py-1 text-xs font-medium text-white bg-purple-600 hover:bg-purple-700 rounded transition-colors disabled:opacity-50"
                        title="Generate invoice PDF"
                      >
                        <i class="fas" :class="generatingPDF[invoice.ID] ? 'fa-spinner fa-spin' : 'fa-file-pdf'"></i> 
                        {{ generatingPDF[invoice.ID] ? 'Generating...' : 'Generate PDF' }}
                      </button>
                      
                      <button 
                        v-if="invoice.state === 'INVOICE_STATE_APPROVED'"
                        @click.stop="sendInvoice(invoice)"
                        class="px-3 py-1 text-xs font-medium text-white bg-blue-600 hover:bg-blue-700 rounded transition-colors"
                      >
                        <i class="fas fa-paper-plane mr-1"></i> Send
                      </button>
                      <button 
                        v-if="invoice.state === 'INVOICE_STATE_SENT'"
                        @click.stop="showPaymentModal(invoice)"
                        class="px-3 py-1 text-xs font-medium text-white bg-green-600 hover:bg-green-700 rounded transition-colors"
                      >
                        <i class="fas fa-check mr-1"></i> Mark Paid
                      </button>
                      <button 
                        v-if="invoice.state === 'INVOICE_STATE_SENT' || invoice.state === 'INVOICE_STATE_PAID'"
                        @click.stop="regeneratePDF(invoice)"
                        class="px-3 py-1 text-xs font-medium text-white bg-purple-600 hover:bg-purple-700 rounded transition-colors"
                        title="Regenerate invoice PDF"
                      >
                        <i class="fas fa-file-pdf mr-1"></i> Regenerate PDF
                      </button>
                      <button 
                        v-if="invoice.state !== 'INVOICE_STATE_VOID' && invoice.state !== 'INVOICE_STATE_PAID'"
                        @click.stop="voidInvoice(invoice)"
                        class="px-3 py-1 text-xs font-medium text-white bg-red-600 hover:bg-red-700 rounded transition-colors"
                      >
                        <i class="fas fa-times mr-1"></i> Void
                      </button>
                    </div>
                  </div>
                  
                  <!-- Invoice metadata -->
                  <div class="grid grid-cols-3 gap-4 mb-3 text-sm">
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
                      <span class="ml-2" :class="isOverdue(invoice) ? 'text-red-700 font-bold' : 'text-gray-900'">
                        {{ formatDate(invoice.date_due || invoice.due_at) }}
                      </span>
                      <span v-if="isOverdue(invoice)" class="ml-2 text-xs font-bold text-red-700">
                        ({{ getDaysOverdue(invoice) }} days overdue)
                      </span>
                    </div>
                  </div>

                  <!-- Line items -->
                  <div v-if="invoice.line_items && invoice.line_items.length > 0">
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
                        <tr v-for="item in invoice.line_items" :key="item.ID">
                          <td class="px-2 py-1.5 text-gray-700">{{ formatLineItemType(item.type) }}</td>
                          <td class="px-2 py-1.5 text-gray-700">{{ item.description }}</td>
                          <td class="px-2 py-1.5 text-right text-gray-700">{{ item.quantity > 0 ? item.quantity.toFixed(2) : '-' }}</td>
                          <td class="px-2 py-1.5 text-right text-gray-700">{{ item.rate > 0 ? formatCurrency(item.rate) : '-' }}</td>
                          <td class="px-2 py-1.5 text-right font-medium text-gray-900">{{ formatCurrency(item.amount / 100) }}</td>
                </tr>
              </tbody>
                      <tfoot class="bg-gray-50 font-semibold">
                        <tr>
                          <td colspan="4" class="px-2 py-1.5 text-right text-gray-900">Total</td>
                          <td class="px-2 py-1.5 text-right text-gray-900">{{ formatCurrency(invoice.total_amount) }}</td>
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
            Enter the date the payment was received for Invoice #{{ selectedInvoiceForPayment ? formatInvoiceNumber(selectedInvoiceForPayment) : '' }}
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
            @click="markInvoicePaid"
            class="px-4 py-2 text-xs font-medium text-white bg-green-600 rounded-md hover:bg-green-700 transition-colors"
          >
            <i class="fas fa-check mr-2"></i>Mark as Paid
          </button>
        </div>
      </div>
    </div>

    <!-- Email Send Modal -->
    <div v-if="showEmailModal" class="fixed inset-0 bg-gray-500/75 flex items-center justify-center z-50" @click="closeEmailModal">
      <div class="bg-white rounded-lg shadow-xl max-w-3xl w-full mx-4 max-h-[92vh] flex flex-col" @click.stop>
        <!-- Header -->
        <div class="bg-gray-900 px-4 py-3 rounded-t-lg">
          <div class="flex justify-between items-center">
            <h3 class="text-sm font-semibold text-white">
              Send Invoice #{{ selectedInvoiceForEmail ? formatInvoiceNumber(selectedInvoiceForEmail) : '' }}
            </h3>
            <button @click="closeEmailModal" class="text-gray-400 hover:text-gray-200 transition-colors">
              <i class="fas fa-times text-lg"></i>
            </button>
          </div>
        </div>
        
        <!-- Body -->
        <div class="px-4 py-3 overflow-y-auto flex-1">
          <!-- To Field with Tags -->
          <div class="mb-2">
            <label class="block text-xs font-medium text-gray-700 mb-0.5">To: <span class="text-red-500">*</span></label>
            <div class="rounded-md bg-white outline-1 -outline-offset-1 outline-gray-300 p-1 focus-within:outline-2 focus-within:-outline-offset-2 focus-within:outline-sage">
              <div class="flex flex-wrap gap-1 items-center">
                <span 
                  v-for="(email, index) in emailTo" 
                  :key="index"
                  class="inline-flex items-center gap-0.5 bg-gray-200 text-gray-700 px-1.5 py-0.5 rounded text-xs"
                >
                  {{ email }}
                  <button @click="removeEmailTo(index)" class="hover:text-gray-900">
                    <i class="fas fa-times" style="font-size: 8px;"></i>
                  </button>
                </span>
                <input 
                  v-model="emailToInput"
                  @keydown="handleToKeydown"
                  @blur="addEmailTo"
                  type="email"
                  placeholder="Add email..."
                  class="flex-1 min-w-24 border-0 p-0 text-xs text-gray-900 placeholder:text-gray-400 focus:outline-none focus:ring-0"
                  style="font-size: 0.75rem; line-height: 1;"
                />
              </div>
            </div>
          </div>

          <!-- CC Field with Tags -->
          <div class="mb-2">
            <label class="block text-xs font-medium text-gray-700 mb-0.5">Cc:</label>
            <div class="rounded-md bg-white outline-1 -outline-offset-1 outline-gray-300 p-1 focus-within:outline-2 focus-within:-outline-offset-2 focus-within:outline-sage">
              <div class="flex flex-wrap gap-1 items-center">
                <span 
                  v-for="(email, index) in emailCC" 
                  :key="index"
                  class="inline-flex items-center gap-0.5 bg-gray-200 text-gray-700 px-1.5 py-0.5 rounded text-xs"
                >
                  {{ email }}
                  <button @click="removeEmailCC(index)" class="hover:text-gray-900">
                    <i class="fas fa-times" style="font-size: 8px;"></i>
                  </button>
                </span>
                <input 
                  v-model="emailCCInput"
                  @keydown="handleCCKeydown"
                  @blur="addEmailCC"
                  type="email"
                  placeholder="Add email..."
                  class="flex-1 min-w-24 border-0 p-0 text-xs text-gray-900 placeholder:text-gray-400 focus:outline-none focus:ring-0"
                  style="font-size: 0.75rem; line-height: 1;"
                />
              </div>
            </div>
          </div>

          <!-- Subject Field -->
          <div class="mb-2">
            <label class="block text-xs font-medium text-gray-700 mb-0.5">Subject: <span class="text-red-500">*</span></label>
            <input 
              v-model="emailSubject" 
              type="text"
              required
              class="block w-full rounded-md bg-white py-1 px-2 text-xs text-gray-900 outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline-2 focus:-outline-offset-2 focus:outline-sage"
              style="font-size: 0.75rem; line-height: 1;"
            />
          </div>

          <!-- Message Body -->
          <div class="mb-2">
            <label class="block text-xs font-medium text-gray-700 mb-0.5">Message: <span class="text-red-500">*</span></label>
            <textarea 
              v-model="emailBody" 
              rows="5" 
              required
              placeholder="Message (fill in accomplishments before sending)"
              class="block w-full rounded-md bg-white px-2 py-1 text-xs text-gray-900 outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline-2 focus:-outline-offset-2 focus:outline-sage"
              style="font-size: 0.75rem; line-height: 1.2;"
            ></textarea>
          </div>

          <!-- Attachment Info -->
          <div class="mb-2 flex items-center text-xs text-gray-600">
            <i class="fas mr-1 size-3" :class="selectedInvoiceForEmail?.file ? 'fa-paperclip text-green-600' : 'fa-file-pdf text-yellow-600'"></i>
            <span>{{ selectedInvoiceForEmail?.file ? 'Invoice PDF attached' : 'PDF will be generated on send' }}</span>
          </div>

          <!-- Email Preview (Collapsible) -->
          <div class="border-t border-gray-200 pt-2">
            <button 
              @click="showPreview = !showPreview"
              class="flex items-center justify-between w-full text-left text-xs font-medium text-gray-700 hover:text-gray-900 mb-1"
            >
              <span>
                <i class="fas fa-eye mr-1 size-3"></i> Email Preview
              </span>
              <i class="fas size-3 transition-transform" :class="showPreview ? 'fa-chevron-up' : 'fa-chevron-down'"></i>
            </button>
            <div v-if="showPreview" class="border border-gray-200 rounded-md p-3 bg-gray-50 max-h-48 overflow-y-auto">
              <!-- Simulated email -->
              <div class="bg-white p-3 rounded shadow-sm">
                <!-- Email Header -->
                <div class="border-b border-sage pb-2 mb-2">
                  <h1 class="text-sm font-bold text-sage">Snowpack Data</h1>
                  <p class="text-xs font-semibold text-gray-600 uppercase tracking-wide mt-0.5">
                    Invoice #{{ selectedInvoiceForEmail ? String(selectedInvoiceForEmail.ID).padStart(6, '0') : '' }}
                  </p>
                </div>
                
                <!-- Email Body -->
                <div class="text-xs text-gray-800 mb-2 whitespace-pre-wrap">{{ emailBody || '[Your message will appear here]' }}</div>
                
                <!-- Invoice Link Button -->
                <div class="text-center my-2">
                  <a href="#" class="inline-block px-3 py-1.5 bg-sage text-white font-semibold rounded text-xs">
                    📄 View Invoice PDF
                  </a>
                </div>
                
                <!-- Email Footer -->
                <div class="mt-3 pt-2 border-t text-xs text-gray-600 bg-gray-50 p-2 rounded">
                  <p class="font-semibold">Best,</p>
                  <p class="font-semibold text-sage">Snowpack Data</p>
                  <p class="text-gray-500 mt-1">
                    2261 Market Street STE 22279<br>
                    San Francisco, CA 94114<br>
                    <a href="mailto:billing@snowpack-data.com" class="text-sage">billing@snowpack-data.com</a>
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="border-t border-gray-200 px-4 py-3 bg-white rounded-b-lg flex justify-between items-center">
          <button
            @click="markSentOffline"
            :disabled="sendingEmail"
            class="text-xs text-gray-600 hover:text-gray-900 font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <i class="fas fa-check-circle mr-1.5 size-3"></i>Mark Sent Offline
          </button>
          <div class="flex gap-2">
            <button
              @click="closeEmailModal"
              class="px-3 py-1.5 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
            >
              Cancel
            </button>
            <button
              @click="sendEmailAndInvoice"
              class="px-3 py-1.5 text-xs font-medium text-white bg-sage rounded-md hover:bg-sage-dark transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="sendingEmail || emailTo.length === 0 || !emailSubject || !emailBody"
            >
              <i class="fas mr-1.5 size-3" :class="sendingEmail ? 'fa-spinner fa-spin' : 'fa-paper-plane'"></i>
              {{ sendingEmail ? 'Sending...' : 'Send Invoice' }}
            </button>
          </div>
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

.bg-sage {
  background-color: #58837e;
}

.bg-sage-dark {
  background-color: #476b67;
}

.bg-sage-light {
  background-color: #6d9d97;
}

.hover\:bg-sage-dark:hover {
  background-color: #476b67;
}

.text-sage {
  color: #58837e;
}

.border-sage {
  border-color: #58837e;
}

.border-sage-dark {
  border-color: #476b67;
}

.focus\:ring-sage:focus {
  --tw-ring-color: #58837e;
}
</style> 
