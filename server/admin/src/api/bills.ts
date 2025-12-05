import type { Bill } from '../types/Bill';
import { fetchAll, fetchById, create, update, remove } from './apiUtils';

/**
 * API service for bill-related operations
 */
const billsAPI = {
  /**
   * Get all bills
   * @returns Promise with array of bills
   */
  async getBills(): Promise<Bill[]> {
    return fetchAll<Bill>('bills');
  },

  /**
   * Get a single bill by ID
   * @param id - Bill ID
   * @returns Promise with bill data
   */
  async getBill(id: number): Promise<Bill> {
    return fetchById<Bill>('bills', id);
  },

  /**
   * Create a new bill
   * @param bill - Bill data
   * @returns Promise with created bill
   */
  async createBill(bill: Bill): Promise<Bill> {
    return create<Bill>('bills', bill);
  },

  /**
   * Update an existing bill
   * @param bill - Bill data to update
   * @returns Promise with updated bill
   */
  async updateBill(bill: Bill): Promise<Bill> {
    return update<Bill>('bills', bill.ID, bill);
  },

  /**
   * Delete a bill
   * @param id - ID of bill to delete
   * @returns Promise with deletion result
   */
  async deleteBill(id: number): Promise<any> {
    return remove('bills', id);
  },
  
  /**
   * Change the state of a bill
   * @param id - Bill ID
   * @param state - New state value
   * @returns Promise with updated bill
   */
  async changeBillState(id: number, state: string): Promise<Bill> {
    return create<Bill>(`bills/state/${id}/${state}`, {});
  }
};

// Export as default
export default billsAPI;

// Export individual functions for backward compatibility
export const getBills = billsAPI.getBills;
export const getBill = billsAPI.getBill;
export const createBill = billsAPI.createBill;
export const updateBill = billsAPI.updateBill;
export const deleteBill = billsAPI.deleteBill;
export const changeBillState = billsAPI.changeBillState; 