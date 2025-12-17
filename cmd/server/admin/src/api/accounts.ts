import type { Account } from '../types/Account';
import { fetchAll, fetchById, createWithFormData, updateWithFormData, remove } from './apiUtils';

/**
 * Validates account data before sending to the API
 * @param account - Account data to validate
 * @returns Object with validation result and any error messages
 */
function validateAccount(account: Partial<Account>): { isValid: boolean; errors: string[] } {
  const errors: string[] = [];
  
  // Check required fields
  if (!account.name) errors.push('Account name is required');
  if (!account.billing_frequency) errors.push('Billing frequency is required');
  
  return { 
    isValid: errors.length === 0,
    errors 
  };
}

/**
 * Prepares account data for the backend by transforming it to the expected format
 * @param account - Account data to transform
 * @returns Prepared account data as FormData
 */
function prepareAccountForApi(account: Account): FormData {
  const formData = new FormData();
  
  // Set required fields
  formData.set("name", account.name);
  
  // Add other fields from the Account interface
  if (account.legal_name) {
    formData.set("legal_name", account.legal_name);
  }
  
  if (account.address) {
    formData.set("address", account.address);
  }
  
  if (account.website) {
    formData.set("website", account.website);
  }
  
  if (account.email) {
    formData.set("email", account.email);
  }
  
  if (account.type) {
    formData.set("type", account.type);
  }
  
  // Always set billing_frequency with a default if not provided
  if (account.billing_frequency) {
    formData.set("billing_frequency", account.billing_frequency);
  } else {
    // Default to monthly if not specified
    formData.set("billing_frequency", "BILLING_TYPE_MONTHLY");
  }
  
  formData.set("budget_hours", account.budget_hours.toString());
  formData.set("budget_dollars", account.budget_dollars.toString());
  formData.set("projects_single_invoice", account.projects_single_invoice ? "true" : "false");
  
  
  return formData;
}

/**
 * API service for account-related operations
 */
const accountsAPI = {
  /**
   * Fetch all accounts
   * @returns Promise with account data
   */
  async fetchAccounts(): Promise<Account[]> {
    return fetchAll<Account>('accounts');
  },

  /**
   * Get a single account by ID
   * @param id - Account ID
   * @returns Promise with account data
   */
  async getAccount(id: number): Promise<Account> {
    return fetchById<Account>('accounts', id);
  },

  /**
   * Create a new account
   * @param account - The account data to create
   * @returns Promise with the created account
   */
  async createAccount(account: Account): Promise<Account> {
    // Validate account data
    const validation = validateAccount(account);
    if (!validation.isValid) {
      throw new Error(`Validation failed: ${validation.errors.join(', ')}`);
    }
    
    try {
      // Prepare data and use the createWithFormData utility
      const formData = prepareAccountForApi(account);
      return createWithFormData<Account>('accounts', formData);
    } catch (error) {
      console.error('Failed to create account:', error);
      throw error;
    }
  },

  /**
   * Update an existing account
   * @param account - The account data to update
   * @returns Promise with the updated account
   */
  async updateAccount(account: Account): Promise<Account> {
    // Validate account data
    const validation = validateAccount(account);
    if (!validation.isValid) {
      throw new Error(`Validation failed: ${validation.errors.join(', ')}`);
    }
    
    try {
      // Prepare data and use the updateWithFormData utility
      const formData = prepareAccountForApi(account);
      return updateWithFormData<Account>('accounts', account.ID, formData);
    } catch (error) {
      console.error('Failed to update account:', error);
      throw error;
    }
  },

  /**
   * Delete an account
   * @param id - The ID of the account to delete
   * @returns Promise with the response data
   */
  async deleteAccount(id: number): Promise<any> {
    try {
      return remove('accounts', id);
    } catch (error) {
      console.error('Failed to delete account:', error);
      throw error;
    }
  },

  /**
   * Invite a user to an account
   * @param accountId - The ID of the account
   * @param email - The email of the user to invite
   * @returns Promise with the invited user data
   */
  async inviteUser(accountId: number, email: string): Promise<any> {
    try {
      const formData = new FormData();
      formData.append('email', email);
      // Assuming createWithFormData handles FormData correctly for a POST request.
      // The path 'accounts/' + accountId + '/invite' is specific for this action.
      return createWithFormData(`accounts/${accountId}/invite`, formData);
    } catch (error) {
      console.error('Failed to invite user:', error);
      throw error;
    }
  }
};

// Export as default
export default accountsAPI;

// Export individual functions for backward compatibility
export const fetchAccounts = accountsAPI.fetchAccounts;
export const getAccount = accountsAPI.getAccount;
export const createAccount = accountsAPI.createAccount;
export const updateAccount = accountsAPI.updateAccount;
export const deleteAccount = accountsAPI.deleteAccount;
export const inviteUser = accountsAPI.inviteUser;



