<template>
  <div class="p-4 md:p-6 lg:p-8 bg-white min-h-screen">
    <h1 class="text-3xl font-bold text-gray-800 mb-6">Account Settings</h1>

    <!-- Loading and Error States -->
    <div v-if="loading" class="text-center py-10">
      <p class="text-lg text-gray-600">Loading account data...</p>
    </div>
    <div v-if="error" class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-6" role="alert">
      <strong class="font-bold">Error!</strong>
      <span class="block sm:inline"> {{ error }}</span>
    </div>

    <div v-if="!loading && !error && account" class="space-y-6">
      <!-- Account Details Section -->
      <div class="bg-white p-4 md:p-5 rounded-lg shadow-xl">
        <h2 class="text-xl font-semibold text-gray-700 mb-3">Account Information</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <p class="text-sm text-gray-500">Legal Name</p>
            <p class="text-base text-gray-900">{{ account.legal_name || 'N/A' }}</p>
          </div>
          <div>
            <p class="text-sm text-gray-500">Contact Email</p>
            <p class="text-base text-gray-900">{{ account.email || 'N/A' }}</p>
          </div>
          <div>
            <p class="text-sm text-gray-500">Website</p>
            <p class="text-base text-gray-900">
              <a v-if="account.website" :href="account.website" target="_blank" class="text-blue-500 hover:underline">{{ account.website }}</a>
              <span v-else>N/A</span>
            </p>
          </div>
          <div>
            <p class="text-sm text-gray-500">Address</p>
            <p class="text-base text-gray-900">{{ account.address || 'N/A' }}</p>
          </div>
          <!-- Add more fields as necessary -->
        </div>
      </div>

      <!-- Attached Clients Section -->
      <div class="bg-white p-4 md:p-5 rounded-lg shadow-xl">
        <h2 class="text-xl font-semibold text-gray-700 mb-3">Users</h2>
        <div v-if="!account.clients || account.clients.length === 0" class="text-sm text-gray-500 py-3">
          No clients attached to this account.
        </div>
        <ul v-else class="space-y-2">
          <li v-for="client in account.clients" :key="client.ID" class="p-2 border border-gray-200 rounded-md hover:shadow-sm">
            <p class="text-sm font-medium text-gray-800">
              <span v-if="client.first_name || client.last_name">
                {{ client.first_name }} {{ client.last_name }}
              </span>
              <span v-else>
                User ID: {{ client.ID }}
              </span>
            </p>
            <p class="text-xs text-gray-600">Email: {{ client.email }}</p>
            <p v-if="client.title" class="text-xs text-gray-500">Title: {{ client.title }}</p>
            <p v-if="client.status" class="text-xs mt-1">
              Status: 
              <span :class="{
                'bg-green-100 text-green-700': client.status === 'Active',
                'bg-yellow-100 text-yellow-700': client.status === 'Pending',
                'bg-gray-100 text-gray-700': client.status !== 'Active' && client.status !== 'Pending'
              }" class="px-2 py-0.5 rounded-full font-medium">
                {{ client.status }}
              </span>
            </p>
          </li>
        </ul>
      </div>

      <!-- Assets Section -->
      <div class="bg-white p-4 md:p-5 rounded-lg shadow-xl">
        <h2 class="text-xl font-semibold text-gray-700 mb-3">Assets</h2>
        <div v-if="!account.assets || account.assets.length === 0" class="text-sm text-gray-500 py-3">
          No assets attached to this account.
        </div>
        <ul v-else class="space-y-1">
          <AssetDisplayItem
            v-for="asset in account.assets"
            :key="asset.ID"
            :asset="asset"
            :is-read-only="true"
            class="mb-1" 
            />
        </ul>
      </div>

    </div>
    <div v-if="!loading && !error && !account" class="text-center py-10">
        <p class="text-gray-500">Account data could not be loaded.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
// import type { Account, User } from '../types/Account'; // Keep Account, remove User
import type { Account } from '../types/Account'; // Only Account is needed directly
// import type { Asset } from '../types/Asset';     // Remove Asset import
import { fetchAccountDetails } from '../api/portalService'; // Import the API function
import AssetDisplayItem from '../components/assets/AssetDisplayItem.vue'; // Import AssetDisplayItem

// Define interfaces for the data structures
// These should ideally be imported from a shared types file if they exist
// interface Client {
//   id: string | number;
//   name: string;
//   email?: string;
//   // Add other relevant client fields
// }

// interface Asset {
//   id: string | number;
//   name: string;
//   type: string;
//   status: string;
//   // Add other relevant asset fields
// }

// interface Account {
//   legalName?: string;
//   contactEmail?: string;
//   website?: string;
//   address?: string;
//   clients?: Client[];
//   assets?: Asset[];
//   // Add other relevant account fields
// }

const account = ref<Account | null>(null);
const loading = ref(true);
const error = ref<string | null>(null);

// Placeholder function to fetch account data
// In a real application, this would make an API call
async function fetchAccountData() {
  loading.value = true;
  error.value = null;
  try {
    // Simulate API call
    // await new Promise(resolve => setTimeout(resolve, 1000));
    
    // Placeholder data - replace with actual API response
    // account.value = {
    //   legalName: 'Snowpack Data Inc.',
    //   contactEmail: 'contact@snowpack.ai',
    //   website: 'https://snowpack.ai',
    //   address: '123 Glacier Lane, Summitville, CO 80202',
    //   clients: [
    //     { id: 'client1', name: 'Example Client A', email: 'clienta@example.com' },
    //     { id: 'client2', name: 'Example Client B', email: 'clientb@example.com' },
    //   ],
    //   assets: [
    //     { id: 'asset1', name: 'Main Database Server', type: 'Server', status: 'Active' },
    //     { id: 'asset2', name: 'Data Warehouse', type: 'Database', status: 'Active' },
    //     { id: 'asset3', name: 'Reporting Tool License', type: 'Software', status: 'Active' },
    //   ],
    // };

    account.value = await fetchAccountDetails();

  } catch (e) {
    console.error('Failed to fetch account data:', e);
    error.value = 'Failed to load account information. Please try again later.';
    account.value = null; // Ensure account is null on error
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  fetchAccountData();
});
</script>

<style scoped>
/* Add any specific styles for the Settings page here */
</style> 