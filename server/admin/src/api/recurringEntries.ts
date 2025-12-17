import api from './index';

export interface RecurringEntry {
  ID: number;
  employee_id: number;
  employee?: any;
  type: string;
  description: string;
  amount: number;
  frequency: string;
  start_date: string;
  end_date?: string;
  is_active: boolean;
  last_generated_date?: string;
  last_generated_for?: string;
}

export async function getRecurringEntries() {
  const response = await api.get('/api/admin/recurring-entries');
  return response.data;
}

export async function createRecurringEntry(data: Partial<RecurringEntry>) {
  const response = await api.post('/api/admin/recurring-entries', data);
  return response.data;
}

export async function updateRecurringEntry(id: number, data: Partial<RecurringEntry>) {
  const response = await api.put(`/api/admin/recurring-entries/${id}`, data);
  return response.data;
}

export async function deleteRecurringEntry(id: number) {
  const response = await api.delete(`/api/admin/recurring-entries/${id}`);
  return response.data;
}

export async function syncAllEmployees() {
  const response = await api.post('/api/admin/recurring-entries/sync');
  return response.data;
}

export async function generateRecurringEntries() {
  const response = await api.post('/api/admin/recurring-entries/generate');
  return response.data;
}

