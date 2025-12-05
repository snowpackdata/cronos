import { fetchAll, fetchById, create } from './apiUtils';
import type { DraftInvoice } from '../types/DraftInvoice';
// import type { DraftInvoice, DraftEntry } from '../types/DraftInvoice';

/**
 * API service for draft invoice operations
 */
const draftInvoicesAPI = {
  /**
   * Get all draft invoices
   * @returns Promise with array of draft invoices
   */
  async getDraftInvoices(): Promise<DraftInvoice[]> {
    return fetchAll<DraftInvoice>('invoices/draft');
  },

  /**
   * Get a single draft invoice by ID
   * @param id - Invoice ID
   * @returns Promise with draft invoice data
   */
  async getDraftInvoice(id: number): Promise<DraftInvoice> {
    return fetchById<DraftInvoice>('invoices', id);
  },

  /**
   * Approve a draft invoice
   * @param id - Invoice ID
   * @returns Promise with operation result
   */
  async approveDraftInvoice(id: number): Promise<any> {
    return create<any>(`invoices/state/${id}/approve`, {});
  },

  /**
   * Void a draft invoice
   * @param id - Invoice ID
   * @returns Promise with operation result
   */
  async voidDraftInvoice(id: number): Promise<any> {
    return create<any>(`invoices/state/${id}/void`, {});
  },

  /**
   * Approve an entry within a draft invoice
   * @param id - Entry ID
   * @returns Promise with operation result
   */
  async approveEntry(id: number): Promise<any> {
    return create<any>(`entries/state/${id}/approve`, {});
  },

  /**
   * Void an entry within a draft invoice
   * @param id - Entry ID
   * @returns Promise with operation result
   */
  async voidEntry(id: number): Promise<any> {
    return create<any>(`entries/state/${id}/void`, {});
  }
};

// Export as default
export default draftInvoicesAPI;

// Export individual functions
export const getDraftInvoices = draftInvoicesAPI.getDraftInvoices;
export const getDraftInvoice = draftInvoicesAPI.getDraftInvoice;
export const approveDraftInvoice = draftInvoicesAPI.approveDraftInvoice;
export const voidDraftInvoice = draftInvoicesAPI.voidDraftInvoice;
export const approveEntry = draftInvoicesAPI.approveEntry;
export const voidEntry = draftInvoicesAPI.voidEntry; 