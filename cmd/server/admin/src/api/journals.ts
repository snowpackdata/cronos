import { api } from './apiUtils';
import type { Journal, BalanceSummary, LedgerEntry, ReconciliationReport } from '../types/Journal';

// Get list of journal entries with optional filters
export async function getJournals(filters?: {
  start_date?: string;
  end_date?: string;
  account?: string;
  invoice_id?: number;
  bill_id?: number;
  sub_account?: string;
  include_offline?: boolean;
}): Promise<Journal[]> {
  const params = new URLSearchParams();
  
  if (filters?.start_date) {
    params.append('start_date', filters.start_date);
  }
  if (filters?.end_date) {
    params.append('end_date', filters.end_date);
  }
  if (filters?.account) {
    params.append('account', filters.account);
  }
  if (filters?.invoice_id) {
    params.append('invoice_id', filters.invoice_id.toString());
  }
  if (filters?.bill_id) {
    params.append('bill_id', filters.bill_id.toString());
  }
  if (filters?.sub_account) {
    params.append('sub_account', filters.sub_account);
  }
  if (filters?.include_offline) {
    params.append('include_offline', 'true');
  }

  const response = await api.get<Journal[]>(
    `/api/cronos/journals?${params.toString()}`
  );
  return response.data;
}

// Get combined general ledger (Journal DB + Beancount)
export async function getCombinedLedger(filters?: {
  start_date?: string;
  end_date?: string;
}): Promise<LedgerEntry[]> {
  const params = new URLSearchParams();
  
  if (filters?.start_date) {
    params.append('start_date', filters.start_date);
  }
  if (filters?.end_date) {
    params.append('end_date', filters.end_date);
  }

  const response = await api.get<LedgerEntry[]>(
    `/api/cronos/ledger/combined?${params.toString()}`
  );
  return response.data;
}

// Get reconciliation report
export async function getReconciliationReport(asOfDate?: string): Promise<ReconciliationReport> {
  const params = new URLSearchParams();
  
  if (asOfDate) {
    params.append('as_of_date', asOfDate);
  }

  const response = await api.get<ReconciliationReport>(
    `/api/cronos/ledger/reconciliation?${params.toString()}`
  );
  return response.data;
}

// Get account balances summary with date range filter
export async function getAccountBalances(startDate?: string, endDate?: string): Promise<BalanceSummary> {
  const params = new URLSearchParams();
  
  if (startDate) {
    params.append('start_date', startDate);
  }
  if (endDate) {
    params.append('end_date', endDate);
  }

  const response = await api.get<BalanceSummary>(
    `/api/cronos/accounts/balances?${params.toString()}`
  );
  return response.data;
}

