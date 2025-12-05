import api from './index';
import type { ExpenseTag } from '../types/ExpenseTag';

// Get all expense tags with spend summaries
export async function getExpenseTags(activeOnly: boolean = false): Promise<ExpenseTag[]> {
  const response = await api.get('/api/expense-tags', {
    params: { active_only: activeOnly }
  });
  return response.data;
}

// Create a new expense tag
export async function createExpenseTag(data: {
  name: string;
  description: string;
  active: boolean;
  budget: number | null; // Budget in cents
}): Promise<ExpenseTag> {
  const response = await api.post('/api/expense-tags', data);
  return response.data;
}

// Update an existing expense tag
export async function updateExpenseTag(
  id: number,
  data: {
    name: string;
    description: string;
    active: boolean;
    budget: number | null; // Budget in cents
  }
): Promise<ExpenseTag> {
  const response = await api.put(`/api/expense-tags/${id}`, data);
  return response.data;
}

// Delete an expense tag
export async function deleteExpenseTag(id: number): Promise<void> {
  await api.delete(`/api/expense-tags/${id}`);
}

