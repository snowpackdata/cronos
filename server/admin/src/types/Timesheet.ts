/**
 * Interface representing a Timesheet Entry in the system
 */
export interface TimesheetEntry {
  entry_id: number;
  user_id: number;
  employee_name: string;
  billing_code_id: number;
  billing_code_name: string;
  billing_code: string;
  project_id: number;
  project_name: string;
  start: string;
  end: string;
  duration_hours: number;
  fee: number;
  state: string;
  notes: string;
  impersonate_as_user_id: number | null;
  is_being_impersonated: boolean;
  is_meeting: boolean;
}

/**
 * Constants for timesheet entry states
 */
export const ENTRY_STATES = [
  'ENTRY_STATE_UNAFFILIATED',
  'ENTRY_STATE_DRAFT',
  'ENTRY_STATE_REJECTED',
  'ENTRY_STATE_APPROVED',
  'ENTRY_STATE_EXCLUDED',
  'ENTRY_STATE_INVOICED',
  'ENTRY_STATE_SENT',
  'ENTRY_STATE_PAID',
  'ENTRY_STATE_VOID'
];

/**
 * Creates a new empty timesheet entry with default values
 */
export function createEmptyTimesheetEntry(): TimesheetEntry {
  return {
    entry_id: 0,
    user_id: 0,
    employee_name: '',
    billing_code_id: 0,
    billing_code_name: '',
    billing_code: '',
    project_id: 0,
    project_name: '',
    start: new Date().toISOString(),
    end: new Date(Date.now() + 3600000).toISOString(), // 1 hour later
    duration_hours: 1,
    fee: 0,
    state: 'ENTRY_STATE_DRAFT',
    notes: '',
    impersonate_as_user_id: null,
    is_being_impersonated: false,
    is_meeting: false
  };
} 