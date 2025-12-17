/**
 * Interface representing a Project in the system
 */

import { createEmptyAccount } from './Account';
import type { Asset } from './Asset';

// Define AdminStaff interface for preloaded staff information
export interface Staff {
  ID: number;
  user_id: number;
  user?: {
    ID: number;
    email: string;
    role: string;
    is_admin: boolean;
  };
  first_name?: string;
  last_name?: string;
  title?: string;
  email?: string;
  is_active?: boolean;
  is_owner?: boolean;
  employment_status?: 'EMPLOYMENT_STATUS_ACTIVE' | 'EMPLOYMENT_STATUS_INACTIVE' | 'EMPLOYMENT_STATUS_TERMINATED';
  start_date?: string | Date;
  end_date?: string | Date;
  headshot_asset_id?: number;
  headshot_asset?: {
    url?: string;
  };
  capacity_weekly?: number;
  compensation_type?: 'COMPENSATION_TYPE_FULLY_VARIABLE' | 'COMPENSATION_TYPE_SALARIED' | 'COMPENSATION_TYPE_BASE_PLUS_VARIABLE';
  is_salaried?: boolean; // Keep for backward compatibility with backend
  salary_annualized?: number;
  base_salary?: number;
  is_variable_hourly?: boolean;
  is_fixed_hourly?: boolean;
  hourly_rate?: number;
  entry_pay_eligible_state?: string;
  // Include optional server fields
  CreatedAt?: string;
  UpdatedAt?: string;
  DeletedAt?: string | null;
}

// Define StaffingAssignment interface
export interface StaffingAssignment {
  ID: number;
  project_id: number;
  employee_id: number; // Changed from staff_id
  commitment?: number; // Legacy/average commitment for backward compatibility
  employee?: Staff; // Updated to use Staff
  start_date?: string | Date; // Changed from assignment_active_start
  end_date?: string | Date;   // Changed from assignment_active_end
  commitment_schedule?: string; // JSON string of CommitmentSchedule
  // Add other assignment-specific fields if needed, e.g., assigned_at, allocation_percentage
}

export interface Project {
  ID: number;
  name: string;
  account_id: number;
  account?: any; // Account reference object
  active_start: string; // Project start date
  active_end: string; // Project end date
  budget_hours: number;
  budget_dollars: number;
  budget_cap_hours: number;
  budget_cap_dollars: number;
  internal: boolean;
  billing_frequency: string;
  project_type: string;
  ae_id?: number; // Optional to match Go's *uint
  sdr_id?: number; // Optional to match Go's *uint
  description?: string; // Added new description field
  staffing_assignments?: StaffingAssignment[]; // Added for preloaded staffing assignments
  assets?: Asset[]; // Added for preloaded assets
  // Include optional server fields to avoid mapping errors
  CreatedAt?: string;
  UpdatedAt?: string;
  DeletedAt?: string | null;
  billing_codes?: any[]; // The server may include associated billing codes
}

/**
 * Constants for project types
 */
export const PROJECT_TYPES = [
  'PROJECT_TYPE_NEW',
  'PROJECT_TYPE_EXISTING'
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
 * Creates a new empty project with default values
 */
export function createEmptyProject(): Project {
  // Use UTC date functions to ensure consistent date handling
  const today = new Date();
  const endDate = new Date();

  // Default end date is 6 months from now, use UTC functions
  endDate.setUTCMonth(today.getUTCMonth() + 6);

  // Format the dates as YYYY-MM-DD strings in UTC
  const todayFormatted = `${today.getUTCFullYear()}-${String(today.getUTCMonth() + 1).padStart(2, '0')}-${String(today.getUTCDate()).padStart(2, '0')}`;
  const endDateFormatted = `${endDate.getUTCFullYear()}-${String(endDate.getUTCMonth() + 1).padStart(2, '0')}-${String(endDate.getUTCDate()).padStart(2, '0')}`;

  const account = createEmptyAccount();
  return {
    ID: 0,
    name: '',
    account_id: 0,
    account: account,
    active_start: todayFormatted,
    active_end: endDateFormatted,
    budget_hours: 0,
    budget_dollars: 0,
    budget_cap_hours: 0,
    budget_cap_dollars: 0,
    internal: false,
    billing_frequency: 'BILLING_TYPE_MONTHLY',
    project_type: 'PROJECT_TYPE_NEW',
    description: '', // Initialize new description field
    staffing_assignments: [], // Initialize staffing assignments
    assets: [] // Initialize assets array
  };
}
