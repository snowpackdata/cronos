import axios from 'axios';

// Changed to a named export so it can be imported by portalService.ts
export const apiClient = axios.create({
  // baseURL: '/api', // Your API base URL, already handled by Vite proxy for dev
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add the token to headers
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('snowpack_token'); // Make sure this key matches what you use
    if (token) {
      config.headers['x-access-token'] = token;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Token refresh function with loop prevention
let refreshAttempts = 0;
const MAX_REFRESH_ATTEMPTS = 1; // Only try once to prevent infinite loops

export const refreshToken = async (): Promise<boolean> => {
  // Prevent infinite loops
  if (refreshAttempts >= MAX_REFRESH_ATTEMPTS) {
    console.log('Max refresh attempts reached - redirecting to login');
    localStorage.removeItem('snowpack_token');
    if (!window.location.pathname.includes('/login')) {
      window.location.href = '/login';
    }
    return false;
  }
  
  refreshAttempts++;
  
  try {
    const token = localStorage.getItem('snowpack_token');
    if (!token) {
      return false;
    }

    const response = await apiClient.post('/refresh_token');
    
    if (response.data && response.data.token) {
      localStorage.setItem('snowpack_token', response.data.token);
      refreshAttempts = 0; // Reset counter on successful refresh
      return true;
    }
    
    return false;
  } catch (error) {
    console.error('Token refresh failed:', error);
    return false;
  }
};

// Enhanced response interceptor with automatic token refresh
apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    // Handle authentication errors (401 Unauthorized, 403 Forbidden)
    if (error.response && (error.response.status === 401 || error.response.status === 403)) {
      console.log(`Authentication error (${error.response.status}) - attempting token refresh`);
      
      // Try to refresh the token before giving up
      const refreshSuccess = await refreshToken();
      
      if (refreshSuccess) {
        // Retry the original request with the new token
        const originalRequest = error.config;
        const newToken = localStorage.getItem('snowpack_token');
        if (newToken) {
          originalRequest.headers['x-access-token'] = newToken;
          return apiClient(originalRequest);
        }
      }
      
      // If refresh failed or no new token, redirect to login
      console.log('Token refresh failed or token expired - redirecting to login');
      localStorage.removeItem('snowpack_token'); 
      refreshAttempts = 0; // Reset counter
      
      // Always redirect to login unless already on a public page
      const publicPaths = ['/login', '/register'];
      const isPublicPage = publicPaths.some(path => window.location.pathname.startsWith(path));
      
      if (!isPublicPage) {
        window.location.href = '/login';
      }
    }
    
    // Handle network errors or other issues that might indicate authentication problems
    if (error.code === 'NETWORK_ERROR' || error.code === 'ERR_NETWORK') {
      console.log('Network error - checking if token might be the issue');
      const token = localStorage.getItem('snowpack_token');
      if (token) {
        // If we have a token but get network errors, it might be expired
        // Try to refresh it as a precaution
        const refreshSuccess = await refreshToken();
        if (!refreshSuccess) {
          localStorage.removeItem('snowpack_token');
          refreshAttempts = 0; // Reset counter
          if (!window.location.pathname.includes('/login')) {
            window.location.href = '/login';
          }
        }
      }
    }
    
    return Promise.reject(error);
  }
);

// Import a
import * as portalService from './portalService';

export const portalAPI = {
  ...portalService,
};

// Optional: Export the apiClient if it needs to be used directly elsewhere,
// otherwise, components can just use the portalAPI object.
// export default apiClient; // Removed default export in favor of named export for clarity

// If you want to keep the default export, you can do it like this:
// export default apiClient; 