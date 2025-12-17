<template>
  <div class="asset-item flex items-center justify-between py-1 px-2 bg-slate-50 hover:bg-slate-100 border border-slate-200 rounded-md text-xs">
    <div class="flex items-center overflow-hidden mr-2">
      <i :class="getFileIcon(asset.asset_type || asset.content_type)" class="fa-lg mr-2 flex-shrink-0" :title="asset.asset_type || asset.content_type || 'File'"></i>
      <a 
        href="#" 
        @click.prevent="handleAssetNameClick" 
        class="font-medium text-blue-600 hover:text-blue-800 hover:underline truncate"
        :title="asset.name"
      >
        <span v-if="isLoading" class="italic text-gray-500">Loading...</span>
        <span v-else>{{ asset.name }}</span>
      </a>
    </div>
    <div class="flex items-center space-x-2 flex-shrink-0">
      <span 
        v-if="!props.isReadOnly && isRefreshableSource && timeUntilExpiration && timeUntilExpiration !== 'Invalid date' && timeUntilExpiration !== 'Date error'" 
        class="text-slate-500 text-[10px] italic whitespace-nowrap"
        :title="`URL Expires: ${props.asset.expires_at ? new Date(props.asset.expires_at).toLocaleString() : 'N/A'}`"
      >
        Expires {{ timeUntilExpiration }}
      </span>
      <button 
        @click="confirmDelete" 
        class="text-red-500 hover:text-red-700 p-1 rounded-full hover:bg-red-100 transition-colors duration-150 ease-in-out"
        title="Delete Asset"
        v-if="!props.isReadOnly && (props.projectId || props.accountId)" 
      >
        <i class="fas fa-trash-alt"></i>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { formatDistanceToNow, parseISO } from 'date-fns';
import { deleteProjectAsset, deleteAsset as apiDeleteAsset } from '../../api/assets';
import type { Asset } from '../../types/Asset';

const props = defineProps<{
  asset: Asset;
  projectId?: number;   // For admin project context
  accountId?: number;   // For admin account context (delete functionality pending API)
  isReadOnly?: boolean; // For portal (view-only) context
}>();

const emit = defineEmits(['delete-asset', 'asset-updated']);

const isLoading = ref(false);

const getFileIcon = (fileTypeInput: string | null | undefined): string => {
  const type = (fileTypeInput || '').toLowerCase();
  let iconClass = 'fas fa-file'; // Default icon
  let colorClass = 'text-gray-500'; // Default color

  if (!type) {
    // Keep default icon and color for unknown or non-specified types
  } else if (type.startsWith('image/')) {
    iconClass = 'fas fa-file-image';
    colorClass = 'text-purple-600'; 
  } else if (type.includes('pdf')) {
    iconClass = 'fas fa-file-pdf';
    colorClass = 'text-red-600';
  } else if (type.includes('application/vnd.google-apps.document') || type.includes('wordprocessingml.document') || type.includes('msword')) {
    iconClass = 'fas fa-file-word';
    colorClass = 'text-blue-600';
  } else if (type.includes('application/vnd.google-apps.spreadsheet') || type.includes('spreadsheetml.sheet') || type.includes('ms-excel')) {
    iconClass = 'fas fa-file-excel';
    colorClass = 'text-green-600';
  } else if (type.includes('application/vnd.google-apps.presentation') || type.includes('presentationml.presentation') || type.includes('ms-powerpoint')) {
    iconClass = 'fas fa-file-powerpoint';
    colorClass = 'text-orange-500'; 
  } else if (type.includes('application/zip') || type.includes('application/x-rar-compressed') || type.includes('application/x-tar') || type.includes('application/x-7z-compressed')) {
    iconClass = 'fas fa-file-archive';
    colorClass = 'text-yellow-600'; 
  } else if (type.startsWith('text/csv')) {
    iconClass = 'fas fa-file-csv';
    colorClass = 'text-lime-600';
  } else if (type.startsWith('text/')) {
    iconClass = 'fas fa-file-alt';
    colorClass = 'text-slate-600';
  } else if (type.startsWith('video/')) {
    iconClass = 'fas fa-file-video';
    colorClass = 'text-pink-600'; 
  } else if (type.startsWith('audio/')) {
    iconClass = 'fas fa-file-audio';
    colorClass = 'text-sky-600'; 
  } else if (type.includes('application/octet-stream') || type === 'file') {
    // More generic binary or placeholder 'file' type
    iconClass = 'fas fa-file-download'; // Or 'fa-file-binary' if you have it
    colorClass = 'text-indigo-600';
  }
  // Note: The base 'fas fa-file' with 'text-gray-500' is the ultimate fallback if no conditions are met.

  return `${iconClass} ${colorClass}`; // Combine icon and color class
};

const isRefreshableSource = computed(() => {
  // Check if the asset source URL looks like a GCS URL or if it has an expiry date.
  // This implies it might be a signed URL that can be refreshed.
  return (props.asset.url || '').includes('storage.googleapis.com') || !!props.asset.expires_at;
});

const handleAssetNameClick = async () => {
  if (!props.asset) {
    console.error('Asset is undefined');
    return;
  }

  isLoading.value = true;
  try {
    // For GCS-stored assets (has gcs_object_path), use the download proxy endpoint
    // This provides a clean URL and forces download instead of browser display
    if (props.asset.gcs_object_path) {
      const downloadUrl = `/api/assets/${props.asset.ID}/download`;
      
      // Fetch with authentication headers
      const token = localStorage.getItem('snowpack_token');
      const response = await fetch(downloadUrl, {
        headers: {
          'x-access-token': token || ''
        }
      });
      
      if (!response.ok) {
        throw new Error(`Download failed: ${response.statusText}`);
      }
      
      // Create blob and download
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = props.asset.name || `asset_${props.asset.ID}`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } else if (props.asset.url) {
      // For external links (Google Docs, etc.), open directly
      window.open(props.asset.url, '_blank');
    } else {
      console.error('Asset URL is missing for asset:', props.asset.name);
    }

  } catch (error) {
    console.error('Error opening asset:', error);
  } finally {
    isLoading.value = false;
  }
};

const confirmDelete = async () => {
  if (props.isReadOnly) return;

  if (window.confirm("Are you sure you want to delete this file?")) {
    try {
      if (props.projectId && props.asset.ID) {
        await deleteProjectAsset(props.projectId, props.asset.ID);
        emit('delete-asset', props.asset.ID);
      } else if (props.accountId && props.asset.ID) {
        await apiDeleteAsset(props.asset.ID);
        emit('delete-asset', props.asset.ID);
      } else {
        console.warn('Cannot delete asset: Context (Project ID or Account ID) or Asset ID is missing for asset:', props.asset.name);
        alert('Cannot delete asset: Context or Asset ID not provided.');
      }
    } catch (error) {
      console.error("Failed to delete asset:", error);
      alert(`Failed to delete asset: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
};

const timeUntilExpiration = computed(() => {
  if (!props.asset.expires_at) return null;
  try {
    const expiryDate = parseISO(props.asset.expires_at);
    if (isNaN(expiryDate.getTime())) {
      return 'Invalid date';
    }
    return formatDistanceToNow(expiryDate, { addSuffix: true });
  } catch (e) {
    console.error('Error parsing expiry date for asset:', props.asset.name, props.asset.expires_at, e);
    return 'Date error';
  }
});

</script>

<style scoped>
/* All styling is primarily handled by Tailwind CSS utility classes in the template. */
/* This block is here for any truly component-specific overrides if necessary. */
</style> 