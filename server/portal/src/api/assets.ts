import { portalAPI } from '../api'; // Corrected import path

/**
 * Refreshes the signed URL for a GCS asset in the Portal.
 * @param assetId The ID of the asset to refresh.
 * @returns A promise that resolves to an object containing the new URL and expiration time.
 */
export async function refreshAssetUrl(assetId: number): Promise<{ new_url: string; new_expires_at: string }> {
  try {
    // Assuming portalAPI will be extended to have a method like refreshAssetUrl
    const response = await portalAPI.refreshAssetUrl(assetId); 
    if (response && response.new_url && response.new_expires_at) {
      return response;
    }
    throw new Error('Invalid response structure from refreshAssetUrl API');
  } catch (error) {
    console.error(`Error refreshing asset URL for asset ID ${assetId}:`, error);
    throw error; // Re-throw to be handled by the caller
  }
} 