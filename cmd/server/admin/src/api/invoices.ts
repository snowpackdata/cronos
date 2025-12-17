import type { Invoice } from '../types/Invoice';
import { fetchAll, fetchById, create, update, remove } from './apiUtils';

/**
 * API service for invoice-related operations
 */
const invoicesAPI = {
  /**
   * Get all invoices
   * @returns Promise with array of invoices
   */
  async getInvoices(): Promise<Invoice[]> {
    return fetchAll<Invoice>('invoices/accepted');
  },

  /**
   * Get a single invoice by ID
   * @param id - Invoice ID
   * @returns Promise with invoice data
   */
  async getInvoice(id: number): Promise<Invoice> {
    return fetchById<Invoice>('invoices', id);
  },

  /**
   * Create a new invoice
   * @param invoice - Invoice data
   * @returns Promise with created invoice
   */
  async createInvoice(invoice: Invoice): Promise<Invoice> {
    return create<Invoice>('invoices', invoice);
  },

  /**
   * Update an existing invoice
   * @param invoice - Invoice data to update
   * @returns Promise with updated invoice
   */
  async updateInvoice(invoice: Invoice): Promise<Invoice> {
    return update<Invoice>('invoices', invoice.ID, invoice);
  },

  /**
   * Delete an invoice
   * @param id - ID of invoice to delete
   * @returns Promise with deletion result
   */
  async deleteInvoice(id: number): Promise<any> {
    return remove('invoices', id);
  },
  
  /**
   * Change the state of an invoice
   * @param id - Invoice ID
   * @param state - New state value
   * @returns Promise with updated invoice
   */
  async changeInvoiceState(id: number, state: string): Promise<Invoice> {
    return create<Invoice>(`invoices/state/${id}/${state}`, {});
  },

  /**
   * Get draft invoices
   * @returns Promise with array of draft invoices
   */
  async getDraftInvoices(): Promise<Invoice[]> {
    return fetchAll<Invoice>('invoices/draft');
  }
};

// Export as default
export default invoicesAPI;

// Export individual functions for backward compatibility
export const getInvoices = invoicesAPI.getInvoices;
export const getInvoice = invoicesAPI.getInvoice;
export const createInvoice = invoicesAPI.createInvoice;
export const updateInvoice = invoicesAPI.updateInvoice;
export const deleteInvoice = invoicesAPI.deleteInvoice;
export const changeInvoiceState = invoicesAPI.changeInvoiceState;
export const getDraftInvoices = invoicesAPI.getDraftInvoices; 