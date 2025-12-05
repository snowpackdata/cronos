export interface OfflineJournal {
  ID: number; // GORM uses uppercase ID
  date: string;
  account: string;
  sub_account: string;
  description: string;
  debit: number; // in cents
  credit: number; // in cents
  content_hash: string;
  source: string;
  status: 'pending_review' | 'approved' | 'duplicate' | 'excluded' | 'posted';
  imported_at: string;
  reviewed_at?: string;
  reviewed_by?: number;
  notes?: string;
  reconciled_expense_id?: number;
  reconciled_at?: string;
  reconciled_by?: number;
  CreatedAt: string; // GORM uses uppercase
  UpdatedAt: string; // GORM uses uppercase
  DeletedAt?: string | null; // GORM uses uppercase
}

export interface OfflineJournalImportResponse {
  imported: number;
  skipped: number;
  message: string;
}

export function formatCurrency(cents: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(cents / 100);
}

export function getStatusColor(status: string): string {
  switch (status) {
    case 'pending_review':
      return 'bg-yellow-100 text-yellow-800';
    case 'approved':
      return 'bg-green-100 text-green-800';
    case 'duplicate':
      return 'bg-red-100 text-red-800';
    case 'excluded':
      return 'bg-gray-100 text-gray-800';
    case 'posted':
      return 'bg-blue-100 text-blue-800';
    default:
      return 'bg-gray-100 text-gray-800';
  }
}

