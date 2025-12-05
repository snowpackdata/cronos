import axios from 'axios';

// Add TypeScript declaration for import.meta.env
declare interface ImportMeta {
  readonly env: {
    readonly MODE: string
    readonly BASE_URL: string
    readonly DEV: boolean 
    readonly PROD: boolean
    readonly [key: string]: string | boolean | undefined
  }
}

/**
 * Configure Axios with default settings for the application
 */

// During development, the API is proxied via Vite's server.proxy
// In production, the API is served from the same domain
// So we use relative URLs instead of a fixed base URL
axios.defaults.baseURL = '';

/**
 * Gets auth token from either localStorage or cookies
 * This ensures token availability across different contexts
 * @returns {string} The authentication token or empty string if not found
 */
function getAuthToken(): string {
  // First try to get the token from localStorage
  let token = localStorage.getItem('snowpack_token');
  
  // If not found in localStorage, try to get it from cookies
  if (!token) {
    const cookies = document.cookie.split(';');
    for (let i = 0; i < cookies.length; i++) {
      const cookie = cookies[i].trim();
      if (cookie.startsWith('x-access-token=')) {
        token = cookie.substring('x-access-token='.length);
        break;
      }
    }
    
    // If token found in cookie but not in localStorage, store it in localStorage for future use
    if (token) {
      localStorage.setItem('snowpack_token', token);
    }
  }
  
  // Debugging for production environments - log token existence but not the actual token
  const currentUrl = window.location.href;
  if (!token && !currentUrl.includes('login') && !currentUrl.includes('register')) {
    console.warn('Auth token not found in localStorage or cookies. This may cause API requests to fail.');
    
    // Log cookie and localStorage state for debugging (without exposing sensitive data)
    console.log('Cookies exist:', document.cookie.length > 0);
    console.log('localStorage available:', typeof localStorage !== 'undefined');
    
    // Attempt to re-sync from the other storage method on the next navigation
    window.addEventListener('beforeunload', () => {
      // Clear invalid token state to force re-authentication
      localStorage.removeItem('snowpack_token');
    });
  }
  
  return token || '';
}

// Add request interceptor to include the token with every request
axios.interceptors.request.use(config => {
  // Don't warn about missing tokens on login/register pages
  const currentPath = window.location.pathname;
  const isAuthPage = currentPath.includes('/login') || currentPath.includes('/register');
  const isDevelopment = import.meta.env.DEV || 
                     window.location.hostname === 'localhost' || 
                     window.location.hostname === '127.0.0.1';
  
  // Get token using the new helper function
  let token = getAuthToken();

  // Add a timestamp parameter to prevent caching issues in production
  // This can help with requests that might be cached by browsers or CDNs
  if (config.params) {
    config.params._ts = new Date().getTime();
  } else {
    config.params = { _ts: new Date().getTime() };
  }

  if (token) {
    // Use x-access-token header as expected by the backend
    if (config.headers) {
      config.headers['x-access-token'] = token;
      
      // Log header value in development without showing full token
      if (isDevelopment) {
        console.log('Request includes auth token: Yes (token length:', token.length, ')');
      }
    }
  } else if (!isAuthPage) {
    // Missing token warning retained for debugging
    console.warn('No authentication token found. Request may fail if authentication is required.');
    
    // Try to redirect to login if we're not on an auth page and no token exists
    if (!isDevelopment && !isAuthPage) {
      console.warn('Token missing in production - auto redirecting to login');
      setTimeout(() => {
        window.location.href = '/login';
      }, 500);
    }
  }
  
  return config;
}, error => {
  return Promise.reject(error);
});

// Add response interceptor for better error handling
axios.interceptors.response.use(response => {
  
  // Check if the response is JSON when expected
  const contentType = response.headers['content-type'];
  if (contentType && contentType.includes('application/json')) {
    return response;
  } else if (response.config.url && !response.config.url.includes('html') && !response.config.url.includes('css')) {
    // Only warn for endpoints that should return JSON
    console.warn(`Response from ${response.config.url} is not JSON (${contentType})`);
    // Still return the response so the app can handle it
    return response;
  }
  
  return response;
}, error => {
  console.error('Response error:', error.message);
  if (error.response) {
    console.error('Status:', error.response.status);
    console.error('URL:', error.config?.url);
    console.error('Response data:', error.response.data);
    const contentType = error.response.headers && error.response.headers['content-type'];
    console.error('Content-Type:', contentType);
  } else if (error.request) {
    // Request was made but no response received
    console.error('No response received:', error.request);
  } else {
    // Something else caused the error
    console.error('Error setting up request:', error.message);
  }
  return Promise.reject(error);
});

export default axios; 