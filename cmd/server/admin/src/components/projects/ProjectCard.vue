<script setup lang="ts">
import type { Project } from '../../types/Project';

// Define props and emits
defineProps<{
  project: Project
}>();

defineEmits<{
  (e: 'edit', project: Project): void
}>();

// Format billing frequency for display
const formatBillingFrequency = (frequency: string) => {
  const frequencies: Record<string, string> = {
    'BILLING_TYPE_WEEKLY': 'Weekly',
    'BILLING_TYPE_BIWEEKLY': 'Bi-Weekly',
    'BILLING_TYPE_MONTHLY': 'Monthly',
    'BILLING_TYPE_BIMONTHLY': 'Bi-Monthly',
    'BILLING_TYPE_PROJECT': 'Project'
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

// Get project status
const getProjectStatus = (project: Project) => {
  const now = new Date();
  const startDate = project.active_start ? new Date(project.active_start) : null;
  const endDate = project.active_end ? new Date(project.active_end) : null;

  if (startDate && startDate > now) {
    return {
      label: 'Upcoming',
      color: 'bg-blue-50 text-blue-700 ring-blue-600/20'
    };
  } else if (endDate && endDate < now) {
    return {
      label: 'Ended',
      color: 'bg-gray-100 text-gray-700 ring-gray-600/20'
    };
  } else {
    return {
      label: 'Active',
      color: 'bg-green-50 text-green-700 ring-green-600/20'
    };
  }
};
</script>

<template>
  <div class="overflow-hidden rounded-lg border border-gray-200 bg-white hover:border-sage hover:bg-gray-50 transition-all">
    <!-- Project Content - Compact Single View -->
    <div class="px-3 py-2">
      <div class="flex items-start justify-between gap-2">
        <!-- Left: Project Info -->
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2">
            <h3 class="text-sm font-semibold text-gray-900 truncate">{{ project.name }}</h3>
            <span 
              class="inline-flex items-center rounded px-1.5 py-0.5 text-xs font-medium ring-1 ring-inset flex-shrink-0"
              :class="getProjectStatus(project).color">
              {{ getProjectStatus(project).label }}
            </span>
            <span 
              v-if="project.internal"
              class="inline-flex items-center rounded bg-purple-50 px-1.5 py-0.5 text-xs font-medium text-purple-700 ring-1 ring-inset ring-purple-700/10 flex-shrink-0">
              Internal
            </span>
          </div>
          
          <p class="text-xs text-gray-500 truncate">{{ project.account ? project.account.name : 'No Account' }}</p>
          
          <!-- Compact Stats Row -->
          <div class="flex items-center gap-3 text-xs text-gray-600 mt-1">
            <span class="flex items-center gap-1">
              <i class="fas fa-calendar-alt text-gray-400"></i>
              {{ formatBillingFrequency(project.billing_frequency) }}
            </span>
            <span v-if="project.budget_dollars" class="flex items-center gap-1">
              <i class="fas fa-dollar-sign text-gray-400"></i>
              {{ formatCurrency(project.budget_dollars) }}{{ formatBudgetFrequency(project.billing_frequency) }}
            </span>
            <span v-if="project.staffing_assignments && project.staffing_assignments.length > 0" class="flex items-center gap-1">
              <i class="fas fa-users text-gray-400"></i>
              {{ project.staffing_assignments.length }}
            </span>
            <span v-if="project.billing_codes && project.billing_codes.length > 0" class="flex items-center gap-1">
              <i class="fas fa-barcode text-gray-400"></i>
              {{ project.billing_codes.length }}
            </span>
          </div>
        </div>

        <!-- Right: Actions -->
        <div class="flex items-center gap-1 flex-shrink-0" @click.stop>
          <button
            @click="$emit('edit', project)"
            class="inline-flex items-center rounded-md bg-blue-50 px-2 py-1 text-xs font-medium text-blue-700 ring-1 ring-inset ring-blue-700/10 hover:bg-blue-100 transition-colors"
            title="Edit Project"
          >
            <i class="fas fa-pencil-alt"></i>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
