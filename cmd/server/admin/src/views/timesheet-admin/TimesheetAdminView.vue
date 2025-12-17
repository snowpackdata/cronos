<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { getDraftInvoices } from '../../api/draftInvoices';
import { type DraftInvoice, type DraftEntry, type DraftAdjustment, formatDate, formatCurrency, getEntryStateClass, getEntryStateDisplayName } from '../../types/DraftInvoice';
import axios from 'axios';

// State for draft invoices
const draftInvoices = ref<DraftInvoice[]>([]);
const selectedInvoiceId = ref<number | null>(null);
const isLoading = ref(true);
const error = ref<string | null>(null);
const editingNotes = ref<Record<number, boolean>>({});
const entryNotes = ref<Record<number, string>>({});

// Inline adjustment state
const inlineAdjustment = ref<{
  invoiceId: number | null;
  type: string;
  amount: number;
  notes: string;
}>({
  invoiceId: null,
  type: 'ADJUSTMENT_TYPE_CREDIT',
  amount: 0,
  notes: '',
});

// Replace modal state with inline editing state
const isAddingAdjustment = ref<Record<number, boolean>>({});

// Get selected invoice
const selectedInvoice = computed(() => {
  if (!selectedInvoiceId.value) return null;
  return draftInvoices.value.find(inv => inv.ID === selectedInvoiceId.value) || null;
});

// Fetch draft invoices on component mount
onMounted(async () => {
  await fetchDraftInvoices();
});

// Fetch all draft invoices
const fetchDraftInvoices = async () => {
  console.log('fetchDraftInvoices called');
  isLoading.value = true;
  error.value = null;
  
  try {
    const invoices = await getDraftInvoices();
    console.log('Fetched invoices:', invoices.length);
    draftInvoices.value = invoices;
    
    // Auto-select first invoice if none selected or if selected invoice no longer exists
    if (draftInvoices.value.length > 0) {
      if (!selectedInvoiceId.value || !draftInvoices.value.find(inv => inv.ID === selectedInvoiceId.value)) {
        console.log('Auto-selecting first invoice:', draftInvoices.value[0].ID);
        selectedInvoiceId.value = draftInvoices.value[0].ID;
      } else {
        console.log('Keeping selected invoice:', selectedInvoiceId.value);
      }
    } else {
      selectedInvoiceId.value = null;
    }
    
    console.log('fetchDraftInvoices complete, selectedInvoiceId:', selectedInvoiceId.value);
  } catch (err) {
    console.error('Error fetching draft invoices:', err);
    error.value = 'Failed to load draft invoices. Please try again.';
  } finally {
    isLoading.value = false;
  }
};

// Start editing entry notes
const handleEditEntryNotes = (entry: DraftEntry) => {
  editingNotes.value[entry.entry_id] = true;
  entryNotes.value[entry.entry_id] = entry.notes;
};

// Save edited entry notes
const handleSaveEntryNotes = async (entry: DraftEntry) => {
  try {
    const formData = new FormData();
    formData.append('notes', entryNotes.value[entry.entry_id]);
    
    await axios.put(`/api/entries/${entry.entry_id}`, formData);
    
    entry.notes = entryNotes.value[entry.entry_id];
    editingNotes.value[entry.entry_id] = false;
    
    await fetchDraftInvoices();
  } catch (err) {
    console.error('Error updating entry notes:', err);
    error.value = 'Failed to save notes. Please try again.';
  }
};

// Cancel editing entry notes
const handleCancelEditNotes = (entryId: number) => {
  editingNotes.value[entryId] = false;
  delete entryNotes.value[entryId];
};

// Handle entry void
const handleVoidEntry = async (entryId: number) => {
  try {
    // Use the correct updated endpoint for voiding entry
    await axios.post(`/api/entries/state/${entryId}/void`);
    await fetchDraftInvoices(); // Refresh data
  } catch (err) {
    console.error('Error voiding entry:', err);
    error.value = 'Failed to void entry. Please try again.';
  }
};

// Handle entry unvoid (restore to draft)
const handleUnvoidEntry = async (entryId: number) => {
  try {
    // Use the same state API pattern to set state back to draft
    await axios.post(`/api/entries/state/${entryId}/draft`);
    await fetchDraftInvoices(); // Refresh data
  } catch (err) {
    console.error('Error unvoiding entry:', err);
    error.value = 'Failed to restore entry. Please try again.';
  }
};

// Handle entry approval
const handleApproveEntry = async (entryId: number) => {
  console.log('handleApproveEntry called with entryId:', entryId);
  try {
    // Use the state API to approve the entry
    console.log('Posting to:', `/api/entries/state/${entryId}/approve`);
    const response = await axios.post(`/api/entries/state/${entryId}/approve`);
    console.log('Approve response:', response);
    await fetchDraftInvoices(); // Refresh data
  } catch (err) {
    console.error('Error approving entry:', err);
    error.value = 'Failed to approve entry. Please try again.';
  }
};

// Handle entry rejection
const handleRejectEntry = async (entryId: number) => {
  console.log('handleRejectEntry called with entryId:', entryId);
  try {
    // Use the state API to reject the entry
    console.log('Posting to:', `/api/entries/state/${entryId}/reject`);
    const response = await axios.post(`/api/entries/state/${entryId}/reject`);
    console.log('Reject response:', response);
    await fetchDraftInvoices(); // Refresh data
  } catch (err) {
    console.error('Error rejecting entry:', err);
    error.value = 'Failed to reject entry. Please try again.';
  }
};

// Handle entry exclusion
const handleExcludeEntry = async (entryId: number) => {
  try {
    // Use the state API to exclude the entry
    await axios.post(`/api/entries/state/${entryId}/exclude`);
    await fetchDraftInvoices(); // Refresh data
  } catch (err) {
    console.error('Error excluding entry:', err);
    error.value = 'Failed to exclude entry. Please try again.';
  }
};

// Handle approve invoice
const handleApproveInvoice = async (invoiceId: number) => {
  try {
    // Use the correct endpoint for approving invoice with POST method
    await axios.post(`/api/invoices/${invoiceId}/approve`);
    await fetchDraftInvoices(); // Refresh data
  } catch (err) {
    console.error('Error approving invoice:', err);
    error.value = 'Failed to approve invoice. Please try again.';
  }
};

// Handle invoice void
const handleVoidInvoice = async (invoiceId: number) => {
  if (confirm('Are you sure you want to void this invoice? This action cannot be undone.')) {
    try {
      // Use the correct endpoint for voiding invoice with POST method
      await axios.post(`/api/invoices/${invoiceId}/void`);
      await fetchDraftInvoices(); // Refresh data
    } catch (err) {
      console.error('Error voiding invoice:', err);
      error.value = 'Failed to void invoice. Please try again.';
    }
  }
};

// New inline adjustment functions
const startAddingAdjustment = (invoiceId: number) => {
  // Reset form and set the current invoice
  inlineAdjustment.value = {
    invoiceId: invoiceId,
    type: 'ADJUSTMENT_TYPE_CREDIT',
    amount: 0,
    notes: '',
  };
  
  // Set this invoice to adjustment mode
  isAddingAdjustment.value[invoiceId] = true;
};

const cancelAddingAdjustment = (invoiceId: number) => {
  isAddingAdjustment.value[invoiceId] = false;
};

// Add adjustment to invoice - modified to use inline data
const handleAddAdjustment = async () => {
  const invoiceId = inlineAdjustment.value.invoiceId;
  if (!invoiceId) return;
  
  // Validate input
  if (inlineAdjustment.value.amount <= 0) {
    error.value = 'Amount must be greater than zero.';
    return;
  }
  
  try {
    error.value = null;
    
    // Create form data for the request
    const formData = new FormData();
    formData.append('invoice_id', invoiceId.toString());
    formData.append('type', inlineAdjustment.value.type);
    
    // Apply visual multiplier for credits
    let amount = inlineAdjustment.value.amount;
    if (inlineAdjustment.value.type === 'ADJUSTMENT_TYPE_CREDIT') {
      amount = -amount;
    }
    formData.append('amount', amount.toString());
    formData.append('notes', inlineAdjustment.value.notes || '');
    
    // Use the correct API endpoint with FormData
    await axios({
      method: 'post',
      url: '/api/adjustments/0',
      data: formData,
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    });
    
    // Close the inline form
    isAddingAdjustment.value[invoiceId] = false;
    
    await fetchDraftInvoices(); // Refresh data
  } catch (error: any) {
    console.error('Error adding adjustment:', error);
    // More detailed error message
    if (error.response) {
      // The server responded with an error
      error.value = `Failed to add adjustment (${error.response.status}): ${error.response.data?.message || 'Server error'}`;
    } else if (error.request) {
      // The request was made but no response was received
      error.value = 'Failed to add adjustment: No response from server. Please try again.';
    } else {
      // Something happened in setting up the request
      error.value = `Failed to add adjustment: ${error.message || 'Unknown error'}`;
    }
  }
};

// Format adjustment type for display
const formatAdjustmentType = (type: string): string => {
  return type.replace('ADJUSTMENT_TYPE_', '');
};

// Get color class for adjustment based on type
const getAdjustmentColorClass = (adjustment: DraftAdjustment): string => {
  if (adjustment.type === 'ADJUSTMENT_TYPE_CREDIT') {
    return 'text-green-600';
  } else {
    return 'text-red-600';
  }
};

// Check if entry is voided
const isVoidedEntry = (entry: DraftEntry): boolean => {
  return entry.state === 'ENTRY_STATE_VOID';
};

// Handle adjustment void
const handleVoidAdjustment = async (adjustmentId: number) => {
  try {
    // Use the correct endpoint for voiding adjustment
    await axios.post(`/api/adjustments/state/${adjustmentId}/void`);
    await fetchDraftInvoices(); // Refresh data
  } catch (err) {
    console.error('Error voiding adjustment:', err);
    error.value = 'Failed to void adjustment. Please try again.';
  }
};

// Handle adjustment unvoid (restore to draft)
const handleUnvoidAdjustment = async (adjustmentId: number) => {
  try {
    // Use the same state API pattern to set state back to draft
    await axios.post(`/api/adjustments/state/${adjustmentId}/draft`);
    await fetchDraftInvoices(); // Refresh data
  } catch (err) {
    console.error('Error unvoiding adjustment:', err);
    error.value = 'Failed to restore adjustment. Please try again.';
  }
};

// Handle adjustment approval
const handleApproveAdjustment = async (adjustmentId: number) => {
  try {
    // Use the state API to approve the adjustment
    await axios.post(`/api/adjustments/state/${adjustmentId}/approve`);
    await fetchDraftInvoices(); // Refresh data
  } catch (err) {
    console.error('Error approving adjustment:', err);
    error.value = 'Failed to approve adjustment. Please try again.';
  }
};

// Check if adjustment is voided
const isVoidedAdjustment = (adjustment: DraftAdjustment): boolean => {
  return adjustment.state === 'ADJUSTMENT_STATE_VOID';
};

// Check if invoice period has ended (period_end is in the past)
const isInvoicePeriodEnded = (invoice: DraftInvoice): boolean => {
  const periodEnd = new Date(invoice.period_end);
  const now = new Date();
  
  // Set times to midnight for accurate date comparison
  periodEnd.setHours(0, 0, 0, 0);
  now.setHours(0, 0, 0, 0);
  
  return periodEnd < now;
};

// Get invoice status badge class based on period end date
const getInvoiceStatusBadgeClass = (invoice: DraftInvoice): string => {
  return isInvoicePeriodEnded(invoice) 
    ? 'bg-green-50 text-green-700' 
    : 'bg-blue-50 text-blue-700';
};

// Check if all entries and adjustments are approved or voided
const canApproveInvoice = (invoice: DraftInvoice | null): boolean => {
  if (!invoice) return false;
  const allEntriesReady = invoice.line_items.every(entry => 
    entry.state === 'ENTRY_STATE_APPROVED' || entry.state === 'ENTRY_STATE_VOID'
  );
  const allAdjustmentsReady = !invoice.adjustments || invoice.adjustments.every(adjustment => 
    adjustment.state === 'ADJUSTMENT_STATE_APPROVED' || adjustment.state === 'ADJUSTMENT_STATE_VOID'
  );
  return allEntriesReady && allAdjustmentsReady;
};

// Check if invoice has unaccepted entries
const hasUnacceptedEntries = (invoice: DraftInvoice): boolean => {
  const hasUnacceptedEntry = invoice.line_items.some(entry => 
    entry.state !== 'ENTRY_STATE_APPROVED' && entry.state !== 'ENTRY_STATE_VOID'
  );
  const hasUnacceptedAdjustment = invoice.adjustments?.some(adjustment => 
    adjustment.state !== 'ADJUSTMENT_STATE_APPROVED' && adjustment.state !== 'ADJUSTMENT_STATE_VOID'
  );
  return hasUnacceptedEntry || hasUnacceptedAdjustment;
};

// Check if invoice is overdue (period ended > 14 days ago)
const isOverdue = (invoice: DraftInvoice): boolean => {
  if (!isInvoicePeriodEnded(invoice)) return false;
  const periodEnd = new Date(invoice.period_end);
  const now = new Date();
  const daysDiff = Math.floor((now.getTime() - periodEnd.getTime()) / (1000 * 60 * 60 * 24));
  return daysDiff > 14;
};

// Categorize invoices into Kanban columns
const needsReviewInvoices = computed(() => {
  return draftInvoices.value.filter(inv => hasUnacceptedEntries(inv));
});

const draftInvoicesColumn = computed(() => {
  return draftInvoices.value.filter(inv => 
    !hasUnacceptedEntries(inv) && 
    canApproveInvoice(inv) && 
    !isInvoicePeriodEnded(inv)
  );
});

const readyInvoices = computed(() => {
  return draftInvoices.value.filter(inv => 
    !hasUnacceptedEntries(inv) && 
    canApproveInvoice(inv) && 
    isInvoicePeriodEnded(inv) && 
    !isOverdue(inv)
  );
});

const overdueInvoices = computed(() => {
  return draftInvoices.value.filter(inv => 
    !hasUnacceptedEntries(inv) && 
    canApproveInvoice(inv) && 
    isOverdue(inv)
  );
});

// Approve all draft entries and adjustments for an invoice
const handleApproveAll = async (invoice: DraftInvoice) => {
  try {
    error.value = null;
    
    // Approve all draft entries
    const draftEntries = invoice.line_items.filter(entry => entry.state === 'ENTRY_STATE_DRAFT');
    for (const entry of draftEntries) {
      await axios.post(`/api/entries/state/${entry.entry_id}/approve`);
    }
    
    // Approve all draft adjustments
    const draftAdjustments = invoice.adjustments?.filter(adj => adj.state === 'ADJUSTMENT_STATE_DRAFT') || [];
    for (const adjustment of draftAdjustments) {
      await axios.post(`/api/adjustments/state/${adjustment.ID}/approve`);
    }
    
    await fetchDraftInvoices(); // Refresh data
  } catch (err) {
    console.error('Error approving all items:', err);
    error.value = 'Failed to approve all items. Please try again.';
  }
};
</script>

<template>
  <div class="px-4 sm:px-6 lg:px-8">
    <!-- Kanban Board -->
    <div v-if="!isLoading && draftInvoices.length > 0" class="mt-4 mb-6">
      <div class="grid grid-cols-4 gap-3">
        <!-- Needs Review Column -->
        <div class="bg-gray-50 rounded-lg p-2 min-h-[120px] flex flex-col">
          <div class="text-xs font-semibold text-gray-700 mb-2 px-1 flex-shrink-0">
            Needs Review ({{ needsReviewInvoices.length }})
          </div>
          <div class="space-y-2 overflow-y-auto flex-1" style="max-height: 180px;">
            <button
              v-for="inv in needsReviewInvoices"
              :key="inv.ID"
              @click="selectedInvoiceId = inv.ID"
              :class="[
                'w-full text-left px-2 py-2 rounded-md text-xs transition-all',
                selectedInvoiceId === inv.ID
                  ? 'bg-red-100 border-2 border-red-400 shadow-md'
                  : 'bg-white border border-gray-200 hover:border-red-300 hover:shadow-sm'
              ]"
            >
              <div class="font-medium text-gray-900 truncate">{{ inv.account_name }}</div>
              <div v-if="inv.project_name" class="text-gray-600 truncate text-[10px] mt-0.5">{{ inv.project_name }}</div>
              <div class="text-gray-500 text-[10px] mt-1">
                {{ formatDate(inv.period_start) }} - {{ formatDate(inv.period_end) }}
              </div>
            </button>
          </div>
        </div>
        
        <!-- Draft Column -->
        <div class="bg-gray-50 rounded-lg p-2 min-h-[120px] flex flex-col">
          <div class="text-xs font-semibold text-gray-700 mb-2 px-1 flex-shrink-0">
            Draft ({{ draftInvoicesColumn.length }})
          </div>
          <div class="space-y-2 overflow-y-auto flex-1" style="max-height: 180px;">
            <button
              v-for="inv in draftInvoicesColumn"
              :key="inv.ID"
              @click="selectedInvoiceId = inv.ID"
              :class="[
                'w-full text-left px-2 py-2 rounded-md text-xs transition-all',
                selectedInvoiceId === inv.ID
                  ? 'bg-blue-100 border-2 border-blue-400 shadow-md'
                  : 'bg-white border border-gray-200 hover:border-blue-300 hover:shadow-sm'
              ]"
            >
              <div class="font-medium text-gray-900 truncate">{{ inv.account_name }}</div>
              <div v-if="inv.project_name" class="text-gray-600 truncate text-[10px] mt-0.5">{{ inv.project_name }}</div>
              <div class="text-gray-500 text-[10px] mt-1">
                {{ formatDate(inv.period_start) }} - {{ formatDate(inv.period_end) }}
              </div>
            </button>
          </div>
        </div>
        
        <!-- Ready Column -->
        <div class="bg-gray-50 rounded-lg p-2 min-h-[120px] flex flex-col">
          <div class="text-xs font-semibold text-gray-700 mb-2 px-1 flex-shrink-0">
            Ready ({{ readyInvoices.length }})
          </div>
          <div class="space-y-2 overflow-y-auto flex-1" style="max-height: 180px;">
            <button
              v-for="inv in readyInvoices"
              :key="inv.ID"
              @click="selectedInvoiceId = inv.ID"
              :class="[
                'w-full text-left px-2 py-2 rounded-md text-xs transition-all',
                selectedInvoiceId === inv.ID
                  ? 'bg-green-100 border-2 border-green-400 shadow-md'
                  : 'bg-white border border-gray-200 hover:border-green-300 hover:shadow-sm'
              ]"
            >
              <div class="font-medium text-gray-900 truncate">{{ inv.account_name }}</div>
              <div v-if="inv.project_name" class="text-gray-600 truncate text-[10px] mt-0.5">{{ inv.project_name }}</div>
              <div class="text-gray-500 text-[10px] mt-1">
                {{ formatDate(inv.period_start) }} - {{ formatDate(inv.period_end) }}
              </div>
            </button>
          </div>
        </div>
        
        <!-- Overdue Column -->
        <div class="bg-gray-50 rounded-lg p-2 min-h-[120px] flex flex-col">
          <div class="text-xs font-semibold text-gray-700 mb-2 px-1 flex-shrink-0">
            Overdue ({{ overdueInvoices.length }})
          </div>
          <div class="space-y-2 overflow-y-auto flex-1" style="max-height: 180px;">
            <button
              v-for="inv in overdueInvoices"
              :key="inv.ID"
              @click="selectedInvoiceId = inv.ID"
              :class="[
                'w-full text-left px-2 py-2 rounded-md text-xs transition-all',
                selectedInvoiceId === inv.ID
                  ? 'bg-orange-100 border-2 border-orange-400 shadow-md'
                  : 'bg-white border border-gray-200 hover:border-orange-300 hover:shadow-sm'
              ]"
            >
              <div class="font-medium text-gray-900 truncate">{{ inv.account_name }}</div>
              <div v-if="inv.project_name" class="text-gray-600 truncate text-[10px] mt-0.5">{{ inv.project_name }}</div>
              <div class="text-gray-500 text-[10px] mt-1">
                {{ formatDate(inv.period_start) }} - {{ formatDate(inv.period_end) }}
              </div>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="isLoading" class="flex flex-col items-center justify-center p-10 bg-white rounded-lg shadow mt-6">
      <i class="fas fa-spinner fa-spin text-3xl text-blue-400 mb-4"></i>
      <span class="text-gray-700">Loading draft invoices...</span>
    </div>
    
    <!-- Error state -->
    <div v-else-if="error" class="flex flex-col items-center justify-center p-10 bg-white rounded-lg shadow mt-6">
      <i class="fas fa-exclamation-circle text-3xl text-red-400 mb-4"></i>
      <span class="text-gray-700 mb-2">{{ error }}</span>
      <button @click="fetchDraftInvoices" class="btn-secondary mt-4">
        <i class="fas fa-sync mr-2"></i> Retry
      </button>
    </div>
    
    <!-- Empty state -->
    <div v-else-if="draftInvoices.length === 0" class="flex flex-col items-center justify-center p-10 bg-white rounded-lg shadow mt-6">
      <i class="fas fa-file-invoice-dollar text-4xl text-blue-400 mb-4"></i>
      <p class="text-base font-medium text-gray-700">No draft invoices found</p>
      <p class="text-sm text-gray-500 mb-4">Draft invoices will appear here once they are generated</p>
    </div>
    
    <!-- Selected invoice details -->
    <div v-else-if="selectedInvoice" class="mt-6 mb-8">
      <!-- Invoice card -->
      <div class="bg-white rounded-lg shadow overflow-hidden">
        <!-- Invoice header -->
        <div class="px-4 py-4 sm:px-6 flex justify-between items-start border-b border-gray-200">
          <div>
            <div class="flex items-center">
              <h3 class="text-base font-medium text-gray-900">{{ selectedInvoice.invoice_name }}</h3>
              <span class="ml-2 px-2 py-0.5 rounded-full text-xs font-medium" :class="getInvoiceStatusBadgeClass(selectedInvoice)">
                {{ isInvoicePeriodEnded(selectedInvoice) ? 'Ready' : 'DRAFT' }}
              </span>
            </div>
            <div class="mt-1 text-xs text-gray-500 space-y-1">
              <div><span class="font-medium">Account:</span> {{ selectedInvoice.account_name }}</div>
              <div v-if="selectedInvoice.project_name"><span class="font-medium">Project:</span> {{ selectedInvoice.project_name }}</div>
              <div><span class="font-medium">Period:</span> {{ formatDate(selectedInvoice.period_start) }} - {{ formatDate(selectedInvoice.period_end) }}</div>
            </div>
          </div>
          <div class="text-right">
            <div class="text-xs text-gray-500">Total Hours</div>
            <div class="text-base font-medium text-gray-900">{{ selectedInvoice.total_hours.toFixed(2) }}</div>
            <div class="text-xs text-gray-500 mt-2">Total Amount</div>
            <div class="text-base font-medium text-gray-900">{{ formatCurrency(selectedInvoice.total_amount) }}</div>
          </div>
        </div>
        
        <!-- Invoice actions -->
        <div class="border-b border-gray-200 bg-gray-50 px-4 py-2 sm:px-6 flex justify-end items-center">
          <div class="flex gap-2">
            <button 
              v-if="!canApproveInvoice(selectedInvoice)"
              @click="handleApproveAll(selectedInvoice)"
              class="inline-flex items-center px-3 py-1 bg-sage-dark text-white text-xs rounded hover:bg-sage transition-colors"
            >
              <i class="fas fa-check-double mr-1"></i> Approve All
            </button>
            <button 
              @click="handleVoidInvoice(selectedInvoice.ID)"
              class="inline-flex items-center px-3 py-1 bg-gray-600 text-white text-xs rounded hover:bg-gray-700 transition-colors"
            >
              <i class="fas fa-ban mr-1"></i> Void
            </button>
            <button 
              v-if="canApproveInvoice(selectedInvoice)"
              @click="handleApproveInvoice(selectedInvoice.ID)"
              class="inline-flex items-center px-3 py-1 bg-sage text-white text-xs rounded hover:bg-sage-dark transition-colors"
            >
              <i class="fas fa-check mr-1"></i> Approve Invoice
            </button>
          </div>
        </div>
        
        <!-- Expandable details section -->
        <div v-if="true" class="divide-y divide-gray-200">
          <!-- Line items section -->
          <div v-if="selectedInvoice.line_items && selectedInvoice.line_items.length > 0" class="px-4 py-3 sm:px-6">
            <h4 class="text-xs font-medium text-gray-700 mb-2">Line Items</h4>
            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200 text-xs">
                <thead>
                  <tr class="bg-gray-50">
                    <th scope="col" class="px-3 py-1.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Staff</th>
                    <th scope="col" class="px-3 py-1.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Billing Code</th>
                    <th scope="col" class="px-3 py-1.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
                    <th scope="col" class="px-3 py-1.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Notes</th>
                    <th scope="col" class="px-3 py-1.5 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Hours</th>
                    <th scope="col" class="px-3 py-1.5 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Fee</th>
                    <th scope="col" class="px-3 py-1.5 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                    <th scope="col" class="px-3 py-1.5 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                  </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">
                  <tr v-for="entry in selectedInvoice.line_items" :key="entry.entry_id" :class="{ 'voided-entry': isVoidedEntry(entry) }">
                    <td class="px-3 py-1.5 whitespace-nowrap">
                      <div class="font-medium text-gray-900">
                        {{ entry.user_name }}
                      </div>
                      <div v-if="entry.is_impersonated" class="text-xs text-gray-500">
                        <span class="italic">impersonated by {{ entry.created_by_name }}</span>
                      </div>
                    </td>
                    <td class="px-3 py-1.5 whitespace-nowrap text-gray-700">{{ entry.billing_code }}</td>
                    <td class="px-3 py-1.5 whitespace-nowrap text-gray-700">{{ formatDate(entry.start_date) }}</td>
                    <td class="px-3 py-1.5 text-gray-700 max-w-md">
                      <div class="flex items-start gap-2">
                        <div class="flex-1">
                          <template v-if="editingNotes[entry.entry_id]">
                            <div class="flex items-start gap-1">
                              <textarea 
                                v-model="entryNotes[entry.entry_id]" 
                                class="flex-1 px-2 py-1 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-sage focus:border-sage"
                                rows="2"
                              ></textarea>
                              <div class="flex gap-1">
                                <button
                                  @click="handleSaveEntryNotes(entry)"
                                  class="text-green-600 hover:text-green-800"
                                  title="Save"
                                >
                                  <i class="fas fa-check text-xs"></i>
                                </button>
                                <button
                                  @click="handleCancelEditNotes(entry.entry_id)"
                                  class="text-gray-500 hover:text-gray-700"
                                  title="Cancel"
                                >
                                  <i class="fas fa-times text-xs"></i>
                                </button>
                              </div>
                            </div>
                          </template>
                          <template v-else>
                            <div class="text-xs break-words whitespace-normal">{{ entry.notes }}</div>
                          </template>
                        </div>
                        <button
                          v-if="!editingNotes[entry.entry_id] && entry.state === 'ENTRY_STATE_DRAFT'"
                          @click="handleEditEntryNotes(entry)"
                          class="text-gray-400 hover:text-sky-700 transition-colors flex-shrink-0"
                          title="Edit"
                        >
                          <i class="fas fa-edit text-xs"></i>
                        </button>
                      </div>
                    </td>
                    <td class="px-3 py-1.5 whitespace-nowrap text-right text-gray-700">{{ entry.duration_hours.toFixed(2) }}</td>
                    <td class="px-3 py-1.5 whitespace-nowrap text-right text-gray-700">{{ formatCurrency(entry.fee) }}</td>
                    <td class="px-3 py-1.5 whitespace-nowrap text-center">
                      <span :class="[getEntryStateClass(entry.state), 'px-1.5 py-0.5 text-xs rounded-full inline-block']">
                        {{ getEntryStateDisplayName(entry.state) }}
                      </span>
                    </td>
                    <td class="px-3 py-1.5 whitespace-nowrap text-right">
                      <template v-if="entry.state === 'ENTRY_STATE_DRAFT'">
                        <div class="flex items-center gap-3">
                          <button
                            @click="handleApproveEntry(entry.entry_id)"
                            class="text-xs text-green-600 hover:text-green-800"
                            title="Approve"
                          >
                            <i class="fas fa-check-circle mr-1"></i>
                          </button>
                          <button
                            @click="handleRejectEntry(entry.entry_id)"
                            class="text-xs text-red-600 hover:text-red-800"
                            title="Reject"
                          >
                            <i class="fas fa-times-circle mr-1"></i>
                          </button>
                        </div>
                      </template>
                      <template v-else-if="entry.state === 'ENTRY_STATE_APPROVED'">
                        <div class="flex items-center gap-3">
                          <button
                            @click="handleExcludeEntry(entry.entry_id)"
                            class="text-xs text-gray-600 hover:text-gray-800"
                            title="Exclude"
                          >
                            <i class="fas fa-eye-slash mr-1"></i>
                          </button>
                          <button
                            @click="handleVoidEntry(entry.entry_id)"
                            class="text-xs text-orange-600 hover:text-orange-800"
                            title="Void"
                          >
                            <i class="fas fa-ban mr-1"></i>
                          </button>
                        </div>
                      </template>
                      <template v-else-if="entry.state === 'ENTRY_STATE_REJECTED' || entry.state === 'ENTRY_STATE_EXCLUDED' || entry.state === 'ENTRY_STATE_VOID'">
                        <button
                          @click="handleUnvoidEntry(entry.entry_id)"
                          class="text-xs text-blue-600 hover:text-blue-800"
                          title="Reset"
                        >
                          <i class="fas fa-undo mr-1"></i>
                        </button>
                      </template>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          
          <!-- Adjustments section -->
          <div class="px-4 py-3 sm:px-6">
            <h4 class="text-xs font-medium text-gray-700 mb-2">Adjustments</h4>
            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200 text-xs">
                <thead>
                  <tr class="bg-gray-50">
                    <th scope="col" class="px-3 py-1.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
                    <th scope="col" class="px-3 py-1.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Notes</th>
                    <th scope="col" class="px-3 py-1.5 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Amount</th>
                    <th scope="col" class="px-3 py-1.5 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                  </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">
                  <!-- Existing adjustments -->
                  <tr v-for="adjustment in selectedInvoice.adjustments" :key="adjustment.ID" :class="{ 'voided-entry': isVoidedAdjustment(adjustment) }">
                    <td class="px-3 py-1.5 whitespace-nowrap font-medium" :class="getAdjustmentColorClass(adjustment)">
                      {{ formatAdjustmentType(adjustment.type) }}
                    </td>
                    <td class="px-3 py-1.5 text-gray-700 max-w-md">
                      <div class="text-xs break-words whitespace-normal">{{ adjustment.notes }}</div>
                    </td>
                    <td class="px-3 py-1.5 whitespace-nowrap text-right font-medium" :class="getAdjustmentColorClass(adjustment)">
                      {{ formatCurrency(adjustment.amount) }}
                    </td>
                    <td class="px-3 py-1.5 whitespace-nowrap text-right">
                      <template v-if="adjustment.state === 'ADJUSTMENT_STATE_DRAFT'">
                        <button
                          @click="handleApproveAdjustment(adjustment.ID)"
                          class="text-xs text-green-600 hover:text-green-800"
                        >
                          <i class="fas fa-check mr-1"></i> Approve
                        </button>
                        <button
                          @click="handleVoidAdjustment(adjustment.ID)"
                          class="text-xs text-gray-500 hover:text-gray-700 ml-2"
                        >
                          <i class="fas fa-ban mr-1"></i> Void
                        </button>
                      </template>
                      <template v-else-if="adjustment.state === 'ADJUSTMENT_STATE_APPROVED'">
                        <button
                          @click="handleUnvoidAdjustment(adjustment.ID)"
                          class="text-xs text-orange-600 hover:text-orange-800"
                        >
                          <i class="fas fa-undo mr-1"></i> Unapprove
                        </button>
                      </template>
                      <template v-else-if="adjustment.state === 'ADJUSTMENT_STATE_VOID'">
                        <button
                          @click="handleUnvoidAdjustment(adjustment.ID)"
                          class="text-xs text-blue-600 hover:text-blue-800"
                        >
                          <i class="fas fa-undo mr-1"></i> Restore
                        </button>
                      </template>
                    </td>
                  </tr>
                  
                  <!-- Inline adjustment form -->
                  <tr v-if="isAddingAdjustment[selectedInvoice.ID]" class="bg-gray-50">
                    <td class="px-3 py-1.5">
                      <select 
                        v-model="inlineAdjustment.type"
                        class="w-full px-2 py-1 text-xxs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
                      >
                        <option value="ADJUSTMENT_TYPE_CREDIT" class="text-xxs">Credit</option>
                        <option value="ADJUSTMENT_TYPE_FEE" class="text-xxs">Fee</option>
                      </select>
                    </td>
                    <td class="px-3 py-1.5">
                      <textarea 
                        v-model="inlineAdjustment.notes"
                        rows="2"
                        placeholder="Enter notes"
                        class="w-full px-2 py-1 text-xxs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
                      ></textarea>
                    </td>
                    <td class="px-3 py-1.5">
                      <input 
                        type="number" 
                        v-model="inlineAdjustment.amount"
                        step="0.01"
                        min="0"
                        placeholder="0.00"
                        class="w-full px-2 py-1 text-xxs border border-gray-300 rounded text-right focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
                      />
                    </td>
                    <td class="px-3 py-1.5 whitespace-nowrap text-right">
                      <button
                        @click="handleAddAdjustment"
                        class="text-xs text-blue-600 hover:text-blue-800"
                      >
                        <i class="fas fa-save mr-1"></i> Save
                      </button>
                      <button
                        @click="cancelAddingAdjustment(selectedInvoice.ID)"
                        class="text-xs text-gray-500 hover:text-gray-700 ml-2"
                      >
                        <i class="fas fa-times mr-1"></i> Cancel
                      </button>
                    </td>
                  </tr>
                  
                  <!-- Empty state with add button -->
                  <tr v-if="!selectedInvoice.adjustments || selectedInvoice.adjustments.length === 0 && !isAddingAdjustment[selectedInvoice.ID]">
                    <td colspan="4" class="px-3 py-4 text-center text-gray-500">
                      <p class="text-xs">No adjustments found</p>
                      <button 
                        @click="startAddingAdjustment(selectedInvoice.ID)"
                        class="mt-2 inline-flex items-center px-2 py-1 text-xs text-blue-600 hover:text-blue-800"
                      >
                        <i class="fas fa-plus-circle mr-1"></i> Add Adjustment
                      </button>
                    </td>
                  </tr>
                </tbody>
                
                <!-- Add a button to add another adjustment if there are already some -->
                <tfoot v-if="(selectedInvoice.adjustments && selectedInvoice.adjustments.length > 0) || isAddingAdjustment[selectedInvoice.ID]">
                  <tr class="bg-gray-50">
                    <td colspan="2" class="px-3 py-1.5 text-right font-medium text-gray-700">Total Adjustments</td>
                    <td class="px-3 py-1.5 text-right font-medium text-gray-900">{{ formatCurrency(selectedInvoice.total_adjustments) }}</td>
                    <td class="px-3 py-1.5 text-right">
                      <button 
                        v-if="!isAddingAdjustment[selectedInvoice.ID]"
                        @click="startAddingAdjustment(selectedInvoice.ID)"
                        class="text-xs text-blue-600 hover:text-blue-800"
                      >
                        <i class="fas fa-plus-circle mr-1"></i> Add
                      </button>
                    </td>
                  </tr>
                </tfoot>
              </table>
            </div>
          </div>
          
          <!-- Expenses section -->
          <div v-if="selectedInvoice.expenses && selectedInvoice.expenses.length > 0" class="px-4 py-3 sm:px-6">
            <h4 class="text-xs font-medium text-gray-700 mb-2">Pass-Through Expenses</h4>
            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200 text-xs">
                <thead>
                  <tr class="bg-gray-50">
                    <th scope="col" class="px-3 py-1.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
                    <th scope="col" class="px-3 py-1.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
                    <th scope="col" class="px-3 py-1.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Submitted By</th>
                    <th scope="col" class="px-3 py-1.5 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Amount</th>
                  </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">
                  <tr v-for="expense in selectedInvoice.expenses" :key="expense.ID">
                    <td class="px-3 py-1.5 whitespace-nowrap text-gray-900">{{ formatDate(expense.date) }}</td>
                    <td class="px-3 py-1.5 text-gray-700 max-w-md">
                      <div class="text-xs break-words whitespace-normal">{{ expense.description }}</div>
                    </td>
                    <td class="px-3 py-1.5 whitespace-nowrap text-gray-700">{{ expense.submitter.first_name }} {{ expense.submitter.last_name }}</td>
                    <td class="px-3 py-1.5 whitespace-nowrap text-right font-medium text-gray-900">{{ formatCurrency(expense.amount) }}</td>
                  </tr>
                </tbody>
                <tfoot>
                  <tr class="bg-gray-50">
                    <td colspan="3" class="px-3 py-1.5 text-right font-medium text-gray-700">Total Expenses</td>
                    <td class="px-3 py-1.5 text-right font-medium text-gray-900">{{ formatCurrency(selectedInvoice.total_expenses) }}</td>
                  </tr>
                </tfoot>
              </table>
            </div>
          </div>
          
          <!-- Invoice summary -->
          <div class="bg-gray-50 px-4 py-3 sm:px-6">
            <div class="grid grid-cols-2 gap-4 text-xs">
              <div>
                <span class="font-medium text-gray-500">Total Hours:</span>
                <span class="ml-2 font-medium text-gray-900">{{ selectedInvoice.total_hours.toFixed(2) }}</span>
              </div>
              <div class="text-right">
                <span class="font-medium text-gray-500">Fees:</span>
                <span class="ml-2 font-medium text-gray-900">{{ formatCurrency(selectedInvoice.total_fees) }}</span>
              </div>
              <div v-if="selectedInvoice.total_expenses > 0">
                <span class="font-medium text-gray-500">Expenses:</span>
                <span class="ml-2 font-medium text-gray-900">{{ formatCurrency(selectedInvoice.total_expenses) }}</span>
              </div>
              <div :class="selectedInvoice.total_expenses > 0 ? 'text-right' : ''">
                <span class="font-medium text-gray-500">Adjustments:</span>
                <span class="ml-2 font-medium text-gray-900">{{ formatCurrency(selectedInvoice.total_adjustments) }}</span>
              </div>
              <div class="col-span-2 border-t border-gray-200 pt-2 mt-2 text-right">
                <span class="font-bold text-gray-700 text-base">Total Amount:</span>
                <span class="ml-2 font-bold text-gray-900 text-base">{{ formatCurrency(selectedInvoice.total_amount) }}</span>
              </div>
            </div>
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
  padding: 0.375rem 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 0.75rem;
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

.voided-entry {
  text-decoration: line-through;
  opacity: 0.6;
}

/* Custom adjustment color classes with more subtle colors */
.text-green-600 {
  color: #059669;
}

.text-red-600 {
  color: #dc2626;
}

/* Custom extra small text size */
.text-xxs {
  font-size: 0.65rem !important;
}
</style> 