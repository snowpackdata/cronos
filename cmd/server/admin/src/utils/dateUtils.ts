/**
 * Utility functions for handling date and time values consistently across the application.
 * These functions ensure consistent handling of dates in UTC to avoid timezone issues.
 */

/**
 * Formats a date for display in the UI
 * @param dateValue - The date to format (string, Date, or null/undefined)
 * @param format - The format to use ('short', 'long', or 'input')
 * @returns Formatted date string or empty string if input is invalid
 */
export function formatDate(dateValue: string | Date | null | undefined, format: 'short' | 'long' | 'input' = 'short'): string {
  if (!dateValue) return '';
  
  try {
    // Convert to Date object if it's a string
    let date: Date;
    if (typeof dateValue === 'string') {
      // For YYYY-MM-DD strings, parse in UTC
      if (dateValue.match(/^\d{4}-\d{2}-\d{2}$/)) {
        const [year, month, day] = dateValue.split('-').map(Number);
        // Create date in UTC
        date = new Date(Date.UTC(year, month - 1, day));
      } else {
        // For other date strings, assume they are in UTC
        date = new Date(dateValue);
      }
    } else {
      date = dateValue;
    }
    
    // Check if date is valid
    if (isNaN(date.getTime())) {
      console.warn('Invalid date value:', dateValue);
      return '';
    }
    
    // Format based on requested format
    switch (format) {
      case 'input':
        // Format for HTML input fields (YYYY-MM-DD)
        // Use UTC to avoid date shifts
        const year = date.getUTCFullYear();
        const month = (date.getUTCMonth() + 1).toString().padStart(2, '0');
        const day = date.getUTCDate().toString().padStart(2, '0');
        return `${year}-${month}-${day}`;
        
      case 'short':
        // Format as MM/DD/YYYY using UTC
        return `${(date.getUTCMonth() + 1).toString().padStart(2, '0')}/${date.getUTCDate().toString().padStart(2, '0')}/${date.getUTCFullYear()}`;
        
      case 'long':
        // Format using toLocaleDateString but with explicit UTC conversion
        const options = { 
          year: 'numeric' as const, 
          month: 'long' as const, 
          day: 'numeric' as const,
          timeZone: 'UTC'
        };
        return date.toLocaleDateString('en-US', options);
      
      default:
        return date.toLocaleDateString('en-US', { timeZone: 'UTC' });
    }
  } catch (error) {
    console.error('Error formatting date:', error);
    return '';
  }
}

/**
 * Parses a date string from the server into a standard format
 * Ensures dates are handled in UTC to prevent timezone issues
 * @param dateString - The date string from the server
 * @returns Normalized date string in YYYY-MM-DD format
 */
export function parseServerDate(dateString: string | null | undefined): string {
  
  if (!dateString) return '';
  
  try {
    // If the string already looks like YYYY-MM-DD, just return it
    if (typeof dateString === 'string' && dateString.match(/^\d{4}-\d{2}-\d{2}$/)) {
      return dateString;
    }
    
    // Parse the date in UTC
    const date = new Date(dateString);
    
    // Check if date is valid
    if (isNaN(date.getTime())) {
      console.warn('Invalid server date value:', dateString);
      return '';
    }
    
    // Format in YYYY-MM-DD using UTC components
    const year = date.getUTCFullYear();
    const month = (date.getUTCMonth() + 1).toString().padStart(2, '0');
    const day = date.getUTCDate().toString().padStart(2, '0');
    
    const formattedDate = `${year}-${month}-${day}`;
    return formattedDate;
  } catch (error) {
    console.error('Error parsing server date:', error);
    return '';
  }
}

/**
 * Format a date string for the server API
 * @param dateStr Date string in YYYY-MM-DD format
 * @returns Date string in YYYY-MM-DD format (with UTC conversion)
 */
export function formatDateForServer(dateStr: string): string {
  
  if (!dateStr) {
    return '';
  }
  
  try {
    // If the input is already in YYYY-MM-DD format from HTML date input
    const dateRegex = /^(\d{4})-(\d{2})-(\d{2})$/;
    const match = dateStr.match(dateRegex);
    
    if (match) {
      return dateStr; // The format is already correct for the server
    }
    
    // For other formats, parse and format to YYYY-MM-DD in UTC
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) {
      console.error('Invalid date:', dateStr);
      // Return current date in UTC as fallback
      const today = new Date();
      return `${today.getUTCFullYear()}-${String(today.getUTCMonth() + 1).padStart(2, '0')}-${String(today.getUTCDate()).padStart(2, '0')}`;
    }
    
    // Format as YYYY-MM-DD using UTC date components
    const year = date.getUTCFullYear();
    const month = String(date.getUTCMonth() + 1).padStart(2, '0');
    const day = String(date.getUTCDate()).padStart(2, '0');
    
    const formatted = `${year}-${month}-${day}`;
    return formatted;
  } catch (error) {
    console.error('Error formatting date for server:', error);
    // Return current date in UTC as fallback
    const today = new Date();
    return `${today.getUTCFullYear()}-${String(today.getUTCMonth() + 1).padStart(2, '0')}-${String(today.getUTCDate()).padStart(2, '0')}`;
  }
}

/**
 * Returns today's date in YYYY-MM-DD format in UTC
 */
export function getTodayFormatted(): string {
  const now = new Date();
  const year = now.getUTCFullYear();
  const month = (now.getUTCMonth() + 1).toString().padStart(2, '0');
  const day = now.getUTCDate().toString().padStart(2, '0');
  
  return `${year}-${month}-${day}`;
}

/**
 * Returns the current date in YYYY-MM-DD format in UTC
 * Alias for getTodayFormatted for better semantic clarity
 */
export function getCurrentDate(): string {
  return getTodayFormatted();
}

/**
 * Returns a date that is a certain number of days in the future or past in UTC
 * @param days - Number of days to add (positive) or subtract (negative)
 * @returns Date string in YYYY-MM-DD format
 */
export function getRelativeDate(days: number): string {
  const date = new Date();
  date.setUTCDate(date.getUTCDate() + days);
  
  const year = date.getUTCFullYear();
  const month = (date.getUTCMonth() + 1).toString().padStart(2, '0');
  const day = date.getUTCDate().toString().padStart(2, '0');
  
  return `${year}-${month}-${day}`;
}

/**
 * Determines if a date is in the past, using UTC comparison
 * @param dateValue - The date to check
 * @returns Boolean indicating if the date is in the past
 */
export function isDateInPast(dateValue: string | Date | null | undefined): boolean {
  if (!dateValue) return false;
  
  try {
    let date: Date;
    
    if (typeof dateValue === 'string') {
      // For YYYY-MM-DD strings, ensure we're parsing in UTC
      if (dateValue.match(/^\d{4}-\d{2}-\d{2}$/)) {
        const [year, month, day] = dateValue.split('-').map(Number);
        date = new Date(Date.UTC(year, month - 1, day));
      } else {
        date = new Date(dateValue);
      }
    } else {
      date = dateValue;
    }
    
    // Set to midnight UTC for comparison
    date.setUTCHours(0, 0, 0, 0);
    
    // Get today at midnight UTC
    const today = new Date();
    today.setUTCHours(0, 0, 0, 0);
    
    return date < today;
  } catch (error) {
    console.error('Error checking if date is in past:', error);
    return false;
  }
} 