<template>
  <div
    :class="[
      'shrink-0 rounded-full flex items-center justify-center bg-sage-pale',
      sizeClasses[size]
    ]"
  >
    <img
      v-if="imageUrl"
      :src="imageUrl"
      :alt="`${employee.first_name} ${employee.last_name}`"
      :class="[
        'rounded-full object-cover',
        sizeClasses[size]
      ]"
    />
    <span
      v-else
      :class="[
        'text-sage font-medium',
        textSizeClasses[size]
      ]"
    >
      {{ initials }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

interface Employee {
  first_name?: string;
  last_name?: string;
  headshot_asset?: {
    url?: string;
  };
}

interface Props {
  employee: Employee;
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md'
});

const sizeClasses = {
  xs: 'h-6 w-6',
  sm: 'h-8 w-8',
  md: 'h-10 w-10',
  lg: 'h-12 w-12',
  xl: 'h-16 w-16'
};

const textSizeClasses = {
  xs: 'text-2xs',
  sm: 'text-2xs',
  md: 'text-sm',
  lg: 'text-base',
  xl: 'text-xl'
};

const imageUrl = computed(() => {
  return props.employee.headshot_asset?.url || null;
});

const initials = computed(() => {
  const first = props.employee.first_name?.[0] || '';
  const last = props.employee.last_name?.[0] || '';
  return (first + last).toUpperCase();
});
</script>

