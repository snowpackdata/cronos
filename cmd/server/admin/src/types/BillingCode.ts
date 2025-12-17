import type { Rate } from "./Rate";

/**
 * Interface representing a Billing Code in the system
 */
export interface BillingCode {
  ID: number;
  code: string;
  name: string;
  type: string;
  description: string;
  account_id: number;
  rounded_to: number;
  project: number;
  active_start: string;
  active_end: string;
  rate_id: number;
  rate: Rate;
  internal_rate_id: number;
  internal_rate: Rate;
  // Additional properties used in createEmptyBillingCode
  account_name: string;
  project_id: number;
  project_name: string;
  rate_description: string;
  internal: boolean;
  active: boolean;
}

/**
 * Creates a new empty billing code with default values
 */
export function createEmptyBillingCode(): BillingCode {
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
    code: '',
    name: '',
    type: '',
    description: '',
    account_id: 0,
    account_name: '',
    rounded_to: 0,
    project: 0,
    project_id: 0,
    project_name: '',
    active_start: todayFormatted,
    active_end: futureFormatted,
    rate_id: 0,
    rate: { 
      ID: 0, 
      name: '', 
      amount: 0, 
      active_from: todayFormatted,
      active_to: futureFormatted,
      internal_only: false
    },
    rate_description: '',
    internal_rate_id: 0,
    internal_rate: { 
      ID: 0, 
      name: '', 
      amount: 0, 
      active_from: todayFormatted,
      active_to: futureFormatted,
      internal_only: true
    },
    internal: false,
    active: true
  };
} 