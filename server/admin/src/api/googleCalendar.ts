import axios from 'axios';

// Calendar event interface matching backend CalendarEvent structure
export interface CalendarEvent {
  id: string;
  summary: string;
  description: string;
  start: string;
  end: string;
}

// Auth status response interface
interface AuthStatusResponse {
  connected: boolean;
  needs_reauth: boolean;
}

// Auth URL response interface
interface AuthURLResponse {
  auth_url: string;
}

// Google Calendar API service
const googleCalendarAPI = {
  /**
   * Get the OAuth URL for Google Calendar authorization
   * Opens a popup window for user to authorize
   * @returns Promise that resolves when authorization is complete
   */
  async authorize(): Promise<void> {
    try {
      const token = localStorage.getItem('snowpack_token');
      const response = await axios.post<AuthURLResponse>(
        '/api/google/auth/url',
        {},
        {
          headers: {
            'x-access-token': token
          }
        }
      );

      const authUrl = response.data.auth_url;

      // Open popup window for OAuth flow
      const popup = window.open(
        authUrl,
        'Google Calendar Authorization',
        'width=600,height=700,left=200,top=100'
      );

      // Listen for authorization completion message
      return new Promise((resolve, reject) => {
        const messageHandler = (event: MessageEvent) => {
          if (event.data.type === 'google_auth_success') {
            window.removeEventListener('message', messageHandler);
            resolve();
          }
        };

        window.addEventListener('message', messageHandler);

        // Check if popup was closed without completing auth
        const checkClosed = setInterval(() => {
          if (popup && popup.closed) {
            clearInterval(checkClosed);
            window.removeEventListener('message', messageHandler);
            reject(new Error('Authorization window was closed'));
          }
        }, 500);
      });
    } catch (error) {
      console.error('Failed to get authorization URL:', error);
      throw error;
    }
  },

  /**
   * Check if user has connected Google Calendar
   * @returns Promise with connection status
   */
  async getAuthStatus(): Promise<AuthStatusResponse> {
    try {
      const token = localStorage.getItem('snowpack_token');
      const response = await axios.get<AuthStatusResponse>(
        '/api/google/auth/status',
        {
          headers: {
            'x-access-token': token
          }
        }
      );
      return response.data;
    } catch (error) {
      console.error('Failed to get auth status:', error);
      throw error;
    }
  },

  /**
   * Disconnect Google Calendar
   * @returns Promise that resolves when disconnected
   */
  async disconnect(): Promise<void> {
    try {
      const token = localStorage.getItem('snowpack_token');
      await axios.delete('/api/google/auth/disconnect', {
        headers: {
          'x-access-token': token
        }
      });
    } catch (error) {
      console.error('Failed to disconnect Google Calendar:', error);
      throw error;
    }
  },

  /**
   * Fetch calendar events for a date range
   * @param startDate - Start date (YYYY-MM-DD)
   * @param endDate - End date (YYYY-MM-DD)
   * @returns Promise with array of calendar events
   */
  async getCalendarEvents(startDate: string, endDate: string): Promise<CalendarEvent[]> {
    try {
      const token = localStorage.getItem('snowpack_token');
      const response = await axios.get<CalendarEvent[]>(
        '/api/google/calendar/events',
        {
          params: {
            start_date: startDate,
            end_date: endDate
          },
          headers: {
            'x-access-token': token
          }
        }
      );
      return response.data;
    } catch (error) {
      console.error('Failed to fetch calendar events:', error);
      throw error;
    }
  }
};

export default googleCalendarAPI;

