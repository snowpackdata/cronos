/**
 * Formats a date string into a human-readable format
 * @param dateString - The date string to format
 * @param style - The format style: 'default', 'short', or 'full'
 * @returns The formatted date string or an empty string if the input is invalid
 */
export function formatDate(dateString: string, style: 'default' | 'short' | 'full' = 'default'): string {
  if (!dateString) return '';
  
  try {
    const date = new Date(dateString);
    
    // Check if the date is valid
    if (isNaN(date.getTime())) {
      return '';
    }
    
    let options: Intl.DateTimeFormatOptions;
    
    switch(style) {
      case 'short':
        options = { 
          month: 'numeric', 
          day: 'numeric', 
          year: 'numeric', 
          timeZone: 'UTC' 
        };
        break;
      case 'full':
        options = { 
          weekday: 'long',
          year: 'numeric', 
          month: 'long', 
          day: 'numeric',
          timeZone: 'UTC'
        };
        break;
      case 'default':
      default:
        options = { 
          year: 'numeric', 
          month: 'short', 
          day: 'numeric',
          timeZone: 'UTC'
        };
    }
    
    return date.toLocaleDateString(undefined, options);
  } catch (error) {
    console.error('Error formatting date:', error);
    return '';
  }
} 