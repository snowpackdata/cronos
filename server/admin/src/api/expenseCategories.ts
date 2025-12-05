import api from './index';
import type { ExpenseCategory } from '../types/ExpenseCategory';

// Get all expense categories
export async function getExpenseCategories(activeOnly: boolean = false): Promise<ExpenseCategory[]> {
  const response = await api.get('/api/expense-categories', {
    params: { active_only: activeOnly }
  });
  return response.data;
}

// Create a new expense category
export async function createExpenseCategory(data: {
  name: string;
  description: string;
}): Promise<ExpenseCategory> {
  const response = await api.post('/api/expense-categories', data);
  return response.data;
}

// Update an existing expense category
export async function updateExpenseCategory(
  id: number,
  data: {
    name: string;
    description: string;
    active: boolean;
  }
): Promise<ExpenseCategory> {
  const response = await api.put(`/api/expense-categories/${id}`, data);
  return response.data;
}

// Delete an expense category
export async function deleteExpenseCategory(id: number): Promise<void> {
  await api.delete(`/api/expense-categories/${id}`);
}

