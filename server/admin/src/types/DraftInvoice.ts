/**
 * Interface representing an invoice line item (timesheet entry)
 */
export interface DraftEntry {
  entry_id: number;
  project_id: number;
  billing_code_id: number;
  billing_code: string;
  notes: string;
  start_date: string;
  duration_hours: number;
  fee: number;
  user_name: string;
  user_role: string;
  impersonate_as_user_id?: number;
  impersonate_as_user_name?: string;
  is_impersonated: boolean;
  created_by_name?: string;
  state: string;
}

/**
 * Interface representing an invoice adjustment
 */
export interface DraftAdjustment {
  ID: number;
  invoice_id?: number;
  bill_id?: number;
  type: string;
  state: string;
  amount: number;
  notes: string;
  created_at?: string;
  updated_at?: string;
}

/**
 * Interface representing a draft invoice
 */
export interface DraftInvoice {
  ID: number;
  invoice_name: string;
  account_id: number;
  account_name: string;
  project_id: number;
  project_name: string;
  period_start: string;
  period_end: string;
  line_items: DraftEntry[];
  expenses: any[]; // Expense objects from backend
  adjustments: DraftAdjustment[];
  total_hours: number;
  total_fees: number;
  total_expenses: number;
  total_adjustments: number;
  total_amount: number;
  period_closed: boolean;
}

/**
 * Constants for draft invoice-related states
 */
export const ENTRY_STATES = [
  'ENTRY_STATE_DRAFT',
  'ENTRY_STATE_APPROVED',
  'ENTRY_STATE_SENT',
  'ENTRY_STATE_PAID',
  'ENTRY_STATE_VOID'
];

export const ADJUSTMENT_TYPES = [
  'ADJUSTMENT_TYPE_CREDIT',
  'ADJUSTMENT_TYPE_FEE'
];

export const ADJUSTMENT_STATES = [
  'ADJUSTMENT_STATE_DRAFT',
  'ADJUSTMENT_STATE_APPROVED',
  'ADJUSTMENT_STATE_SENT',
  'ADJUSTMENT_STATE_PAID',
  'ADJUSTMENT_STATE_VOID'
];

/**
 * Format currency value as USD
 */
export function formatCurrency(value: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD'
  }).format(value);
}

/**
 * Format date string to readable format
 */
export function formatDate(dateString: string): string {
  if (!dateString) return '';
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric', 
    month: 'short', 
    day: 'numeric'
  });
}

/**
 * Get status class for entry state
 */
export function getEntryStateClass(state: string): string {
  switch (state) {
    case 'ENTRY_STATE_UNAFFILIATED':
      return 'bg-orange-100 text-orange-800';
    case 'ENTRY_STATE_DRAFT':
      return 'bg-blue-100 text-blue-800';
    case 'ENTRY_STATE_REJECTED':
      return 'bg-red-100 text-red-800';
    case 'ENTRY_STATE_APPROVED':
      return 'bg-green-100 text-green-800';
    case 'ENTRY_STATE_EXCLUDED':
      return 'bg-gray-100 text-gray-600';
    case 'ENTRY_STATE_SENT':
      return 'bg-indigo-100 text-indigo-800';
    case 'ENTRY_STATE_PAID':
      return 'bg-emerald-100 text-emerald-800';
    case 'ENTRY_STATE_VOID':
      return 'bg-slate-100 text-slate-800';
    default:
      return 'bg-gray-100 text-gray-800';
  }
}

/**
 * Get display name for entry state
 */
export function getEntryStateDisplayName(state: string): string {
  return state.replace('ENTRY_STATE_', '');
} 