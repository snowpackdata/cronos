/**
 * Interface representing an Invoice Line Item
 */
export interface InvoiceLineItem {
  ID: number;
  invoice_id: number;
  type: string;
  description: string;
  quantity: number;
  rate: number;
  amount: number;
  billing_code_id?: number;
  billing_code?: {
    code: string;
    name: string;
  };
  employee_id?: number;
  employee?: {
    first_name: string;
    last_name: string;
  };
  adjustment_id?: number;
  entry_ids?: string;
}

/**
 * Interface representing an Invoice Adjustment
 */
export interface Adjustment {
  ID: number;
  invoice_id: number;
  type: string;
  amount: number;
  notes: string;
  state: string;
}

/**
 * Interface representing an Invoice in the system
 */
export interface Invoice {
  ID: number;
  account_id: number;
  account_name: string;
  account?: {
    name: string;
    id: number;
    email?: string;
  };
  project?: {
    name: string;
    id: number;
  };
  project_id?: number;
  invoice_number: string;
  state: string;
  type: string;
  date_created: string;
  date_approved: string;
  date_sent: string;
  date_paid: string;
  date_voided: string;
  date_due: string; // Due date for the invoice
  description: string; // Description/notes for the invoice
  entries: any[]; // Timesheet entries associated with this invoice
  adjustments: Adjustment[];
  line_items?: InvoiceLineItem[]; // Line items for the invoice
  total_hours: number;
  total_fees: number;
  total_adjustments: number;
  total_amount: number;
  // Additional properties from the backend model
  accepted_at?: string;
  sent_at?: string;
  due_at: string;
  closed_at?: string;
  period_start: string;
  period_end: string;
  file?: string; // GCS file URL for the invoice PDF
}

/**
 * Constants for invoice states
 */
export const INVOICE_STATES = [
  'INVOICE_STATE_DRAFT',
  'INVOICE_STATE_APPROVED',
  'INVOICE_STATE_SENT',
  'INVOICE_STATE_PAID',
  'INVOICE_STATE_VOID'
];

/**
 * Constants for adjustment types
 */
export const ADJUSTMENT_TYPES = [
  'ADJUSTMENT_TYPE_CREDIT',
  'ADJUSTMENT_TYPE_FEE'
];

/**
 * Constants for adjustment states
 */
export const ADJUSTMENT_STATES = [
  'ADJUSTMENT_STATE_DRAFT',
  'ADJUSTMENT_STATE_VOID'
];

/**
 * Creates a new empty invoice with default values
 */
export function createEmptyInvoice(): Invoice {
  const today = new Date();
  const dueDate = new Date();
  dueDate.setDate(today.getDate() + 30); // Default due date is 30 days from today
  
  return {
    ID: 0,
    account_id: 0,
    account_name: '',
    invoice_number: '',
    state: 'INVOICE_STATE_DRAFT',
    type: 'INVOICE_TYPE_AR',
    date_created: today.toISOString(),
    date_approved: '',
    date_sent: '',
    date_paid: '',
    date_voided: '',
    date_due: dueDate.toISOString(),
    description: '',
    entries: [],
    adjustments: [],
    total_hours: 0,
    total_fees: 0,
    total_adjustments: 0,
    total_amount: 0,
    due_at: dueDate.toISOString(),
    period_start: today.toISOString(),
    period_end: dueDate.toISOString()
  };
}

/**
 * Creates a new empty adjustment with default values
 */
export function createEmptyAdjustment(): Adjustment {
  return {
    ID: 0,
    invoice_id: 0,
    type: 'ADJUSTMENT_TYPE_FEE',
    amount: 0,
    notes: '',
    state: 'ADJUSTMENT_STATE_DRAFT'
  };
} 