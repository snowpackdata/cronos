/**
 * Interface representing a Rate in the system
 */
export interface Rate {
  ID: number;
  name: string;
  amount: number;
  active_from: string;
  active_to: string;
  internal_only: boolean;
  CreatedAt?: string;
  UpdatedAt?: string;
  DeletedAt?: string | null;
  BillingCodes?: any[];
}

/**
 * Constants for rate types
 */
export const RATE_TYPES = [
  'RATE_TYPE_INTERNAL_CLIENT_NON_BILLABLE',
  'RATE_TYPE_INTERNAL_CLIENT_BILLABLE',
  'RATE_TYPE_INTERNAL_ADMINISTRATIVE_NON_BILLABLE',
  'RATE_TYPE_INTERNAL_ADMINISTRATIVE',
  'RATE_TYPE_EXTERNAL_CLIENT_BILLABLE',
  'RATE_TYPE_EXTERNAL_CLIENT_NON_BILLABLE',
  'RATE_TYPE_INTERNAL_PROJECT'
];

/**
 * Human-readable rate type mapping
 */
export const RATE_TYPE_NAMES = {
  'RATE_TYPE_INTERNAL_CLIENT_NON_BILLABLE': 'Internal Client Non-Billable',
  'RATE_TYPE_INTERNAL_CLIENT_BILLABLE': 'Internal Client',
  'RATE_TYPE_INTERNAL_ADMINISTRATIVE_NON_BILLABLE': 'Internal Administrative Non-Billable',
  'RATE_TYPE_INTERNAL_ADMINISTRATIVE': 'Internal Administrative',
  'RATE_TYPE_EXTERNAL_CLIENT_BILLABLE': 'External Client',
  'RATE_TYPE_EXTERNAL_CLIENT_NON_BILLABLE': 'External Client Non-Billable',
  'RATE_TYPE_INTERNAL_PROJECT': 'Internal Project'
};

/**
 * Creates a new empty rate with default values
 */
export function createEmptyRate(): Rate {
  // Use UTC date functions to ensure consistent date handling
  const today = new Date();
  const futureDate = new Date();
  
  // Default end date is 1 year in the future, use UTC functions
  futureDate.setUTCFullYear(futureDate.getUTCFullYear() + 1);
  
  // Format the dates as YYYY-MM-DD strings in UTC
  const todayFormatted = `${today.getUTCFullYear()}-${String(today.getUTCMonth() + 1).padStart(2, '0')}-${String(today.getUTCDate()).padStart(2, '0')}`;
  const futureFormatted = `${futureDate.getUTCFullYear()}-${String(futureDate.getUTCMonth() + 1).padStart(2, '0')}-${String(futureDate.getUTCDate()).padStart(2, '0')}`;
  
  return {
    ID: 0,
    name: '',
    amount: 0,
    active_from: todayFormatted,
    active_to: futureFormatted,
    internal_only: false
  };
} 