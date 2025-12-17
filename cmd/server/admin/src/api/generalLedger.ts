import { api } from './apiUtils';

const API_BASE = '/api/cronos/journals';

// Reverse a journal entry with optional corrected entry
export async function reverseJournalEntry(
  id: number,
  data: {
    reason: string;
    create_corrected: boolean;
    corrected?: {
      account: string;
      sub_account: string;
      memo: string;
      debit: number;
      credit: number;
    };
  }
): Promise<void> {
  await api.post(`${API_BASE}/${id}/reverse`, data);
}

