// Export all API services
import accountsAPI from './accounts';
import billingCodesAPI from './billingCodes';
import projectsAPI from './projects';
import ratesAPI from './rates';
import timesheetAPI from './timesheet';
import invoicesAPI from './invoices';
import billsAPI from './bills';
import staffAPI from './staff';

// Import specific functions to resolve naming conflicts
import {
  fetchAccounts, getAccount, createAccount, updateAccount, deleteAccount, inviteUser
} from './accounts';

import {
  getBillingCodes, getBillingCode, createBillingCode, updateBillingCode, deleteBillingCode,
  getActiveBillingCodes as getBillingCodesActive, // Rename to avoid conflict
  getProjectBillingCodes, addBillingCodeToProject, removeBillingCodeFromProject
} from './billingCodes';

import {
  getEntries, getEntry, getUsers, getActiveBillingCodes as getTimesheetActiveBillingCodes, // Rename to avoid conflict
  createEntry, updateEntry, deleteEntry, getEntriesByDateRange, getEntriesByUser, getEntriesByProject
} from './timesheet';

import {
  getInvoices, getInvoice, createInvoice, updateInvoice, deleteInvoice,
  changeInvoiceState, getDraftInvoices
} from './invoices';

import {
  getBills, getBill, createBill, updateBill, deleteBill, changeBillState
} from './bills';

import axios from 'axios';

const apiClient = axios.create({
  // baseURL: '/api', // Your API base URL, already handled by Vite proxy for dev
  // Note: Don't set default Content-Type - let Axios set it automatically based on request data
  // (e.g., application/json for objects, multipart/form-data for FormData)
});

// Request interceptor to add the token to headers
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('snowpack_token'); // Or your specific token key
    if (token) {
      config.headers['x-access-token'] = token;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Loop prevention for token refresh
let refreshAttempts = 0;
const MAX_REFRESH_ATTEMPTS = 1; // Only try once to prevent infinite loops

// Response interceptor for automatic token refresh and retry
apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    const originalRequest = error.config;

    // Check for authentication errors (401, 403) and try to refresh token first
    if (error.response && (error.response.status === 401 || error.response.status === 403)) {
      console.log(`Authentication error (${error.response.status}) - attempting token refresh`);

      // Prevent infinite loops
      if (refreshAttempts >= MAX_REFRESH_ATTEMPTS) {
        console.log('Max refresh attempts reached - redirecting to login');
        localStorage.removeItem('snowpack_token');
        if (!window.location.pathname.includes('/login')) {
          window.location.href = '/login';
        }
        refreshAttempts = 0; // Reset counter
        return Promise.reject(error);
      }

      refreshAttempts++;

      try {
        // Try to refresh the token
        const { refreshToken } = await import('./apiUtils');
        const refreshSuccess = await refreshToken();

        if (refreshSuccess) {
          // Retry the original request with the new token
          const newToken = localStorage.getItem('snowpack_token');
          if (newToken && originalRequest.headers) {
            originalRequest.headers['x-access-token'] = newToken;
            refreshAttempts = 0; // Reset counter on successful refresh
            return apiClient(originalRequest);
          }
        }
      } catch (refreshError) {
        console.error('Token refresh failed:', refreshError);
      }

      // If refresh failed or no new token, redirect to login
      console.log('Token refresh failed or token expired - redirecting to login');
      localStorage.removeItem('snowpack_token');
      refreshAttempts = 0; // Reset counter

      if (!window.location.pathname.includes('/login')) {
        window.location.href = '/login';
      }
    }

    return Promise.reject(error);
  }
);

export default apiClient;

// Export default APIs
export {
  accountsAPI,
  billingCodesAPI,
  projectsAPI,
  ratesAPI,
  timesheetAPI,
  invoicesAPI,
  billsAPI,
  staffAPI
};

// Export individual functions from projects and rates (via wildcard since no conflicts)
export * from './projects';
export * from './rates';

// Export utility functions
export * from './apiUtils';

// Re-export individual account functions
export {
  fetchAccounts, getAccount, createAccount, updateAccount, deleteAccount, inviteUser
};

// Re-export individual billing code functions
export {
  getBillingCodes, getBillingCode, createBillingCode, updateBillingCode, deleteBillingCode,
  getBillingCodesActive as getActiveBillingCodes, // Keep original name in exported API
  getProjectBillingCodes, addBillingCodeToProject, removeBillingCodeFromProject
};

// Re-export individual timesheet functions
export {
  getEntries, getEntry, getUsers, getTimesheetActiveBillingCodes,
  createEntry, updateEntry, deleteEntry, getEntriesByDateRange, getEntriesByUser, getEntriesByProject
};

// Re-export individual invoice functions
export {
  getInvoices, getInvoice, createInvoice, updateInvoice, deleteInvoice,
  changeInvoiceState, getDraftInvoices
};

// Re-export individual bill functions
export {
  getBills, getBill, createBill, updateBill, deleteBill, changeBillState
};

// Re-export staff functions
export {
  fetchStaff, fetchStaffById, createStaff, updateStaff, deleteStaff
} from './staff';

// Re-export tenant functions
export {
  fetchTenant, updateTenant
} from './tenant';

// Export alias for backwards compatibility
export { fetchTenant as getTenant } from './tenant';
