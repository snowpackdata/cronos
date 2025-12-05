import type { Rate } from '../types/Rate';
import { 
  fetchAll, 
  fetchById, 
  remove,
  createWithFormData,
  updateWithFormData
} from './apiUtils';

/**
 * Validates rate data before sending to the API
 * @param rate - Rate data to validate
 * @returns Object with validation result and any error messages
 */
function validateRate(rate: Partial<Rate>): { isValid: boolean; errors: string[] } {
  const errors: string[] = [];
  
  // Check required fields
  if (!rate.name) errors.push('Rate name is required');
  if (rate.amount === undefined || rate.amount === null) errors.push('Amount is required');
  if (!rate.active_from) errors.push('Start date is required');
  
  // Validate date range
  if (rate.active_from && rate.active_to) {
    const startDate = new Date(rate.active_from);
    const endDate = new Date(rate.active_to);
    if (endDate < startDate) {
      errors.push('End date must be after start date');
    }
  }
  
  return { 
    isValid: errors.length === 0,
    errors 
  };
}

/**
 * Prepares a Rate object for API submission
 * @param rate - The rate data to prepare
 * @returns FormData object ready for API submission
 */
function prepareRateForApi(rate: Rate): FormData {
  
  const formData = new FormData();
  
  // Set required fields
  formData.set("name", rate.name);
  formData.set("amount", rate.amount.toString());
  
  // Ensure active_from is a valid string in YYYY-MM-DD format
  if (!rate.active_from) {
    console.error('Missing required active_from date');
    const today = new Date();
    // Use UTC date for consistency
    const formattedDate = `${today.getUTCFullYear()}-${String(today.getUTCMonth() + 1).padStart(2, '0')}-${String(today.getUTCDate()).padStart(2, '0')}`;
    formData.set("active_from", formattedDate);
  } else {
    formData.set("active_from", rate.active_from);
  }
  
  // Add optional fields
  if (rate.active_to) {
    formData.set("active_to", rate.active_to);
  }
  
  if (rate.internal_only !== undefined) {
    formData.set("internal_only", rate.internal_only ? "true" : "false");
  }
  
  // @ts-ignore - For debug purposes only
  for (const [key, value] of formData.entries()) {
  }
  
  return formData;
}

/**
 * API service for rate-related operations
 */
const ratesAPI = {
  /**
   * Get all rates
   * @returns Promise with array of rates
   */
  async getRates(): Promise<Rate[]> {
    return fetchAll<Rate>('rates');
  },

  /**
   * Get a single rate by ID
   * @param id - Rate ID
   * @returns Promise with rate data
   */
  async getRate(id: number): Promise<Rate> {
    return fetchById<Rate>('rates', id);
  },

  /**
   * Create a new rate with validation
   * @param rate - Rate data
   * @returns Promise with created rate
   */
  async createRate(rate: Rate): Promise<Rate> {
    // Validate rate data
    const validation = validateRate(rate);
    if (!validation.isValid) {
      throw new Error(`Validation failed: ${validation.errors.join(', ')}`);
    }
    
    try {
      // Prepare data and use the existing createWithFormData utility
      const formData = prepareRateForApi(rate);
      return createWithFormData<Rate>('rates', formData);
    } catch (error) {
      console.error('Failed to create rate:', error);
      throw error;
    }
  },

  /**
   * Update an existing rate with validation
   * @param rate - Rate data to update
   * @returns Promise with updated rate
   */
  async updateRate(rate: Rate): Promise<Rate> {
    // Validate rate data
    const validation = validateRate(rate);
    if (!validation.isValid) {
      throw new Error(`Validation failed: ${validation.errors.join(', ')}`);
    }
    
    try {
      // Prepare data and use the existing updateWithFormData utility
      const formData = prepareRateForApi(rate);
      return updateWithFormData<Rate>('rates', rate.ID, formData);
    } catch (error) {
      console.error('Failed to update rate:', error);
      throw error;
    }
  },

  /**
   * Delete a rate
   * @param id - ID of rate to delete
   * @returns Promise with deletion result
   */
  async deleteRate(id: number): Promise<any> {
    try {
      return remove('rates', id);
    } catch (error) {
      console.error('Failed to delete rate:', error);
      throw error;
    }
  }
};

// Exported API functions

/**
 * Fetches all rates from the API
 * @returns Promise with array of rates
 */
export async function fetchRates(): Promise<Rate[]> {
  try {
    return await fetchAll<Rate>('rates');
  } catch (error) {
    console.error('Error fetching rates:', error);
    throw error;
  }
}

/**
 * Fetches a single rate by ID from the API
 * @param id - Rate ID
 * @returns Promise with rate data
 */
export const fetchRateById = async (id: number) => {
  try {
    return await fetchById<Rate>('rates', id);
  } catch (error) {
    console.error(`Error fetching rate with ID ${id}:`, error);
    throw error;
  }
};

/**
 * Creates a new rate with validation
 * @param rateData Rate data to create
 * @returns Promise with created rate data
 */
export const createRate = async (rateData: any) => {
  return await ratesAPI.createRate(rateData as Rate);
};

/**
 * Updates an existing rate with validation
 * @param rateData Updated rate data
 * @returns Promise with updated rate data
 */
export const updateRate = async (rateData: any) => {
  // Ensure ID is provided in the rate data
  if (!rateData.ID) {
    throw new Error('Rate ID is required for updates');
  }
  return await ratesAPI.updateRate(rateData as Rate);
};

/**
 * Deletes a rate
 * @param id Rate ID to delete
 * @returns Promise with deletion status
 */
export const deleteRate = async (id: number) => {
  return await ratesAPI.deleteRate(id);
};

// Export the API object explicitly to match the accounts pattern
export { ratesAPI };

export default ratesAPI; 