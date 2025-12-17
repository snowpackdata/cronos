import type { Staff } from '../types/Project';
import { api as apiClient } from './apiUtils';

/**
 * Fetches a list of staff members (Staff).
 * @returns A promise that resolves to an array of Staff.
 */
export const fetchStaff = async (): Promise<Staff[]> => {
  try {
    const response = await apiClient.get('/api/staff');
    return response.data as Staff[];
  } catch (error) {
    console.error('Error fetching staff:', error);
    throw error;
  }
};

/**
 * Fetches a single staff member by ID.
 * @param id - The staff member's ID
 * @returns A promise that resolves to a Staff object.
 */
export const fetchStaffById = async (id: number): Promise<Staff> => {
  try {
    const response = await apiClient.get(`/api/staff/${id}`);
    return response.data as Staff;
  } catch (error) {
    console.error(`Error fetching staff member ${id}:`, error);
    throw error;
  }
};

/**
 * Creates a new staff member.
 * @param staffData - The staff member data
 * @returns A promise that resolves to the created Staff object.
 */
export const createStaff = async (staffData: Partial<Staff>): Promise<Staff> => {
  try {
    const formData = new FormData();

    // Required fields for new staff
    if (staffData.first_name) formData.append('first_name', staffData.first_name);
    if (staffData.last_name) formData.append('last_name', staffData.last_name);
    if (staffData.title) formData.append('title', staffData.title);
    if (staffData.email) formData.append('email', staffData.email);

    // Send both employment_status and is_active for backward compatibility
    const isActive = staffData.employment_status === 'EMPLOYMENT_STATUS_ACTIVE' || staffData.employment_status === 'EMPLOYMENT_STATUS_INACTIVE';
    formData.append('is_active', isActive.toString());

    if (staffData.start_date) formData.append('start_date', staffData.start_date.toString());
    if (staffData.end_date) formData.append('end_date', staffData.end_date.toString());
    if (staffData.capacity_weekly !== undefined) formData.append('capacity_weekly', staffData.capacity_weekly.toString());
    if (staffData.is_salaried !== undefined) formData.append('is_salaried', staffData.is_salaried.toString());
    if (staffData.salary_annualized !== undefined) formData.append('salary_annualized', staffData.salary_annualized.toString());
    if (staffData.base_salary !== undefined) formData.append('base_salary', staffData.base_salary.toString());
    if (staffData.is_variable_hourly !== undefined) formData.append('is_variable_hourly', staffData.is_variable_hourly.toString());
    if (staffData.is_fixed_hourly !== undefined) formData.append('is_fixed_hourly', staffData.is_fixed_hourly.toString());
    if (staffData.hourly_rate !== undefined) formData.append('hourly_rate', staffData.hourly_rate.toString());
    if (staffData.entry_pay_eligible_state) formData.append('entry_pay_eligible_state', staffData.entry_pay_eligible_state);
    if (staffData.employment_status) formData.append('employment_status', staffData.employment_status);
    if (staffData.compensation_type) formData.append('compensation_type', staffData.compensation_type);
    
    // Handle headshot file upload
    if ((staffData as any).headshot) {
      formData.append('headshot', (staffData as any).headshot);
    }

    const response = await apiClient.post('/api/staff/0', formData, {
      headers: {
        'Content-Type': undefined // Let axios set multipart/form-data automatically
      }
    });
    return response.data as Staff;
  } catch (error) {
    console.error('Error creating staff member:', error);
    throw error;
  }
};

/**
 * Updates an existing staff member.
 * @param id - The staff member's ID
 * @param staffData - The updated staff member data
 * @returns A promise that resolves to the updated Staff object.
 */
export const updateStaff = async (id: number, staffData: Partial<Staff>): Promise<Staff> => {
  try {
    const formData = new FormData();

    if (staffData.first_name) formData.append('first_name', staffData.first_name);
    if (staffData.last_name) formData.append('last_name', staffData.last_name);
    if (staffData.title) formData.append('title', staffData.title);
    if (staffData.email) formData.append('email', staffData.email);

    // Send both employment_status and is_active for backward compatibility
    if (staffData.employment_status !== undefined) {
      const isActive = staffData.employment_status === 'EMPLOYMENT_STATUS_ACTIVE' || staffData.employment_status === 'EMPLOYMENT_STATUS_INACTIVE';
      formData.append('is_active', isActive.toString());
    } else if (staffData.is_active !== undefined) {
      formData.append('is_active', staffData.is_active.toString());
    }

    if (staffData.start_date) formData.append('start_date', staffData.start_date.toString());
    if (staffData.end_date) formData.append('end_date', staffData.end_date.toString());
    if (staffData.capacity_weekly !== undefined) formData.append('capacity_weekly', staffData.capacity_weekly.toString());
    if (staffData.is_salaried !== undefined) formData.append('is_salaried', staffData.is_salaried.toString());
    if (staffData.salary_annualized !== undefined) formData.append('salary_annualized', staffData.salary_annualized.toString());
    if (staffData.base_salary !== undefined) formData.append('base_salary', staffData.base_salary.toString());
    if (staffData.is_variable_hourly !== undefined) formData.append('is_variable_hourly', staffData.is_variable_hourly.toString());
    if (staffData.is_fixed_hourly !== undefined) formData.append('is_fixed_hourly', staffData.is_fixed_hourly.toString());
    if (staffData.hourly_rate !== undefined) formData.append('hourly_rate', staffData.hourly_rate.toString());
    if (staffData.entry_pay_eligible_state) formData.append('entry_pay_eligible_state', staffData.entry_pay_eligible_state);
    if (staffData.employment_status) formData.append('employment_status', staffData.employment_status);
    if (staffData.compensation_type) formData.append('compensation_type', staffData.compensation_type);
    
    // Handle headshot file upload
    if ((staffData as any).headshot) {
      formData.append('headshot', (staffData as any).headshot);
    }

    const response = await apiClient.put(`/api/staff/${id}`, formData, {
      headers: {
        'Content-Type': undefined // Let axios set multipart/form-data automatically
      }
    });
    return response.data as Staff;
  } catch (error) {
    console.error(`Error updating staff member ${id}:`, error);
    throw error;
  }
};

/**
 * Deletes a staff member.
 * @param id - The staff member's ID to delete
 * @returns A promise that resolves when the deletion is complete.
 */
export const deleteStaff = async (id: number): Promise<void> => {
  try {
    await apiClient.delete(`/api/staff/${id}`);
  } catch (error) {
    console.error(`Error deleting staff member ${id}:`, error);
    throw error;
  }
};

// Default export for the staff API
const staffAPI = {
  fetchStaff,
  fetchStaffById,
  createStaff,
  updateStaff,
  deleteStaff
};

export default staffAPI;
