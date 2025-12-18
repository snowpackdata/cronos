import axios from 'axios';

// Create axios instance with base configuration
export const api = axios.create({
  headers: {
    'Content-Type': 'application/json',
  },
});

// Loop prevention for token refresh
let refreshAttempts = 0;
const MAX_REFRESH_ATTEMPTS = 1;

// Request interceptor to add the token to headers
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('snowpack_token');
    if (token) {
      config.headers['x-access-token'] = token;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for automatic token refresh and retry
api.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    const originalRequest = error.config;
    
    // Check for authentication errors
    if (error.response && (error.response.status === 401 || error.response.status === 403)) {
      if (refreshAttempts >= MAX_REFRESH_ATTEMPTS) {
        localStorage.removeItem('snowpack_token');
        if (!window.location.pathname.includes('/login')) {
          window.location.href = '/login';
        }
        refreshAttempts = 0;
        return Promise.reject(error);
      }
      
      refreshAttempts++;
      
      try {
        const refreshSuccess = await refreshToken();
        
        if (refreshSuccess) {
          const newToken = localStorage.getItem('snowpack_token');
          if (newToken && originalRequest.headers) {
            originalRequest.headers['x-access-token'] = newToken;
            refreshAttempts = 0;
            return api(originalRequest);
          }
        }
      } catch (refreshError) {
        console.error('Token refresh failed:', refreshError);
      }
      
      localStorage.removeItem('snowpack_token');
      refreshAttempts = 0;
      
      if (!window.location.pathname.includes('/login')) {
        window.location.href = '/login';
      }
    }
    
    return Promise.reject(error);
  }
);

/**
 * Function to refresh token
 */
export const refreshToken = async (): Promise<boolean> => {
  try {
    const response = await api.post('/refresh_token');
    if (response.data && response.data.token) {
      localStorage.setItem('snowpack_token', response.data.token);
      return true;
    }
    return false;
  } catch (error) {
    console.error('Error refreshing token:', error);
    return false;
  }
};

/**
 * Get token from localStorage
 */
export const getToken = (): string | null => {
  return localStorage.getItem('snowpack_token');
};
