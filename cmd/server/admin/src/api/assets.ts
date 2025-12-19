import type { Asset } from '../types/Asset';
import { api as apiClient } from './apiUtils';
import { ASSET_TYPES } from '../types/Asset';

const ASSET_BASE_URL = '/assets'; // Define your actual API base path for assets

/**
 * Fetches a list of assets, optionally filtered by project or account ID.
 * @param projectId Optional project ID to filter assets.
 * @param accountId Optional account ID to filter assets.
 * @returns A promise that resolves to an array of Assets.
 */
export const fetchAssets = async (projectId?: number, accountId?: number): Promise<Asset[]> => {
  try {
    const params = new URLSearchParams();
    if (projectId) params.append('project_id', projectId.toString());
    if (accountId) params.append('account_id', accountId.toString());

    const response = await apiClient.get(`${ASSET_BASE_URL}?${params.toString()}`);
    return response.data as Asset[];
  } catch (error) {
    console.error('Error fetching assets:', error);
    throw error;
  }
};

/**
 * Fetches a single asset by its ID.
 * @param assetId The ID of the asset to fetch.
 * @returns A promise that resolves to an Asset.
 */
export const getAssetById = async (assetId: number): Promise<Asset> => {
  try {
    const response = await apiClient.get(`${ASSET_BASE_URL}/${assetId}`);
    return response.data as Asset;
  } catch (error) {
    console.error(`Error fetching asset with ID ${assetId}:`, error);
    throw error;
  }
};

/**
 * Creates a new asset. If the asset includes a file, it will be sent as FormData.
 * Otherwise, it will be sent as JSON.
 * @param assetData The asset data to create. Includes an optional 'file' property for uploads.
 * @returns A promise that resolves to the created Asset.
 */
export const createAsset = async (assetData: Asset): Promise<Asset> => {
  try {
    let response;
    let apiUrl = ASSET_BASE_URL;

    if (assetData.project_id) {
      apiUrl = `/api/projects/${assetData.project_id}/assets`;
    } else if (assetData.account_id) {
      apiUrl = `/api/accounts/${assetData.account_id}/assets`;
    } else {
      console.error('Asset must be associated with a project or an account.');
      throw new Error('Asset must be associated with a project or an account.');
    }

    if (assetData.file && assetData.asset_type !== ASSET_TYPES.GOOGLE_DOC && assetData.asset_type !== ASSET_TYPES.GOOGLE_SHEET && assetData.asset_type !== ASSET_TYPES.GOOGLE_SLIDES && assetData.asset_type !== ASSET_TYPES.EXTERNAL_LINK) {
      const formData = new FormData();
      formData.append('file', assetData.file);
      formData.append('name', assetData.name);
      formData.append('asset_type', assetData.asset_type);
      formData.append('is_public', String(assetData.is_public));
      if (assetData.project_id) formData.append('project_id', String(assetData.project_id));
      if (assetData.account_id) formData.append('account_id', String(assetData.account_id));
      if (assetData.expires_at) formData.append('expires_at', assetData.expires_at);

      response = await apiClient.post(apiUrl, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
    } else {
      const payload = { ...assetData };
      delete payload.file;
      response = await apiClient.post(apiUrl, payload);
    }
    return response.data as Asset;
  } catch (error) {
    console.error('Error creating asset:', error);
    throw error;
  }
};

/**
 * Updates an existing asset.
 * @param assetId The ID of the asset to update.
 * @param assetData The updated asset data.
 * @returns A promise that resolves to the updated Asset.
 */
export const updateAsset = async (assetId: number, assetData: Partial<Asset>): Promise<Asset> => {
  try {
    const apiUrl = `${ASSET_BASE_URL}/${assetId}`;
    const payload = { ...assetData };
    if (payload.file) {
        console.warn("Updating an asset with a new file is not directly supported by this basic updateAsset function. Consider a dedicated upload or re-create flow.");
        delete payload.file;
    }
    const response = await apiClient.put(apiUrl, payload);
    return response.data as Asset;
  } catch (error) {
    console.error(`Error updating asset with ID ${assetId}:`, error);
    throw error;
  }
};

/**
 * Deletes an asset by its ID (general asset delete, not project-specific).
 * If you need project-specific delete, use deleteProjectAsset.
 * @param assetId The ID of the asset to delete.
 * @returns A promise that resolves when the asset is deleted.
 */
export const deleteAsset = async (assetId: number): Promise<void> => {
  try {
    const apiUrl = `${ASSET_BASE_URL}/${assetId}`;
    await apiClient.delete(apiUrl);
  } catch (error) {
    console.error(`Error deleting asset with ID ${assetId}:`, error);
    throw error;
  }
};

// Function to refresh a GCS asset's signed URL
export const refreshAssetUrl = async (assetId: number): Promise<{ new_url: string; new_expires_at: string }> => {
  const response = await apiClient.post<{ new_url: string; new_expires_at: string }>(`/api/assets/${assetId}/refresh-url`);
  return response.data;
};

/**
 * Deletes a specific asset associated with a project.
 * @param projectId The ID of the project.
 * @param assetId The ID of the asset to delete.
 * @returns A promise that resolves when the project asset is deleted.
 */
export const deleteProjectAsset = async (projectId: number, assetId: number): Promise<void> => {
  await apiClient.delete(`projects/${projectId}/assets/${assetId}`);
}; 