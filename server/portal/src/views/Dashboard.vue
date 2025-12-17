<template>
  <div class="p-4 md:p-6 lg:p-8 bg-white min-h-screen">
    <h1 class="text-3xl font-bold text-gray-800 mb-6">Client Dashboard</h1>

    <!-- Capacity Timeline Section -->
    <div class="mb-6">
      <h2 class="text-2xl font-semibold text-gray-700 mb-4">Capacity Overview</h2>
      <p class="text-sm text-gray-600 mb-4">
        View your team's commitments and actual hours worked across all projects. Bars show committed hours, with fill indicating actual hours worked.
      </p>
      <CapacityGantt />
    </div>

    <div v-if="loading" class="text-center py-10">
      <p class="text-lg text-gray-600">Loading dashboard data...</p>
      <!-- You could add a spinner here -->
    </div>

    <div v-if="error" class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-6" role="alert">
      <strong class="font-bold">Error!</strong>
      <span class="block sm:inline"> {{ error }}</span>
    </div>

    <div v-if="!loading && !error">
      <!-- Project Budgets Section -->
      <div class="space-y-4 pt-6 px-6">
         <h2 class="text-2xl font-semibold text-gray-700 mb-0">Project Budgets</h2>
        <div v-if="sortedProjectBudgets.length === 0 && !loadingBudgets" class="bg-white p-6 rounded-lg shadow-xl text-gray-500">
          No project budget information available.
        </div>
        <div v-for="budget in sortedProjectBudgets" :key="budget.project_id" class="bg-white p-6 rounded-lg shadow-xl">
          <div class="flex justify-between items-center mb-1">
            <h3 class="text-xl font-semibold text-blue-500">{{ budget.project_name }}</h3>
            <span :class="getProjectStatus(budget.project_active_end, budget.project_active_start).class" class="text-xs font-semibold px-2 py-0.5 rounded-full">
              {{ getProjectStatus(budget.project_active_end, budget.project_active_start).text }}
            </span>
          </div>
          <p class="text-xs text-gray-500 mb-1">Billing: {{ budget.billing_frequency.replace('BILLING_TYPE_', '').toLowerCase() }}</p>
          <p v-if="!budget.is_project_based_budget && budget.current_period_start_date && budget.current_period_end_date" class="text-xs text-gray-500 mb-2">
            Current Period: {{ formatDate(budget.current_period_start_date) }} - {{ formatDate(budget.current_period_end_date) }}
          </p>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mt-4">
            <!-- Overall Project Gauges -->
            <div>
              <h4 class="text-md font-medium text-gray-600 mb-1 text-center">Overall Project Budget</h4>
              <div class="grid grid-cols-2 gap-2">
                <div class="h-40">
                  <v-chart :option="getGaugeOption('Hours', budget.total_project_tracked_hours, budget.calculated_total_project_budget_hours, '%', budget.total_project_completion_hours_percent)" class="w-full h-full" />
                  <p class="text-xs text-center text-gray-500">{{ formatNumber(budget.total_project_tracked_hours, 2) }} / {{ formatNumber(budget.calculated_total_project_budget_hours, 0) }} hrs</p>
                </div>
                <div class="h-40">
                  <v-chart :option="getGaugeOption('$', budget.total_project_tracked_dollars, budget.calculated_total_project_budget_dollars, '%', budget.total_project_completion_dollars_percent)" class="w-full h-full" />
                   <p class="text-xs text-center text-gray-500">${{ formatNumber(budget.total_project_tracked_dollars, 2) }} / ${{ formatNumber(budget.calculated_total_project_budget_dollars, 0) }}</p>
                </div>
              </div>
            </div>

            <!-- Current Period Gauges -->
            <div v-if="!budget.is_project_based_budget">
              <h4 class="text-md font-medium text-gray-600 mb-1 text-center">Current Period Budget</h4>
              <div class="grid grid-cols-2 gap-2">
                <div class="h-40">
                  <v-chart :option="getGaugeOption('Hours', budget.current_period_tracked_hours, budget.current_period_budget_hours, '%', budget.current_period_completion_hours_percent)" class="w-full h-full" />
                  <p class="text-xs text-center text-gray-500">{{ formatNumber(budget.current_period_tracked_hours, 2) }} / {{ formatNumber(budget.current_period_budget_hours, 0) }} hrs</p>
                </div>
                <div class="h-40">
                  <v-chart :option="getGaugeOption('$', budget.current_period_tracked_dollars, budget.current_period_budget_dollars, '%', budget.current_period_completion_dollars_percent)" class="w-full h-full" />
                  <p class="text-xs text-center text-gray-500">${{ formatNumber(budget.current_period_tracked_dollars, 2) }} / ${{ formatNumber(budget.current_period_budget_dollars, 0) }}</p>
                </div>
              </div>
            </div>
            <div v-else class="md:col-span-1 flex items-center justify-center">
                <p class="text-sm text-gray-500 text-center">Project-based budget applies to overall only.</p>
            </div>

          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { GaugeChart, BarChart, LineChart } from 'echarts/charts';
import { TitleComponent, TooltipComponent, LegendComponent, GridComponent } from 'echarts/components';
import VChart from 'vue-echarts';
import CapacityGantt from '../components/CapacityGantt.vue';

use([
  CanvasRenderer,
  GaugeChart,
  BarChart,
  LineChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
]);

interface ProjectBudget {
  project_id: number;
  project_name: string;
  billing_frequency: string;
  project_active_start: string;
  project_active_end: string;
  current_period_start_date?: string;
  current_period_end_date?: string;
  is_project_based_budget: boolean;
  total_project_tracked_hours: number;
  calculated_total_project_budget_hours: number;
  total_project_completion_hours_percent: number;
  total_project_tracked_dollars: number;
  calculated_total_project_budget_dollars: number;
  total_project_completion_dollars_percent: number;
  current_period_tracked_hours: number;
  current_period_budget_hours: number;
  current_period_completion_hours_percent: number;
  current_period_tracked_dollars: number;
  current_period_budget_dollars: number;
  current_period_completion_dollars_percent: number;
}

const projectBudgets = ref<ProjectBudget[]>([]);
const loading = ref(true);
const loadingBudgets = ref(true);
const error = ref<string | null>(null);

const sortedProjectBudgets = computed(() => {
  return [...projectBudgets.value].sort((a, b) => {
    const dateA = new Date(a.project_active_end).getTime();
    const dateB = new Date(b.project_active_end).getTime();
    return dateB - dateA;
  });
});

const getProjectStatus = (endDateString: string, startDateString?: string): { text: string; class: string } => {
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  if (startDateString) {
    const startDate = new Date(startDateString);
    if (startDate > today) {
      return { text: 'Upcoming', class: 'bg-blue-100 text-blue-700' };
    }
  }

  if (!endDateString) {
    return { text: 'Ongoing', class: 'bg-yellow-100 text-yellow-700' };
  }
  const endDate = new Date(endDateString);

  if (endDate < today) {
    return { text: 'Completed', class: 'bg-gray-200 text-gray-700' };
  } else {
    return { text: 'Active', class: 'bg-green-200 text-green-800' };
  }
};

const fetchProjectBudgets = async () => {
  loadingBudgets.value = true;
  try {
    const token = localStorage.getItem('snowpack_token');
    if (!token) {
      throw new Error('Authentication token not found. Please log in again.');
    }
    const response = await fetch('/api/portal/project_budgets', {
      headers: { 'x-access-token': token },
    });
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ message: response.statusText }));
      throw new Error(`Failed to fetch project budgets: ${errorData.message || response.statusText} (status: ${response.status})`);
    }
    projectBudgets.value = await response.json();
  } catch (e: any) {
    console.error('Error fetching project budgets:', e);
    error.value = `${error.value ? error.value + '; ' : ''}${e.message || 'Could not load project budgets.'}`;
  }
  loadingBudgets.value = false;
};

const formatDate = (dateString?: string): string => {
  if (!dateString) return 'N/A';
  try {
    const options: Intl.DateTimeFormatOptions = { year: 'numeric', month: 'short', day: 'numeric' };
    return new Date(dateString).toLocaleDateString(undefined, options);
  } catch (e) {
    return dateString;
  }
};

const formatNumber = (num?: number, decimalPlaces = 2): string => {
  if (typeof num !== 'number' || isNaN(num)) {
    return (0).toLocaleString(undefined, { 
      minimumFractionDigits: decimalPlaces, 
      maximumFractionDigits: decimalPlaces 
    });
  }
  return num.toLocaleString(undefined, {
    minimumFractionDigits: decimalPlaces,
    maximumFractionDigits: decimalPlaces
  });
};

const getGaugeOption = (title: string, value?: number, total?: number, unit?: string, percentage?: number) => {
  const val = typeof value === 'number' && !isNaN(value) ? value : 0;
  const tot = typeof total === 'number' && !isNaN(total) && total > 0 ? total : 0;
  
  let displayPercentage = (typeof percentage === 'number' && !isNaN(percentage)) ? percentage : 0;
  if (tot > 0) {
      displayPercentage = (val / tot) * 100;
  }
  displayPercentage = Math.max(0, Math.min(100, displayPercentage));

  return {
    series: [
      {
        type: 'gauge',
        startAngle: 90,
        endAngle: -270,
        pointer: { show: false },
        progress: {
          show: true,
          overlap: false,
          roundCap: true,
          clip: false,
          itemStyle: {
            borderWidth: 1,
            borderColor: '#3B82F6',
            color: '#A5B4FC'
          }
        },
        axisLine: {
          lineStyle: {
            width: 12,
            color: [[1, '#EAEAEA']]
          }
        },
        splitLine: { show: false },
        axisTick: { show: false },
        axisLabel: { show: false },
        title: {
          offsetCenter: [0, '60%'],
          fontSize: 11,
          color: '#666'
        },
        detail: {
          width: '60%',
          lineHeight: 20,
          height: 20,
          fontSize: 16,
          color: '#333',
          backgroundColor: '#fff',
          borderRadius: 3,
          offsetCenter: [0, '0%'],
          valueAnimation: true,
          formatter: function () {
            return formatNumber(displayPercentage, 0) + (unit || '%');
          }
        },
        data: [
          {
            value: displayPercentage,
            name: title,
          },
        ],
      },
    ],
  };
};

onMounted(async () => {
  loading.value = true;
  error.value = null;
  await fetchProjectBudgets();
  if(loadingBudgets.value) {
      loadingBudgets.value = false;
  }
  loading.value = false;
});

</script>

<style scoped>
/* Scoped styles if needed, Tailwind handles most styling */
.echarts {
  /* width and height are now set by w-full h-full on the v-chart and explicit h-40 on parent */
}

/* Custom scrollbar for draft entries (optional) */
.max-h-\[70vh\]::-webkit-scrollbar {
  width: 6px;
}
.max-h-\[70vh\]::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 10px;
}
.max-h-\[70vh\]::-webkit-scrollbar-thumb {
  background: #c7c7c7;
  border-radius: 10px;
}
.max-h-\[70vh\]::-webkit-scrollbar-thumb:hover {
  background: #a3a3a3;
}
</style> 