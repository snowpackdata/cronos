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
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import type { Asset } from '../../types/Asset.ts';

const props = defineProps<{
  asset: Asset;
  projectId?: number;   
  accountId?: number;   
  isReadOnly?: boolean; 
}>();

const isLoading = ref(false);

// Applying the updated getFileIcon function with colors
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
    iconClass = 'fas fa-file-download';
    colorClass = 'text-indigo-600';
  }
  return `${iconClass} ${colorClass}`;
};

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
      const downloadUrl = `/api/portal/assets/${props.asset.ID}/download`;
      
      // Fetch with authentication headers
      const token = localStorage.getItem('portal_token');
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

</script>

<style scoped>
/* All styling is primarily handled by Tailwind CSS utility classes in the template. */
</style> 