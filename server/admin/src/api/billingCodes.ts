import type { BillingCode } from '../types/BillingCode';
import { 
  fetchAll, 
  fetchById, 
  createWithFormData, 
  updateWithFormData, 
  remove
} from './apiUtils';

/**
 * Validates billing code data before sending to the API
 * @param billingCode - Billing code data to validate
 * @returns Object with validation result and any error messages
 */
function validateBillingCode(billingCode: Partial<BillingCode>): { isValid: boolean; errors: string[] } {
  const errors: string[] = [];
  
  // Check required fields
  if (!billingCode.name) errors.push('Billing code name is required');
  if (!billingCode.rate_id) errors.push('Rate is required');
  
  // Add any additional validation rules here
  
  return { 
    isValid: errors.length === 0,
    errors 
  };
}

/**
 * Prepares billing code data for the backend by transforming it to the expected format
 * @param billingCode - Billing code data to transform
 * @returns Prepared billing code data for API
 */
function prepareBillingCodeForApi(billingCode: BillingCode): Record<string, any> {
  // Create a new object with only the fields expected by the backend
  return {
    name: billingCode.name,
    rate_id: billingCode.rate_id,
    description: billingCode.description,
    active: billingCode.active
  };
}

/**
 * API service for billing code-related operations
 */
const billingCodesAPI = {
  /**
   * Get all billing codes
   * @returns Promise with array of billing codes
   */
  async getBillingCodes(): Promise<BillingCode[]> {
    try {
      return await fetchAll<BillingCode>('billing_codes');
    } catch (error) {
      console.error('Error fetching billing codes:', error);
      throw error;
    }
  },

  /**
   * Get a single billing code by ID
   * @param id - Billing code ID
   * @returns Promise with billing code data
   */
  async getBillingCode(id: number): Promise<BillingCode> {
    return fetchById<BillingCode>('billing_codes', id);
  },

  /**
   * Create a new billing code with validation
   * @param billingCode - Billing code data
   * @returns Promise with created billing code
   */
  async createBillingCode(billingCode: BillingCode): Promise<BillingCode> {
    // Validate billing code data
    const validation = validateBillingCode(billingCode);
    if (!validation.isValid) {
      throw new Error(`Validation failed: ${validation.errors.join(', ')}`);
    }
    
    try {
      // Prepare data and use the createWithFormData utility
      const preparedData = prepareBillingCodeForApi(billingCode);
      return createWithFormData<BillingCode>('billing_codes', preparedData);
    } catch (error) {
      console.error('Failed to create billing code:', error);
      throw error;
    }
  },

  /**
   * Update an existing billing code with validation
   * @param billingCode - Billing code data to update
   * @returns Promise with updated billing code
   */
  async updateBillingCode(billingCode: BillingCode): Promise<BillingCode> {
    // Validate billing code data
    const validation = validateBillingCode(billingCode);
    if (!validation.isValid) {
      throw new Error(`Validation failed: ${validation.errors.join(', ')}`);
    }
    
    try {
      // Prepare data and use the updateWithFormData utility
      const preparedData = prepareBillingCodeForApi(billingCode);
      return updateWithFormData<BillingCode>('billing_codes', billingCode.ID, preparedData);
    } catch (error) {
      console.error('Failed to update billing code:', error);
      throw error;
    }
  },

  /**
   * Delete a billing code
   * @param id - ID of billing code to delete
   * @returns Promise with deletion result
   */
  async deleteBillingCode(id: number): Promise<any> {
    try {
      return remove('billing_codes', id);
    } catch (error) {
      console.error('Failed to delete billing code:', error);
      throw error;
    }
  },

  /**
   * Get active billing codes
   * @returns Promise with array of active billing codes
   */
  async getActiveBillingCodes(): Promise<BillingCode[]> {
    return fetchAll<BillingCode>('active_billing_codes');
  },

  /**
   * Get billing codes for a specific project
   * @param projectId - ID of the project
   * @returns Promise with array of billing codes
   */
  async getProjectBillingCodes(projectId: number): Promise<BillingCode[]> {
    return fetchAll<BillingCode>(`projects/${projectId}/billing_codes`);
  },

  /**
   * Add billing code to a project
   * @param projectId - ID of the project
   * @param billingCodeId - ID of the billing code
   * @returns Promise with result of the operation
   */
  async addBillingCodeToProject(projectId: number, billingCodeId: number): Promise<any> {
    try {
      return createWithFormData('projects/' + projectId + '/billing_codes', { billing_code_id: billingCodeId });
    } catch (error) {
      console.error('Failed to add billing code to project:', error);
      throw error;
    }
  },

  /**
   * Remove billing code from a project
   * @param projectId - ID of the project
   * @param billingCodeId - ID of the billing code
   * @returns Promise with result of the operation
   */
  async removeBillingCodeFromProject(projectId: number, billingCodeId: number): Promise<any> {
    try {
      return remove(`projects/${projectId}/billing_codes`, billingCodeId);
    } catch (error) {
      console.error('Failed to remove billing code from project:', error);
      throw error;
    }
  }
};

// Export as default
export default billingCodesAPI;

// Export individual functions for backward compatibility
export const getBillingCodes = async (projectId?: number): Promise<BillingCode[]> => {
  try {
    let endpoint = 'billing_codes';
    if (projectId) {
      endpoint = `projects/${projectId}/billing_codes`;
    }
    
    return await fetchAll<BillingCode>(endpoint);
  } catch (error) {
    console.error('Error fetching billing codes:', error);
    throw error;
  }
};

export const getBillingCode = async (id: number): Promise<BillingCode> => {
  try {
    return await fetchById<BillingCode>('billing_codes', id);
  } catch (error) {
    console.error(`Error fetching billing code with ID ${id}:`, error);
    throw error;
  }
};

export const createBillingCode = billingCodesAPI.createBillingCode;
export const updateBillingCode = billingCodesAPI.updateBillingCode;
export const deleteBillingCode = billingCodesAPI.deleteBillingCode;
export const getActiveBillingCodes = billingCodesAPI.getActiveBillingCodes;
export const getProjectBillingCodes = billingCodesAPI.getProjectBillingCodes;
export const addBillingCodeToProject = billingCodesAPI.addBillingCodeToProject;
export const removeBillingCodeFromProject = billingCodesAPI.removeBillingCodeFromProject;