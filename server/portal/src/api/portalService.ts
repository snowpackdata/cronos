import { apiClient } from './index';
import type { Account } from '../types/Account';
import type { Asset } from '../types/Asset';

// TODO: Define interfaces for PortalProject, PortalInvoice etc. for better type safety
// interface PortalProject {
//   ID: number;
//   name: string;
//   // ... other project fields
// }

// interface PortalInvoice {
//   ID: number;
//   invoice_number: string;
//   status: string;
//   total_amount: number;
//   // ... other invoice fields
//   entries?: any[]; // For draft invoices with line items
// }

/**
 * Fetches projects for the client portal.
 */
export const fetchPortalProjects = async () => {
  try {
    const response = await apiClient.get('/api/portal/projects');
    return response.data; // as PortalProject[] (if interface is defined)
  } catch (error) {
    console.error('Error fetching portal projects:', error);
    throw error;
  }
};

/**
 * Fetches draft invoices for the client portal.
 * These typically include line items (entries).
 */
export const fetchPortalDraftInvoices = async () => {
  try {
    const response = await apiClient.get('/api/portal/invoices/draft');
    return response.data; // as PortalInvoice[] (if interface is defined)
  } catch (error) {
    console.error('Error fetching portal draft invoices:', error);
    throw error;
  }
};

/**
 * Fetches accepted (approved, sent, paid) invoices for the client portal.
 */
export const fetchPortalAcceptedInvoices = async () => {
  try {
    const response = await apiClient.get('/api/portal/invoices/accepted');
    return response.data; // as PortalInvoice[] (if interface is defined)
  } catch (error) {
    console.error('Error fetching portal accepted invoices:', error);
    throw error;
  }
};

/**
 * Fetches comprehensive account details for the settings page.
 * This includes basic account information, associated clients, and assets.
 * Assumes a single endpoint provides all this data.
 */
export const fetchAccountDetails = async (): Promise<Account> => {
  try {
    const response = await apiClient.get('/api/portal/account-details');
    const accountData = response.data as Account;
    if (accountData.assets) {
      accountData.assets = accountData.assets as Asset[];
    }
    return accountData;
  } catch (error) {
    console.error('Error fetching account details:', error);
    throw error;
  }
};

/**
 * Refreshes the signed URL for a GCS asset in the Portal.
 * @param assetId The ID of the asset to refresh.
 * @returns A promise that resolves to an object containing the new URL and expiration time.
 */
export const refreshAssetUrl = async (assetId: number): Promise<{ new_url: string; new_expires_at: string }> => {
  try {
    // The actual endpoint path may vary based on your backend API routes for the portal.
    // Example: '/api/portal/assets/{assetId}/refresh-url'
    const response = await apiClient.post(`/api/portal/assets/${assetId}/refresh-url`, {});
    // Ensure the response.data matches the expected structure
    if (response.data && response.data.new_url && response.data.new_expires_at) {
      return response.data;
    }
    throw new Error('Invalid response structure from refreshAssetUrl API via portalService');
  } catch (error) {
    console.error(`Error refreshing asset URL for asset ID ${assetId} via portalService:`, error);
    throw error; // Re-throw to be handled by the API consumer (e.g., assets.ts)
  }
}; 