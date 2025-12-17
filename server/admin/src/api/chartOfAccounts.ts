import api from './index';
import type { ChartOfAccount, ChartOfAccountCreate, ChartOfAccountUpdate } from '../types/ChartOfAccount';

export async function getChartOfAccounts(params?: {
  account_type?: string;
  active_only?: boolean;
}): Promise<ChartOfAccount[]> {
  const response = await api.get('/api/cronos/chart-of-accounts', { params });
  return response.data;
}

export async function createChartOfAccount(data: ChartOfAccountCreate): Promise<ChartOfAccount> {
  const response = await api.post('/api/cronos/chart-of-accounts', data);
  return response.data;
}

export async function updateChartOfAccount(code: string, updates: ChartOfAccountUpdate): Promise<void> {
  await api.put(`/api/cronos/chart-of-accounts/${code}`, updates);
}

export async function deactivateChartOfAccount(code: string): Promise<void> {
  await api.delete(`/api/cronos/chart-of-accounts/${code}`);
}

export async function seedSystemAccounts(): Promise<void> {
  await api.post('/api/cronos/chart-of-accounts/seed');
}

