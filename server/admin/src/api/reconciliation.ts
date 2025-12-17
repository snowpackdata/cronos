import api from './index';

const API_BASE = '/api/reconciliation';

export interface ExpenseSearchParams {
  query?: string; // Search in description
  date?: string; // YYYY-MM-DD
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

// Remove reconciliation link
export async function unreconcileTransaction(offlineJournalId: number) {
  const response = await api.post(`${API_BASE}/offline-journals/${offlineJournalId}/unreconcile`);
  return response.data;
}

