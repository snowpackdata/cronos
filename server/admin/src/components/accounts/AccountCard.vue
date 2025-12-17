<script setup lang="ts">
import { ref } from 'vue';
import type { Account } from '../../types/Account';
import AssetDisplayItem from '../assets/AssetDisplayItem.vue';

// Define props and emits
defineProps<{
  account: Account
}>();

const emit = defineEmits<{
  (e: 'edit', account: Account): void,
  (e: 'invite-client', account: Account): void,
  (e: 'add-asset', accountId: number): void,
  (e: 'asset-deleted'): void
}>();

// Expandable state
const isExpanded = ref(false);

// Toggle expanded state
const toggleExpanded = () => {
  isExpanded.value = !isExpanded.value;
};

// Format billing frequency for display
const formatBillingFrequency = (frequency: string) => {
  const frequencies: Record<string, string> = {
    'BILLING_TYPE_WEEKLY': 'Weekly',
    'BILLING_TYPE_BIWEEKLY': 'Bi-Weekly',
    'BILLING_TYPE_MONTHLY': 'Monthly',
    'BILLING_TYPE_BIMONTHLY': 'Bi-Monthly',
    'BILLING_TYPE_PROJECT': 'Project-Based'
  };
  return frequencies[frequency] || frequency;
};

// Format budget frequency suffix based on billing frequency
const formatBudgetFrequency = (frequency: string) => {
  switch (frequency) {
    case 'BILLING_TYPE_WEEKLY':
      return '/wk';
    case 'BILLING_TYPE_BIWEEKLY':
      return '/2wk';
    case 'BILLING_TYPE_MONTHLY':
      return '/mo';
    case 'BILLING_TYPE_BIMONTHLY':
      return '/2mo';
    case 'BILLING_TYPE_PROJECT':
      return ' total';
    default:
      return '';
  }
};

// Format currency
const formatCurrency = (amount: number) => {
  if (!amount) return '$0';
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    maximumFractionDigits: 0
  }).format(amount);
};

// Handler for when an asset is deleted
const handleAssetDeleted = () => {
  emit('asset-deleted');
};
</script>

<template>
  <div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow transition hover:shadow-md">
    <!-- Account Header - Clickable to expand/collapse -->
    <div 
      class="flex items-start justify-between p-4 cursor-pointer hover:bg-gray-50 transition-colors"
      @click="toggleExpanded"
    >
      <div class="flex-1">
        <div class="flex items-center gap-x-2">
          <!-- Expand/Collapse Chevron -->
          <i 
            class="fas transition-transform text-gray-400 text-xs"
            :class="isExpanded ? 'fa-chevron-down' : 'fa-chevron-right'"
          ></i>
          
          <h3 class="text-base font-semibold text-gray-900">{{ account.name }}</h3>
          
          <!-- Invoice Type Badge -->
          <span 
            class="inline-flex items-center rounded-md px-1.5 py-0.5 text-xs font-medium ring-1 ring-inset"
            :class="account.projects_single_invoice ? 
              'bg-green-50 text-green-700 ring-green-600/20' : 
              'bg-amber-50 text-amber-700 ring-amber-600/20'">
            {{ account.projects_single_invoice ? 'Combined' : 'Separate' }}
          </span>
        </div>
        <p class="mt-0.5 text-xs text-gray-500 ml-5">{{ account.legal_name || 'No legal name provided' }}</p>
        <!-- Display Client Users Count when collapsed -->
        <p v-if="!isExpanded && account.client_users && account.client_users.length > 0" class="mt-1 text-xs text-gray-500 ml-5">
          {{ account.client_users.length }} client user(s)
        </p>
      </div>
      <div class="flex items-center space-x-2" @click.stop>
        <button
          @click="$emit('edit', account)"
          class="inline-flex items-center rounded-md bg-blue-50 px-2 py-1 text-xs font-medium text-blue-700 ring-1 ring-inset ring-blue-700/10 hover:bg-blue-100 transition-colors"
          title="Edit Account"
        >
          <i class="fas fa-pencil-alt mr-1"></i>
          Edit
        </button>
      </div>
    </div>
    
    <!-- Expandable Card Body -->
    <div v-if="isExpanded" class="border-t border-gray-100 p-4 pt-3">
      <!-- Account Details Section -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-x-6 gap-y-4">
        <!-- Left Column: Billing Details, Email, Address -->
        <div class="space-y-3">
          <div class="flex items-baseline">
            <dt class="text-xs font-medium text-gray-500 mr-2">Billing Frequency:</dt>
            <dd class="text-xs text-gray-900">{{ formatBillingFrequency(account.billing_frequency) }}</dd>
          </div>
          <div class="flex items-baseline">
            <dt class="text-xs font-medium text-gray-500 mr-2">Invoice Type:</dt>
            <dd class="text-xs text-gray-900">
              <span class="inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset"
                :class="account.projects_single_invoice ? 
                  'bg-green-50 text-green-700 ring-green-600/20' : 
                  'bg-amber-50 text-amber-700 ring-amber-600/20'">
                {{ account.projects_single_invoice ? 'Combined' : 'Separate' }}
              </span>
            </dd>
          </div>
          <div class="flex items-baseline">
            <dt class="text-xs font-medium text-gray-500 mr-2">Email:</dt>
            <dd class="text-xs text-gray-900">{{ account.email || 'Not provided' }}</dd>
          </div>
          <div v-if="account.address" class="pt-2 border-t border-gray-100">
            <dt class="text-xs font-medium text-gray-500">Address:</dt>
            <dd class="mt-1 text-xs text-gray-900 whitespace-pre-line">{{ account.address }}</dd>
          </div>
        </div>

        <!-- Right Column: Budget Information -->
        <div>
          <div v-if="account.budget_dollars || account.budget_hours">
            <div class="mb-2 text-xs font-medium text-gray-500">Budget:</div>
            <div class="flex flex-col gap-y-1">
              <div v-if="account.budget_dollars" class="flex items-center">
                <span class="inline-flex items-center rounded-md bg-sage-50 px-2 py-1 text-xs font-medium text-sage-700 ring-1 ring-inset ring-sage-600/20">
                  <i class="fas fa-dollar-sign mr-1.5"></i>
                  {{ formatCurrency(account.budget_dollars) }}{{ formatBudgetFrequency(account.billing_frequency) }}
                </span>
              </div>
              <div v-if="account.budget_hours" class="flex items-center">
                <span class="inline-flex items-center rounded-md bg-indigo-50 px-2 py-1 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-600/20">
                  <i class="fas fa-clock mr-1.5"></i>
                  {{ account.budget_hours }} hours{{ formatBudgetFrequency(account.billing_frequency) }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Client Users and Assets Section -->
      <div class="mt-4 grid grid-cols-1 gap-x-6 gap-y-4 border-t border-gray-100 pt-4 lg:grid-cols-2">
        <!-- Client Users Column -->
        <div>
          <div class="flex justify-between items-center mb-2">
            <h5 class="text-xs font-medium text-gray-900">Client Users:</h5>
            <button
              type="button"
              @click.stop="$emit('invite-client', account)"
              class="rounded bg-indigo-50 px-2 py-1 text-xs font-semibold text-indigo-600 shadow-sm hover:bg-indigo-100"
            >
              Invite Client
            </button>
          </div>
          <div v-if="account.client_users && account.client_users.length > 0" 
               class="space-y-0 divide-y divide-gray-100 border border-gray-200 rounded-lg overflow-hidden">
            <div v-for="clientUser in account.client_users" :key="clientUser.user_id" 
                 class="flex items-center justify-between gap-x-4 py-2.5 px-3 hover:bg-gray-50">
              <div class="min-w-0 flex-1">
                <p class="text-xs font-medium text-gray-800">
                  <template v-if="clientUser.status === 'Pending Registration'">
                    {{ clientUser.email }}
                  </template>
                  <template v-else>
                    {{ clientUser.first_name }} {{ clientUser.last_name }}
                  </template>
                </p>
                <p v-if="clientUser.status !== 'Pending Registration' && clientUser.email" class="text-xs text-gray-500 mt-0.5">
                  {{ clientUser.email }}
                </p>
              </div>
              <div class="flex-shrink-0">
                <span v-if="clientUser.status === 'Pending Registration'"
                      class="inline-flex items-center rounded-md bg-amber-200 px-2.5 py-1 text-xs font-semibold text-amber-800 ring-1 ring-inset ring-amber-400 shadow-sm">
                  Pending
                </span>
                <span v-else-if="clientUser.status === 'Active'"
                      class="inline-flex items-center rounded-md bg-green-200 px-2.5 py-1 text-xs font-semibold text-green-800 ring-1 ring-inset ring-green-400 shadow-sm">
                  Active
                </span>
              </div>
            </div>
          </div>
          <p v-else class="text-xs text-gray-500 italic bg-gray-50 p-3 rounded-lg border border-gray-200 text-center">No client users.</p>
        </div>

        <!-- Assets Column -->
        <div>
          <div class="flex justify-between items-center mb-2">
            <h5 class="text-xs font-medium text-gray-900">Assets:</h5>
            <button 
              type="button"
              @click.stop="$emit('add-asset', account.ID)" 
              class="rounded bg-sage-50 px-2 py-1 text-xs font-semibold text-sage-700 shadow-sm hover:bg-sage-100"
              title="Add File/Asset to this Account"
            >
              <i class="fas fa-plus mr-1"></i> Add Asset
            </button>
          </div>
          <div v-if="account.assets && account.assets.length > 0" class="space-y-2">
            <AssetDisplayItem 
              v-for="asset in account.assets" 
              :key="asset.ID" 
              :asset="asset" 
              :account-id="account.ID"
              @delete-asset="handleAssetDeleted"
            />
          </div>
          <p v-else class="text-xs text-gray-500 italic bg-gray-50 p-3 rounded-lg border border-gray-200 text-center">No assets.</p>
        </div>
      </div>
    </div>
    
    <!-- View Details Button -->
    <div v-if="isExpanded" class="mt-4 pt-2 border-t border-gray-100">
      <button 
        @click="toggleExpanded"
        class="text-xs flex items-center justify-center w-full px-2 py-1.5 font-medium rounded-md bg-gray-50 text-gray-700 hover:bg-gray-100 transition-colors"
      >
        <span>{{ isExpanded ? 'Hide details' : 'View details' }}</span>
        <i :class="[isExpanded ? 'fa-chevron-up' : 'fa-chevron-down', 'fas ml-2']"></i>
      </button>
    </div>
  </div>
</template>

<style scoped>
.bg-sage-50 {
  background-color: #F0F4F0;
}
.text-sage-700 {
  color: #2E6E32;
}
</style> 