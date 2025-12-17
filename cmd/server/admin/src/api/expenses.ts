import api from './index';

export interface Expense {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string | null;
  project_id: number;
  submitter_id: number;
  approver_id?: number;
  invoice_id?: number;
  amount: number;
  date: string;
  description: string;
  state: string;
  receipt_id?: number;
  rejection_reason?: string;
  expense_account_code?: string;
  subaccount_code?: string;
  category_id: number;
  is_reimbursable?: boolean;
  project?: any;
  submitter?: any;
  approver?: any;
  invoice?: any;
  receipt?: any;
  category?: any;
  tags?: any[];
}

export async function getExpenses(params?: {
  status?: string;
  project_id?: number;
}): Promise<Expense[]> {
  const response = await api.get('/api/expenses', { params });
  return response.data;
}

export async function createExpense(formData: FormData): Promise<Expense> {
  const response = await api.post('/api/expenses', formData);
  return response.data;
}

export async function updateExpense(id: number, formData: FormData): Promise<Expense> {
  const response = await api.put(`/api/expenses/${id}`, formData);
  return response.data;
}

export async function deleteExpense(id: number): Promise<void> {
  await api.delete(`/api/expenses/${id}`);
}

export async function submitExpense(id: number): Promise<Expense> {
  const response = await api.post(`/api/expenses/${id}/submit`);
  return response.data;
}

export async function approveExpense(id: number): Promise<Expense> {
  const response = await api.post(`/api/expenses/${id}/approve`);
  return response.data;
}

export async function rejectExpense(id: number, reason: string): Promise<Expense> {
  const response = await api.post(`/api/expenses/${id}/reject`, { reason });
  return response.data;
}

export async function refreshReceiptURL(assetId: number): Promise<{ url: string; expires_at: string }> {
  const response = await api.post(`/api/expenses/receipts/${assetId}/refresh-url`);
  return response.data;
}

