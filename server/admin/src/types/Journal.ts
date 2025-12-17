// Journal entry types for accounting system

export interface Journal {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string | null;
  account: string;
  sub_account: string;
  invoice_id: number | null;
  bill_id: number | null;
  memo: string;
  debit: number;
  credit: number;
}

export interface AccountBalance {
  account: string;
  total_debits: number;
  total_credits: number;
  net_balance: number;
}

export interface BalanceSummary {
  accounts: AccountBalance[];
  total_debits: number;
  total_credits: number;
  net_balance: number;
  is_balanced: boolean;
}

// Combined ledger entry (from Journal DB or Beancount)
export interface LedgerEntry {
  date: string;
  account: string;
  sub_account: string;
  description: string;
  debit: number;
  credit: number;
  source: 'beancount' | 'journal_db';
  invoice_id?: number;
  bill_id?: number;
  tags?: string[];
}

// Reconciliation types
export interface PotentialDuplicate {
  beancount_entry: LedgerEntry;
  journal_entry: LedgerEntry;
  confidence: 'high' | 'medium' | 'low';
}

export interface ReconciliationReport {
  cash_balance_beancount: number;
  cash_balance_journal_db: number;
  difference: number;
  potential_duplicates?: PotentialDuplicate[];
  as_of_date: string;
}

// Format cents to dollars
export function formatCurrency(cents: number): string {
  const dollars = cents / 100;
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(dollars);
}

// Format account name for display
export function formatAccountName(account: string): string {
  return account
    .split('_')
    .map(word => word.charAt(0) + word.slice(1).toLowerCase())
    .join(' ');
}

// Get account category for grouping
export function getAccountCategory(account: string): string {
  // Check expenses first (including EQUIPMENT_EXPENSE, DISTRIBUTIONS)
  if (account.includes('EXPENSE') || account.includes('OPERATING_EXPENSES') || account.includes('DISTRIBUTIONS')) {
    return 'Expenses';
  }
  // Then check assets (EQUIPMENT without _EXPENSE suffix, CASH, etc)
  if (account.includes('CASH') || account.includes('RECEIVABLE') || account.includes('ACCRUED_RECEIVABLES') || 
      account === 'EQUIPMENT' || account.includes('OTHER_ASSETS')) {
    return 'Assets';
  }
  if (account.includes('PAYABLE') || account.includes('ACCRUED_PAYROLL') || account.includes('CREDIT_CARD') || 
      account.includes('OTHER_LIABILITIES')) {
    return 'Liabilities';
  }
  if (account.includes('REVENUE') || account.includes('CREDITS') || account.includes('OTHER_INCOME')) {
    return 'Revenue';
  }
  if (account.includes('EQUITY') || account.includes('OWNERSHIP')) {
    return 'Equity';
  }
  return 'Other';
}

// Get account type color
export function getAccountColor(account: string): string {
  const category = getAccountCategory(account);
  switch (category) {
    case 'Assets':
      return 'text-green-700 bg-green-50';
    case 'Liabilities':
      return 'text-red-700 bg-red-50';
    case 'Revenue':
      return 'text-blue-700 bg-blue-50';
    case 'Expenses':
      return 'text-orange-700 bg-orange-50';
    case 'Equity':
      return 'text-purple-700 bg-purple-50';
    default:
      return 'text-gray-700 bg-gray-50';
  }
}

