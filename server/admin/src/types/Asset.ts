/**
 * Interface representing an Asset in the system.
 * Assets are records of external resources, typically stored in GCS or linked via URL.
 * They can be associated with a Project or an Account.
 */
import type { Project } from './Project';
import type { Account } from './Account';

export interface Asset {
  ID: number;
  project_id?: number | null; // Optional foreign key
  project?: Project;          // Optional associated project
  account_id?: number | null; // Optional foreign key
  account?: Account;          // Optional associated account
  asset_type: string;         // Type of asset (e.g., 'pdf', 'excel', 'google_doc', 'image')
  name: string;               // Display name of the asset
  url: string;                // URL to access the asset (either GCS public URL or external link)
  is_public: boolean;         // Whether the asset is publicly accessible without authentication

  // GCS specific fields (optional)
  bucket_name?: string | null;
  content_type?: string | null; // MIME type
  size?: number | null;         // Size in bytes
  checksum?: string | null;     // For data integrity (e.g., MD5 or CRC32c)
  upload_status?: string | null;// e.g., 'pending', 'completed', 'failed'
  uploaded_by?: number | null;  // User ID of the uploader
  uploaded_at?: string | null;  // ISO 8601 timestamp
  expires_at?: string | null;   // ISO 8601 timestamp for GCS signed URLs or asset lifecycle
  version?: number | null;      // Version number of the asset in GCS
  gcs_object_path?: string | null; // Actual GCS object path, e.g., assets/projects/1/file.txt

  // Frontend specific helper fields (optional)
  file?: File | null; // Used when uploading a new file
}

/**
 * Creates a new empty Asset object with default values.
 * @returns A new Asset object.
 */
export function createEmptyAsset(): Asset {
  return {
    ID: 0,
    project_id: null,
    account_id: null,
    asset_type: 'file', // Default to 'file', can be changed by user
    name: '',
    url: '',
    is_public: false, // Default to private
    
    bucket_name: null,
    content_type: null,
    size: null,
    checksum: null,
    upload_status: 'pending',
    uploaded_by: null,
    uploaded_at: null,
    expires_at: null,
    version: 1,
    gcs_object_path: null,
    file: null,
  };
}

/**
 * Constants for common asset types.
 * This can be expanded as more types are supported.
 */
export const ASSET_TYPES = {
  PDF: 'application/pdf',
  DOCX: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  XLSX: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  CSV: 'text/csv',
  PNG: 'image/png',
  JPEG: 'image/jpeg',
  GOOGLE_DOC: 'application/vnd.google-apps.document',
  GOOGLE_SHEET: 'application/vnd.google-apps.spreadsheet',
  GOOGLE_SLIDES: 'application/vnd.google-apps.presentation',
  EXTERNAL_LINK: 'text/uri-list', // For generic web links
  FILE: 'file' // A generic file type, specific content_type will be set from uploaded file
};

/**
 * Upload status constants for assets.
 */
export const ASSET_UPLOAD_STATUS = {
  PENDING: 'pending',
  UPLOADING: 'uploading',
  COMPLETED: 'completed',
  FAILED: 'failed',
  PROCESSING: 'processing' // e.g., backend is still processing after upload
}; 