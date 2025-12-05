<template>
  <TransitionRoot as="template" :show="isOpen">
    <Dialog class="relative z-10" @close="handleClose">
      <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0" enter-to="opacity-100" leave="ease-in duration-200" leave-from="opacity-100" leave-to="opacity-0">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
      </TransitionChild>

      <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
        <div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
          <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95" enter-to="opacity-100 translate-y-0 sm:scale-100" leave="ease-in duration-200" leave-from="opacity-100 translate-y-0 sm:scale-100" leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95">
            <DialogPanel class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6">
              <form @submit.prevent="handleSubmit">
                <div>
                  <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-sage-100">
                    <i class="fas fa-cloud-upload-alt text-2xl text-sage"></i>
                  </div>
                  <div class="mt-3 text-center sm:mt-5">
                    <DialogTitle as="h3" class="text-base font-semibold leading-6 text-gray-900">
                      {{ assetToEdit ? 'Edit Asset' : 'Add New Asset' }}
                    </DialogTitle>
                    <div class="mt-2">
                      <p class="text-sm text-gray-500">
                        {{ assetToEdit ? 'Update the asset details below.' : 'Upload a file or link an external resource.' }}
                      </p>
                    </div>
                  </div>
                </div>

                <div class="mt-5 sm:mt-6 space-y-4">
                  <div>
                    <label for="asset-name" class="block text-sm font-medium text-gray-700">Asset Name</label>
                    <input type="text" v-model="editableAsset.name" id="asset-name" required class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-sage-dark focus:ring-sage-dark sm:text-sm" placeholder="e.g., Project Proposal Q3" />
                  </div>

                  <div>
                    <label for="asset-type" class="block text-sm font-medium text-gray-700">Asset Type</label>
                    <select v-model="selectedAssetType" id="asset-type" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-sage-dark focus:ring-sage-dark sm:text-sm">
                      <option value="file">File Upload</option>
                      <option :value="ASSET_TYPES.GOOGLE_DOC">Google Doc Link</option>
                      <option :value="ASSET_TYPES.GOOGLE_SHEET">Google Sheet Link</option>
                      <option :value="ASSET_TYPES.GOOGLE_SLIDES">Google Slides Link</option>
                      <option :value="ASSET_TYPES.EXTERNAL_LINK">External Link (Other)</option>
                    </select>
                  </div>

                  <div v-if="isLinkType(selectedAssetType)">
                    <label for="asset-url" class="block text-sm font-medium text-gray-700">URL</label>
                    <input type="url" v-model="editableAsset.url" id="asset-url" :required="isLinkType(selectedAssetType)" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-sage-dark focus:ring-sage-dark sm:text-sm" placeholder="https://docs.google.com/... or https://example.com/..." />
                  </div>

                  <div v-if="selectedAssetType === 'file'">
                    <label for="asset-file" class="block text-sm font-medium text-gray-700">File</label>
                    <input type="file" @change="handleFileChange" id="asset-file" :required="selectedAssetType === 'file' && !assetToEdit" class="mt-1 block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-sage-50 file:text-sage-700 hover:file:bg-sage-100" />
                    <p v-if="editableAsset.file" class="mt-1 text-xs text-gray-500">Selected: {{ editableAsset.file.name }}</p>
                  </div>
                  
                  <div class="relative flex items-start">
                    <div class="flex h-6 items-center">
                      <input id="is-public" v-model="editableAsset.is_public" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-sage focus:ring-sage-dark" />
                    </div>
                    <div class="ml-3 text-sm leading-6">
                      <label for="is-public" class="font-medium text-gray-900">Publicly Accessible</label>
                      <p class="text-gray-500 text-xs">If checked, the asset may be accessible via its URL without login.</p>
                    </div>
                  </div>

                </div>

                <div class="mt-5 sm:mt-6 sm:grid sm:grid-flow-row-dense sm:grid-cols-2 sm:gap-3">
                  <button type="submit" :disabled="isLoading" class="inline-flex w-full justify-center rounded-md bg-sage px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-sage-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sage sm:col-start-2 disabled:opacity-50">
                    {{ isLoading ? (assetToEdit ? 'Updating...' : 'Uploading...') : (assetToEdit ? 'Update Asset' : 'Add Asset') }}
                  </button>
                  <button type="button" @click="handleClose" class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:col-start-1 sm:mt-0">
                    Cancel
                  </button>
                </div>
              </form>
            </DialogPanel>
          </TransitionChild>
        </div>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue';
import type { Asset } from '../../types/Asset';
import { createEmptyAsset, ASSET_TYPES } from '../../types/Asset';

const props = defineProps<{
  isOpen: boolean;
  assetToEdit?: Asset | null;
  projectId?: number | null;
  accountId?: number | null;
}>();

const emit = defineEmits<{ (e: 'close'): void; (e: 'save', asset: Asset): void }>();

const editableAsset = ref<Asset>(createEmptyAsset());
const selectedAssetType = ref('file'); // Corresponds to the default option in the select
const isLoading = ref(false);

const isLinkType = (type: string) => {
  return type === ASSET_TYPES.GOOGLE_DOC || 
         type === ASSET_TYPES.GOOGLE_SHEET || 
         type === ASSET_TYPES.GOOGLE_SLIDES || 
         type === ASSET_TYPES.EXTERNAL_LINK;
};

watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    if (props.assetToEdit) {
      editableAsset.value = { ...props.assetToEdit, file: null }; // Reset file on open if editing
      // Attempt to determine selectedAssetType from existing asset_type
      if (isLinkType(props.assetToEdit.asset_type)) {
        selectedAssetType.value = props.assetToEdit.asset_type;
      } else if (props.assetToEdit.asset_type) {
        // If it's a file type (e.g. application/pdf), default to 'file' for the dropdown
        selectedAssetType.value = 'file'; 
      }
    } else {
      editableAsset.value = createEmptyAsset();
      selectedAssetType.value = 'file';
      if (props.projectId) {
        editableAsset.value.project_id = props.projectId;
      }
      if (props.accountId) {
        editableAsset.value.account_id = props.accountId;
      }
    }
  } else {
    // Clear file when closing, especially if not saved
    if (editableAsset.value.file) {
        editableAsset.value.file = null;
    }
  }
});

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement;
  if (target.files && target.files[0]) {
    editableAsset.value.file = target.files[0];
    if (!editableAsset.value.name) { // Auto-fill name if empty
        editableAsset.value.name = target.files[0].name;
    }
  }
};

const handleSubmit = async () => {
  isLoading.value = true;
  try {
    const assetToSave = { ...editableAsset.value };
    
    if (selectedAssetType.value === 'file') {
      if (!assetToSave.file && !props.assetToEdit) { // Require file for new file asset
        alert('Please select a file to upload.');
        isLoading.value = false;
        return;
      }
      assetToSave.asset_type = assetToSave.file?.type || ASSET_TYPES.FILE; // Use actual file type or generic
      // URL will be set by backend for file uploads
      assetToSave.url = ''; 
    } else {
      assetToSave.asset_type = selectedAssetType.value;
      assetToSave.file = null; // Ensure no file object for link types
      if (!assetToSave.url) {
        alert('Please enter a URL for the link.');
        isLoading.value = false;
        return;
      }
    }

    // Remove target project/account if the other is set, ensure only one is primary
    if (assetToSave.project_id && assetToSave.account_id) {
        // Prioritize project if both somehow get set, or clear one based on context
        // For now, let's assume the component is opened with one or the other, not both.
        // If opened from project context, account_id should be null & vice-versa.
        // This logic might need refinement based on how it's invoked.
        if(props.projectId) assetToSave.account_id = null;
        else if(props.accountId) assetToSave.project_id = null;
    }

    emit('save', assetToSave);
    // handleClose(); // Optionally close on save, or let parent handle
  } catch (error) {
    console.error('Error saving asset:', error);
    alert('Failed to save asset. Check console for details.');
  } finally {
    isLoading.value = false;
  }
};

const handleClose = () => {
  emit('close');
  editableAsset.value = createEmptyAsset(); // Reset form
  selectedAssetType.value = 'file';
  isLoading.value = false;
};

</script> 