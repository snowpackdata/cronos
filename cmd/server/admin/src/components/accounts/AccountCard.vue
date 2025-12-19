<script setup lang="ts">
import type { Account } from '../../types/Account';

// Define props and emits
defineProps<{
  account: Account
}>();

defineEmits<{
  (e: 'edit', account: Account): void,
  (e: 'invite-client', account: Account): void,
  (e: 'add-asset', accountId: number): void,
  (e: 'asset-deleted'): void
}>();

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
</script>

<template>
  <div class="overflow-hidden rounded-lg border border-gray-200 bg-white hover:border-sage hover:bg-gray-50 transition-all">
    <!-- Account Content - Compact Single View -->
    <div class="px-3 py-2">
      <div class="flex items-start justify-between gap-2">
        <!-- Left: Account Info -->
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2">
            <h3 class="text-sm font-semibold text-gray-900 truncate">{{ account.name }}</h3>
            <span 
              class="inline-flex items-center rounded px-1.5 py-0.5 text-xs font-medium ring-1 ring-inset flex-shrink-0"
              :class="account.projects_single_invoice ? 
                'bg-green-50 text-green-700 ring-green-600/20' : 
                'bg-amber-50 text-amber-700 ring-amber-600/20'">
              {{ account.projects_single_invoice ? 'Combined' : 'Separate' }}
            </span>
          </div>
          
          <p v-if="account.legal_name" class="text-xs text-gray-500 truncate">{{ account.legal_name }}</p>
          
          <!-- Compact Stats Row -->
          <div class="flex items-center gap-3 text-xs text-gray-600 mt-1">
            <span class="flex items-center gap-1">
              <i class="fas fa-calendar-alt text-gray-400"></i>
              {{ formatBillingFrequency(account.billing_frequency) }}
            </span>
            <span v-if="account.budget_dollars" class="flex items-center gap-1">
              <i class="fas fa-dollar-sign text-gray-400"></i>
              {{ formatCurrency(account.budget_dollars) }}{{ formatBudgetFrequency(account.billing_frequency) }}
            </span>
            <span v-if="account.email" class="flex items-center gap-1 truncate">
              <i class="fas fa-envelope text-gray-400"></i>
              <span class="truncate">{{ account.email }}</span>
            </span>
          </div>
        </div>

        <!-- Right: Actions -->
        <div class="flex items-center gap-1 flex-shrink-0" @click.stop>
          <button
            @click="$emit('edit', account)"
            class="inline-flex items-center rounded-md bg-blue-50 px-2 py-1 text-xs font-medium text-blue-700 ring-1 ring-inset ring-blue-700/10 hover:bg-blue-100 transition-colors"
            title="Edit Account"
          >
            <i class="fas fa-pencil-alt"></i>
          </button>
          <button
            @click="$emit('invite-client', account)"
            class="inline-flex items-center rounded-md bg-green-50 px-2 py-1 text-xs font-medium text-green-700 ring-1 ring-inset ring-green-700/10 hover:bg-green-100 transition-colors"
            title="Invite Client"
          >
            <i class="fas fa-user-plus"></i>
          </button>
          <button
            @click="$emit('add-asset', account.ID)"
            class="inline-flex items-center rounded-md bg-purple-50 px-2 py-1 text-xs font-medium text-purple-700 ring-1 ring-inset ring-purple-700/10 hover:bg-purple-100 transition-colors"
            title="Add Asset"
          >
            <i class="fas fa-paperclip"></i>
          </button>
        </div>
      </div>
    </div>
  </div>
</template> 