import type { Asset } from './Asset'; // Import the Asset type

// Defines the structure for a User (Client) associated with an Account.
export interface User {
  ID: string | number; // Unique identifier for the user
  email: string;       // Email of the user
  first_name?: string;  // Optional first name
  last_name?: string;   // Optional last name
  title?: string;       // Optional title (e.g., CEO, Manager)
  status?: string;      // User status, e.g., "Active", "Pending"
  // user_id field was previously here but ID serves as the user identifier
}

// Defines the structure for an Asset, to be imported from Asset.ts
// We are re-stating it here just for conceptual clarity during this step,
// but in Settings.vue it will be imported from portal/src/types/Asset.ts
// export interface Asset {
//   id: string | number;
//   name: string;
//   type: string;
//   status: string;
// }

// Defines the main Account structure for the Settings page.
export interface Account {
  id: string | number;          // Unique identifier for the account
  legal_name?: string;           // Legal name of the account holder
  email?: string;        // Primary contact email for the account
  website?: string;             // Account's website
  address?: string;             // Physical or mailing address
  clients?: User[];           // Array of clients associated with the account
  assets?: Asset[];             // Array of assets associated with the account, explicitly using Asset type
} 