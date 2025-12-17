import api from './index';
import type { Subaccount, SubaccountCreate, SubaccountUpdate } from '../types/Subaccount';

export async function getSubaccounts(params?: {
  account_code?: string;
  type?: string;
  active_only?: boolean;
}): Promise<Subaccount[]> {
  const response = await api.get('/api/cronos/subaccounts', { params });
  return response.data;
}

export async function createSubaccount(data: SubaccountCreate): Promise<Subaccount> {
  const response = await api.post('/api/cronos/subaccounts', data);
  return response.data;
}

export async function updateSubaccount(code: string, updates: SubaccountUpdate): Promise<void> {
  await api.put(`/api/cronos/subaccounts/${code}`, updates);
}

export async function deactivateSubaccount(code: string): Promise<void> {
  await api.delete(`/api/cronos/subaccounts/${code}`);
}

