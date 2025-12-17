import api from './index';

const API_BASE = '/api/reconciliation';

export interface ExpenseSearchParams {
  query?: string; // Search in description
  date?: string; // YYYY-MM-DD
  amount?: number; // in cents
}

export interface BillSearchParams {
  query?: string; // Search in employee name or bill name
  date?: string; // YYYY-MM-DD (payment date)
  amount?: number; // in cents
}

// Search for expenses that could match an offline journal transaction
export async function searchExpensesForReconciliation(params: ExpenseSearchParams) {
  const response = await api.get(`${API_BASE}/expenses/search`, { params });
  return response.data;
}

// Reconcile an expense with an offline journal transaction
export async function reconcileExpenseWithOfflineJournal(expenseId: number, offlineJournalId: number) {
  const response = await api.post(`${API_BASE}/expenses/${expenseId}/reconcile`, {
    offline_journal_id: offlineJournalId
  });
  return response.data;
}

// Remove reconciliation link from expense
export async function unreconcileTransaction(offlineJournalId: number) {
  const response = await api.post(`${API_BASE}/offline-journals/${offlineJournalId}/unreconcile`);
  return response.data;
}

// Search for paid bills that could match an offline journal transaction (payroll)
export async function searchBillsForReconciliation(params: BillSearchParams) {
  const response = await api.get(`${API_BASE}/bills/search`, { params });
  return response.data;
}

// Reconcile a bill (payroll) with an offline journal transaction
export async function reconcileBillWithOfflineJournal(billId: number, offlineJournalId: number) {
  const response = await api.post(`${API_BASE}/bills/${billId}/reconcile`, {
    offline_journal_id: offlineJournalId
  });
  return response.data;
}

// Remove reconciliation link from bill
export async function unreconcileBill(billId: number) {
  const response = await api.post(`${API_BASE}/bills/${billId}/unreconcile`);
  return response.data;
}

export interface InvoiceSearchParams {
  query?: string; // Search in account/project/invoice name
  date?: string; // YYYY-MM-DD (payment date)
  amount?: number; // in cents
}

// Search for paid invoices that could match an offline journal transaction (client payment)
export async function searchInvoicesForReconciliation(params: InvoiceSearchParams) {
  const response = await api.get(`${API_BASE}/invoices/search`, { params });
  return response.data;
}

// Reconcile an invoice (client payment) with an offline journal transaction
export async function reconcileInvoiceWithOfflineJournal(invoiceId: number, offlineJournalId: number) {
  const response = await api.post(`${API_BASE}/invoices/${invoiceId}/reconcile`, {
    offline_journal_id: offlineJournalId
  });
  return response.data;
}

// Remove reconciliation link from invoice
export async function unreconcileInvoice(invoiceId: number) {
  const response = await api.post(`${API_BASE}/invoices/${invoiceId}/unreconcile`);
  return response.data;
}

