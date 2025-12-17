import type { Asset } from './Asset'; // Added import for Asset type

/**
 * Interface representing an Account in the system
 */
export interface Client {
  ID: number;
  email: string;
  role: string;
}

export interface Account {
  ID: number;
  id?: number; // For frontend compatibility
  name: string;
  legal_name: string;
  address: string;
  website: string;
  email: string;
  type: string;
  billing_frequency: string;
  budget_hours: number;
  budget_dollars: number;
  projects_single_invoice: boolean;
  clients: Client[] | null;
  client_users?: any[]; // Added client_users property
  assets?: Asset[]; // Added assets property
  
  // Properties for backward compatibility - to be deprecated
  budget_amount?: number;
  budget_year?: number;
  number_of_clients?: number;
}

/**
 * Constants for account types
 */
export const ACCOUNT_TYPES = [
  'ACCOUNT_TYPE_INTERNAL',
  'ACCOUNT_TYPE_CLIENT'
];

/**
 * Constants for billing frequencies
 */
export const BILLING_FREQUENCIES = [
  'BILLING_TYPE_MONTHLY',
  'BILLING_TYPE_PROJECT',
  'BILLING_TYPE_BIWEEKLY',
  'BILLING_TYPE_WEEKLY',
  'BILLING_TYPE_BIMONTHLY'
];

/**
 * Creates a new empty account object with default values
 */
export function createEmptyAccount(): Account {
  return {
    ID: 0,
    name: '',
    legal_name: '',
    address: '',
    website: '',
    email: '',
    type: 'ACCOUNT_TYPE_CLIENT',
    billing_frequency: 'BILLING_TYPE_MONTHLY',
    budget_hours: 0,
    budget_dollars: 0,
    projects_single_invoice: true,
    clients: [],
    client_users: [], // Initialize client_users
    assets: [], // Initialize assets
  };
} 