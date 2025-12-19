import axios from 'axios';

/**
 * Normalize API URLs to ensure they always include the /api prefix
 */
const normalizeApiUrl = (endpoint: string): string => {
  if (endpoint.startsWith('/api/')) {
    return endpoint;
  }
  if (endpoint.startsWith('/')) {
    return `/api${endpoint}`;
  }
  return `/api/${endpoint}`;
};

// Create axios instance with base configuration
export const api = axios.create({
  headers: {
    'Content-Type': 'application/json',
  },
});

// Loop prevention for token refresh
let refreshAttempts = 0;
const MAX_REFRESH_ATTEMPTS = 1; // Only try once to prevent infinite loops

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
        const refreshSuccess = await refreshToken();
        
        if (refreshSuccess) {
          // Retry the original request with the new token
          const newToken = localStorage.getItem('snowpack_token');
          if (newToken && originalRequest.headers) {
            originalRequest.headers['x-access-token'] = newToken;
            refreshAttempts = 0; // Reset counter on successful refresh
            return api(originalRequest);
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
    
    // Handle network errors or other issues that might indicate authentication problems
    if (error.code === 'NETWORK_ERROR' || error.code === 'ERR_NETWORK') {
      console.log('Network error - checking if token might be the issue');
      const token = localStorage.getItem('snowpack_token');
      if (token) {
        // If we have a token but get network errors, it might be expired
        // Try to refresh it as a precaution
        try {
          const refreshSuccess = await refreshToken();
          if (!refreshSuccess) {
            localStorage.removeItem('snowpack_token');
            refreshAttempts = 0; // Reset counter
            if (!window.location.pathname.includes('/login')) {
              window.location.href = '/login';
            }
          }
        } catch (refreshError) {
          console.error('Token refresh failed on network error:', refreshError);
        }
      }
    }
    
    return Promise.reject(error);
  }
);

// Generic fetch all items of a resource
export async function fetchAll<T>(endpoint: string, params?: Record<string, any>): Promise<T[]> {
  const normalizedUrl = normalizeApiUrl(endpoint);
  const response = await api.get<T[]>(normalizedUrl, { params });

  // Special case for bills endpoint debugging
  if (endpoint === 'bills') {
  }

  return response.data;
}

// Generic fetch item by ID
export async function fetchById<T>(endpoint: string, id: number): Promise<T> {
  const normalizedUrl = normalizeApiUrl(`${endpoint}/${id}`);
  const response = await api.get<T>(normalizedUrl);
  return response.data;
}

// Generic create
export async function create<T>(endpoint: string, data: any): Promise<T> {
  const normalizedUrl = normalizeApiUrl(`${endpoint}/0`);
  const response = await api.post<T>(normalizedUrl, data);
  return response.data;
}

// Generic update
export async function update<T>(endpoint: string, id: number, data: any): Promise<T> {
  if (!id || isNaN(id) || id <= 0) {
    throw new Error(`Invalid ID for update: ${id}`);
  }
  
  const normalizedUrl = normalizeApiUrl(`${endpoint}/${id}`);
  
  try {
    const response = await api.put<T>(normalizedUrl, data);
    return response.data;
  } catch (error: any) {
    // Enhanced error handling with specific messages for different status codes
    if (error.response) {
      const status = error.response.status;
      if (status === 404) {
        console.error(`Entity not found: ${endpoint}/${id}. The entry may have been deleted or is in a state that doesn't allow updates.`);
        throw new Error(`Not Found: The entry with ID ${id} does not exist or cannot be updated in its current state.`);
      } else if (status === 403) {
        console.error(`Permission denied: ${endpoint}/${id}`);
        throw new Error(`Permission denied: You don't have permission to update this entry.`);
      } else if (status === 409) {
        // Handle the 409 Conflict for entries that can't be edited due to state
        const errorData = error.response.data || {};
        const errorMessage = errorData.error || 'The entry cannot be updated in its current state';
        console.error(`Conflict error for ${endpoint}/${id}:`, errorMessage);
        throw new Error(`Cannot update: ${errorMessage}`);
      }
    }
    console.error(`Error updating ${endpoint}/${id}:`, error.message);
    throw error;
  }
}

// Generic delete
export async function remove(endpoint: string, id: number): Promise<void> {
  const normalizedUrl = normalizeApiUrl(`${endpoint}/${id}`);
  await api.delete(normalizedUrl);
}

// Helper type for form data values
type FormDataValue = string | Blob | File;

// Create with FormData
export async function createWithFormData<T>(endpoint: string, data: FormData | Record<string, any>): Promise<T> {
  const normalizedUrl = normalizeApiUrl(`${endpoint}/0`);
  let formData: FormData;
  
  // If data is already FormData, use it directly
  if (data instanceof FormData) {
    formData = data;
  } else {
    // Otherwise, create a new FormData object from the record
    formData = new FormData();
    Object.entries(data).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        // Convert value to string if it's not a valid FormData value type
        if (typeof value === 'object' && !(value instanceof Blob) && !(value instanceof File)) {
          formData.append(key, JSON.stringify(value));
        } else {
          formData.append(key, value as FormDataValue);
        }
      }
    });
  }

  // When using FormData, let the browser set the Content-Type to include the boundary
  const response = await api.post<T>(normalizedUrl, formData, {
    headers: {
      'Content-Type': undefined // Let browser set this with the proper boundary
    }
  });
  
  return response.data;
}

// Update with FormData
export async function updateWithFormData<T>(endpoint: string, id: number, data: FormData | Record<string, any>): Promise<T> {
  const normalizedUrl = normalizeApiUrl(`${endpoint}/${id}`);
  let formData: FormData;
  
  // If data is already FormData, use it directly
  if (data instanceof FormData) {
    formData = data;
  } else {
    // Otherwise, create a new FormData object from the record
    formData = new FormData();
    Object.entries(data).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        // Convert value to string if it's not a valid FormData value type
        if (typeof value === 'object' && !(value instanceof Blob) && !(value instanceof File)) {
          formData.append(key, JSON.stringify(value));
        } else {
          formData.append(key, value as FormDataValue);
        }
      }
    });
  }
  
  // When using FormData, let the browser set the Content-Type to include the boundary
  const response = await api.put<T>(normalizedUrl, formData, {
    headers: {
      'Content-Type': undefined // Let browser set this with the proper boundary
    }
  });
  
  return response.data;
}

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