import { api } from './apiUtils';
import type { OfflineJournal, OfflineJournalImportResponse } from '../types/OfflineJournal';

const API_BASE = '/api/cronos/offline-journals';

// Upload a CSV file
export async function uploadCSVFile(
  file: File,
  options: {
    dateCol: number;
    descCol: number;
    amountCol: number;
    hasHeader: boolean;
    dateFormat?: string;
  }
): Promise<OfflineJournalImportResponse> {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('date_col', options.dateCol.toString());
  formData.append('desc_col', options.descCol.toString());
  formData.append('amount_col', options.amountCol.toString());
  formData.append('has_header', options.hasHeader.toString());
  if (options.dateFormat) {
    formData.append('date_format', options.dateFormat);
  }

  const response = await api.post(`${API_BASE}/upload-csv`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return response.data;
}

// Get offline journals with optional filters
export async function getOfflineJournals(params?: {
  start_date?: string;
  end_date?: string;
  status?: string;
}): Promise<OfflineJournal[]> {
  const response = await api.get(API_BASE, { params });
  return response.data || [];
}

// Update single offline journal status
export async function updateOfflineJournalStatus(
  id: number,
  status: string,
  notes?: string
): Promise<void> {
  await api.put(`${API_BASE}/${id}`, { status, notes });
}

// Bulk update offline journal statuses
export async function bulkUpdateOfflineJournalStatus(
  ids: number[],
  status: string
): Promise<void> {
  await api.post(`${API_BASE}/bulk-update`, { ids, status });
}

// Get offline journals grouped by transaction
export async function getOfflineJournalTransactions(params?: {
  start_date?: string;
  end_date?: string;
  status?: string;
}): Promise<Record<string, OfflineJournal[]>> {
  const response = await api.get(`${API_BASE}/transactions`, { params });
  return response.data || {};
}

// Categorize a CSV transaction with FROM and TO accounts
export async function categorizeCSVTransaction(data: {
  date: string;
  description: string;
  from_account: string;
  from_subaccount: string;
  to_account: string;
  to_subaccount: string;
}): Promise<void> {
  await api.post(`${API_BASE}/categorize`, data);
}

// Approve and book a transaction pair
export async function approveTransactionPair(data: {
  date: string;
  description: string;
}): Promise<{ message: string; booked: number }> {
  const response = await api.post(`${API_BASE}/approve-transaction`, data);
  return response.data;
}

// Edit offline journal entry
export async function editOfflineJournal(
  id: number,
  data: {
    account: string;
    sub_account: string;
    debit: number;
    credit: number;
  }
): Promise<void> {
  await api.put(`${API_BASE}/${id}/edit`, data);
}

// Delete offline journal entry
export async function deleteOfflineJournal(id: number): Promise<void> {
  await api.delete(`${API_BASE}/${id}`);
}

// Post approved offline journals to the General Ledger
export async function postOfflineJournalsToGL(ids: number[]): Promise<{ message: string; count: string }> {
  const response = await api.post(`${API_BASE}/post-to-gl`, { ids });
  return response.data;
}

