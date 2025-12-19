<template>
  <div class="px-4 py-6 sm:px-6 lg:px-8">
    <div class="sm:flex sm:items-center">
      <div class="sm:flex-auto">
        <h1 class="text-base font-semibold leading-6 text-blue">Client Accounts</h1>
      </div>
      <div class="mt-3 sm:ml-16 sm:mt-0 sm:flex-none">
        <button
          type="button"
          @click="openAccountDrawer()"
          class="block rounded-md bg-sage px-2.5 py-1.5 text-center text-xs font-semibold text-white shadow-sm hover:bg-sage-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sage"
        >
          <i class="fas fa-plus-circle mr-1"></i> Create new account
        </button>
      </div>
    </div>
    
    <!-- Account Cards -->
    <div class="mt-4 flow-root">
      <ul role="list" class="grid grid-cols-1 gap-4">
        <li v-for="account in filteredAccounts" :key="account.ID">
          <AccountCard
            :account="account"
            @edit="openAccountDrawer"
            @invite-client="openInviteClientModal"
            @add-asset="openAssetUploaderModal"
            @asset-deleted="handleAssetDeleted"
          />
        </li>
        <li v-if="filteredAccounts.length === 0" class="col-span-full py-5">
          <div class="flex flex-col items-center justify-center p-10 bg-white rounded-lg shadow">
            <i class="fas fa-building text-5xl text-gray-300 mb-4"></i>
            <p class="text-lg font-medium text-gray-dark">No client accounts found</p>
            <p class="text-gray mb-4">Click "Create new account" to add one</p>
          </div>
        </li>
      </ul>
    </div>
    
    <!-- Account Drawer -->
    <AccountDrawer
      :is-open="isAccountDrawerOpen"
      :account-data="selectedAccount"
      @close="closeAccountDrawer"
      @save="saveAccount"
      @delete="handleDeleteFromDrawer"
    />
    
    <!-- Asset Uploader Modal -->
    <AssetUploaderModal 
      :is-open="isAssetUploaderOpen" 
      :account-id="selectedAccountIdForAsset"
      @close="closeAssetUploaderModal" 
      @save="handleSaveAsset"
    />
    
    <!-- Delete Confirmation Modal -->
    <ConfirmationModal
      :show="showDeleteModal"
      title="Delete Account"
      message="Are you sure you want to delete this account? This action cannot be undone and will also remove all projects associated with this account."
      @confirm="deleteAccount"
      @cancel="showDeleteModal = false"
    />

    <!-- Invite Client Modal -->
    <div v-if="showInviteClientModal" class="fixed inset-0 z-50 overflow-y-auto bg-gray-500 bg-opacity-75 transition-opacity" aria-labelledby="modal-title" role="dialog" aria-modal="true">
      <div class="flex min-h-full items-center justify-center p-4 text-center sm:p-0">
        <div class="relative transform overflow-hidden rounded-lg bg-white text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg">
          <div class="bg-white px-4 pb-4 pt-5 sm:p-6 sm:pb-4">
            <div class="sm:flex sm:items-start">
              <div class="mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-indigo-100 sm:mx-0 sm:h-10 sm:w-10">
                <i class="fas fa-user-plus text-indigo-600"></i>
              </div>
              <div class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left">
                <h3 class="text-base font-semibold leading-6 text-gray-900" id="modal-title">
                  Invite New Client to {{ accountToInviteClientTo?.name }}
                </h3>
                <div class="mt-2">
                  <p class="text-sm text-gray-500">
                    Enter the email address of the client you want to invite. They will receive an email to set up their account.
                  </p>
                  <input 
                    type="email" 
                    v-model="newClientEmail"
                    placeholder="client@example.com"
                    class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                  />
                </div>
              </div>
            </div>
          </div>
          <div class="bg-gray-50 px-4 py-3 sm:flex sm:flex-row-reverse sm:px-6">
            <button 
              type="button" 
              @click="handleSendInvite"
              :disabled="!newClientEmail.trim()"
              class="inline-flex w-full justify-center rounded-md bg-sage px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-sage-dark sm:ml-3 sm:w-auto disabled:opacity-50">
              Send Invite
            </button>
            <button 
              type="button" 
              @click="closeInviteClientModal"
              class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto">
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
    <!-- End Invite Client Modal -->
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { accountsAPI, fetchAccounts, createAccount, updateAccount, deleteAccount as deleteAccountAPI } from '../../api';
import AccountDrawer from '../../components/accounts/AccountDrawer.vue';
import AccountCard from '../../components/accounts/AccountCard.vue';
import ConfirmationModal from '../../components/ConfirmationModal.vue';
import AssetUploaderModal from '../../components/assets/AssetUploaderModal.vue';
import { createAsset } from '../../api/assets';
import type { Asset } from '../../types/Asset';
import type { Account } from '../../types/Account';

// State
const accounts = ref<Account[]>([]);
const isAccountDrawerOpen = ref(false);
const selectedAccount = ref<any | null>(null);
const showDeleteModal = ref(false);
const accountToDelete = ref<any | null>(null);

// New state for Asset Uploader
const isAssetUploaderOpen = ref(false);
const selectedAccountIdForAsset = ref<number | null>(null);

// State for Invite Client Modal
const showInviteClientModal = ref(false);
const accountToInviteClientTo = ref<any | null>(null);
const newClientEmail = ref('');

// Filter accounts to show only clients (non-internal)
const filteredAccounts = computed(() => {
  return accounts.value.filter(account => account.type !== 'ACCOUNT_TYPE_INTERNAL');
});

// Fetch data
const loadAccounts = async () => {
  try {
    const fetchedAccounts = await fetchAccounts();
    // Assuming fetchAccounts() now returns accounts with their assets pre-populated
    // The loop below that individually fetched assets for each account has been removed.
    accounts.value = fetchedAccounts || []; // Ensure accounts is an empty array if fetchedAccounts is null/undefined
  } catch (error) {
    console.error('Failed to load accounts:', error);
    accounts.value = []; // Ensure accounts is an empty array on error
  }
};

onMounted(loadAccounts);

// Drawer functions
const openAccountDrawer = (account: any | null = null) => {
  if (account) {
    // Make sure we include both id and ID properties
    selectedAccount.value = {
      ...account,
      id: account.ID, // Include lowercase 'id' for compatibility with the drawer
    };
  } else {
    selectedAccount.value = account;
  }
  isAccountDrawerOpen.value = true;
};

const closeAccountDrawer = () => {
  isAccountDrawerOpen.value = false;
  selectedAccount.value = null;
};

// Save account
const saveAccount = async (accountData: any) => {
  try {
    if (selectedAccount.value && selectedAccount.value.ID) {
      // The accountData should already contain the ID if it's an update
      // Ensure the ID from selectedAccount is part of the payload if not already.
      const payload = { ...accountData, ID: selectedAccount.value.ID };
      await updateAccount(payload); 
    } else {
      await createAccount(accountData);
    }
    await loadAccounts(); // Refresh accounts and their assets
    closeAccountDrawer();
  } catch (error) {
    console.error('Failed to save account:', error);
  }
};

// Delete account
const deleteAccount = async () => {
  if (accountToDelete.value && accountToDelete.value.ID) {
    try {
      await deleteAccountAPI(accountToDelete.value.ID);
      await loadAccounts(); // Refresh accounts and their assets
      showDeleteModal.value = false;
      accountToDelete.value = null;
    } catch (error) {
      console.error('Failed to delete account:', error);
    }
  }
};

const handleDeleteFromDrawer = (accountId: number) => {
  const account = accounts.value.find(acc => acc.ID === accountId);
  if (account) {
    accountToDelete.value = account;
    showDeleteModal.value = true; // Corrected to use showDeleteModal and accountToDelete
  }
};

// Asset Uploader Modal Functions
const openAssetUploaderModal = (accountId: number) => {
  selectedAccountIdForAsset.value = accountId;
  isAssetUploaderOpen.value = true;
};

const closeAssetUploaderModal = () => {
  isAssetUploaderOpen.value = false;
  selectedAccountIdForAsset.value = null;
};

const handleSaveAsset = async (assetData: Asset) => {
  try {
    await createAsset(assetData);
    await loadAccounts(); // Refresh accounts and their assets
    closeAssetUploaderModal();
  } catch (error) {
    console.error('Error saving asset:', error);
  }
};

// Handler for when an asset is deleted from AssetDisplayItem
const handleAssetDeleted = async () => {
  // AssetDisplayItem already shows a confirmation, so we just reload.
  await loadAccounts();
  // Optionally, show a success notification here if desired.
};

// Invite Client Modal Functions
const openInviteClientModal = (account: any) => {
  accountToInviteClientTo.value = account;
  newClientEmail.value = ''; // Clear previous email
  showInviteClientModal.value = true;
};

const closeInviteClientModal = () => {
  showInviteClientModal.value = false;
  accountToInviteClientTo.value = null;
  newClientEmail.value = '';
};

const handleSendInvite = async () => {
  if (!newClientEmail.value.trim() || !accountToInviteClientTo.value) {
    alert('Please enter a valid email and ensure an account is selected.');
    return;
  }
  try {
    await accountsAPI.inviteUser(accountToInviteClientTo.value.ID, newClientEmail.value);
    
    // Invite sent successfully - silently refresh
    closeInviteClientModal();

    // Refresh accounts list to get the latest data including the new pending user
    const updatedAccounts = await fetchAccounts(); 
    accounts.value = updatedAccounts || [];

  } catch (error) {
    console.error('Failed to send invite:', error);
    alert(`Failed to send invite: ${(error as Error).message || 'Unknown error'}`);
  }
};
</script>

<style scoped>
.bg-sage-50 {
  background-color: #F0F4F0;
}
.text-sage-700 {
  color: #2E6E32;
}
</style> 