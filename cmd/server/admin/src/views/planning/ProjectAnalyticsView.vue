<template>
  <div class="flex h-screen overflow-hidden bg-gray-50">
    <!-- Sidebar - Project List -->
    <div class="w-64 bg-white border-r border-gray-200 flex flex-col">
      <div class="p-3 border-b border-gray-200">
        <h1 class="text-base font-semibold text-gray-900">Projects</h1>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search..."
          class="mt-2 block w-full rounded-md border-gray-300 shadow-sm focus:border-sage focus:ring-sage text-xs"
        />
      </div>
      
      <div class="flex-1 overflow-y-auto">
        <div v-if="projectsLoading" class="p-3 text-center text-xs text-gray-500">
          Loading...
        </div>
        <div v-else-if="projectsError" class="p-3 text-center text-xs text-red-600">
          {{ projectsError }}
        </div>
        <div v-else>
          <button
            v-for="project in sortedFilteredProjects"
            :key="project.ID"
            @click="selectProject(project.ID)"
            :class="[
              'w-full text-left px-3 py-2 hover:bg-gray-50 transition-colors border-l-2',
              selectedProjectId === project.ID ? 'bg-sage bg-opacity-5 border-sage' : 'border-transparent',
              isProjectEnded(project) ? 'opacity-60' : ''
            ]"
          >
            <div class="flex items-start justify-between">
              <div class="flex-1 min-w-0">
                <p class="text-xs font-medium text-gray-900 truncate">
                  {{ project.name }}
                  <span v-if="isProjectEnded(project)" class="ml-1 text-xs text-gray-400">(Complete)</span>
                </p>
                <p class="text-xs text-gray-700 truncate mt-0.5 font-medium">{{ project.account?.name }}</p>
              </div>
              <div class="ml-2 flex-shrink-0">
                <span
                  v-if="!isProjectEnded(project)"
                  :class="[
                    'inline-block w-2 h-2 rounded-full',
                    getProjectStatusColor(project)
                  ]"
                  :title="getProjectStatusTitle(project)"
                ></span>
                <i v-else class="fas fa-check-circle text-xs text-gray-400" title="Complete"></i>
              </div>
            </div>
          </button>
        </div>
        <div v-if="!projectsLoading && sortedFilteredProjects.length === 0" class="p-3 text-center text-xs text-gray-500">
          No projects found
        </div>
      </div>
    </div>

    <!-- Main Content Area -->
    <div class="flex-1 overflow-y-auto">
      <div v-if="!selectedProjectId" class="flex items-center justify-center h-full">
        <div class="text-center text-gray-500">
          <i class="fas fa-chart-line text-5xl mb-3 text-gray-400"></i>
          <p class="text-lg">Select a project to view analytics</p>
        </div>
      </div>

      <!-- Loading State -->
      <div v-else-if="dataLoading" class="flex items-center justify-center h-full">
        <div class="flex items-center">
          <i class="fas fa-spinner fa-spin text-sage text-2xl mr-3"></i>
          <span class="text-gray-700">Loading project data...</span>
        </div>
      </div>

      <!-- Error State -->
      <div v-else-if="dataError" class="flex items-center justify-center h-full p-8">
        <div class="rounded-md bg-red-50 p-4 max-w-md">
          <div class="flex">
            <div class="flex-shrink-0">
              <i class="fas fa-exclamation-circle text-red-400"></i>
            </div>
            <div class="ml-3">
              <h3 class="text-sm font-medium text-red-800">Error loading data</h3>
              <div class="mt-2 text-sm text-red-700">
                <p>{{ dataError }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Main Analytics Content -->
      <div v-else-if="projectData" class="p-6">
        <div class="mb-4">
          <h2 class="text-lg font-semibold text-gray-900">{{ projectData.project_name }}</h2>
        </div>
        <!-- Compact Status Cards -->
        <div class="grid grid-cols-5 gap-4 mb-6">
          <div class="bg-white border border-gray-200 rounded px-3 py-2">
            <dt class="text-xs font-medium text-gray-500 uppercase">Status</dt>
            <dd class="mt-1">
              <span
                :class="[
                  'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium',
                  getBudgetStatusClass()
                ]"
              >
                {{ getBudgetStatusLabel() }}
              </span>
            </dd>
          </div>

          <div class="bg-white border border-gray-200 rounded px-3 py-2">
            <dt class="text-xs font-medium text-gray-500 uppercase">Hours</dt>
            <dd class="mt-1 text-lg font-semibold text-gray-900">
              {{ projectData.hours_completion.toFixed(1) }}%
            </dd>
            <dd class="text-xs text-gray-500">
              {{ projectData.total_tracked_hours.toFixed(0) }} / {{ projectData.total_budget_hours.toFixed(0) }}h
            </dd>
          </div>

          <div class="bg-white border border-gray-200 rounded px-3 py-2">
            <dt class="text-xs font-medium text-gray-500 uppercase">Hours Remaining</dt>
            <dd class="mt-1 text-lg font-semibold text-gray-900">
              {{ (projectData.total_budget_hours - projectData.total_tracked_hours).toFixed(0) }}h
            </dd>
            <dd class="text-xs text-gray-500">
              {{ (100 - projectData.hours_completion).toFixed(1) }}% left
            </dd>
          </div>

          <div class="bg-white border border-gray-200 rounded px-3 py-2">
            <dt class="text-xs font-medium text-gray-500 uppercase">Budget</dt>
            <dd class="mt-1 text-lg font-semibold text-gray-900">
              {{ projectData.dollars_completion.toFixed(1) }}%
            </dd>
            <dd class="text-xs text-gray-500">
              ${{ formatCurrency(projectData.total_revenue) }} / ${{ formatCurrency(projectData.total_budget_dollars) }}
            </dd>
          </div>

          <div class="bg-white border border-gray-200 rounded px-3 py-2">
            <dt class="text-xs font-medium text-gray-500 uppercase">Budget Remaining</dt>
            <dd class="mt-1 text-lg font-semibold text-gray-900">
              ${{ formatCurrency(projectData.total_budget_dollars - projectData.total_revenue) }}
            </dd>
            <dd class="text-xs text-gray-500">
              {{ (100 - projectData.dollars_completion).toFixed(1) }}% left
            </dd>
          </div>
        </div>

        <!-- Budget Burn-up Chart -->
        <div class="bg-white border border-gray-200 rounded-lg p-4 mb-6">
          <div class="mb-3">
            <h2 class="text-sm font-semibold text-gray-900">Budget Burn-up</h2>
          </div>
          <div v-if="projectData.burndown_data && projectData.burndown_data.length > 0">
            <div style="position: relative; height: 300px;">
              <canvas ref="burndownChart"></canvas>
            </div>
          </div>
          <div v-else class="text-center text-gray-500 py-12 text-sm">
            <p>No burndown data available</p>
          </div>
        </div>

        <!-- Profitability Metrics -->
        <div class="bg-white border border-gray-200 rounded-lg p-4">
          <h2 class="text-sm font-semibold text-gray-900 mb-3">Profitability</h2>
          <div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
          <div>
            <dt class="text-xs font-medium text-gray-500 uppercase tracking-wide">Total Revenue</dt>
            <dd class="mt-1 text-2xl font-semibold text-gray-900">
              ${{ formatCurrency(projectData.total_revenue) }}
            </dd>
            <dd class="text-xs text-gray-500 mt-0.5">Accrued receivables</dd>
          </div>
          <div>
            <dt class="text-xs font-medium text-gray-500 uppercase tracking-wide">Total Cost</dt>
            <dd class="mt-1 text-2xl font-semibold text-gray-900">
              ${{ formatCurrency(projectData.total_cost) }}
            </dd>
            <dd class="text-xs text-gray-500 mt-0.5">Accrued payable</dd>
          </div>
          <div>
            <dt class="text-xs font-medium text-gray-500 uppercase tracking-wide">Net Profit</dt>
            <dd
              :class="[
                'mt-1 text-2xl font-semibold',
                projectData.total_profit >= 0 ? 'text-green-600' : 'text-red-600'
              ]"
            >
              ${{ formatCurrency(projectData.total_profit) }}
            </dd>
            <dd class="text-xs text-gray-500 mt-0.5">Revenue - Cost</dd>
          </div>
          <div>
            <dt class="text-xs font-medium text-gray-500 uppercase tracking-wide">Gross Margin</dt>
            <dd
              :class="[
                'mt-1 text-2xl font-semibold',
                projectData.profit_margin >= 0 ? 'text-green-600' : 'text-red-600'
              ]"
            >
              {{ projectData.profit_margin.toFixed(1) }}%
            </dd>
            <dd class="text-xs text-gray-500 mt-0.5">Profit / Revenue</dd>
          </div>
        </div>
        <div class="mt-6 pt-6 border-t border-gray-200 grid grid-cols-1 gap-6 sm:grid-cols-2">
          <div>
            <dt class="text-xs font-medium text-gray-500 uppercase tracking-wide">Total Invoiced</dt>
            <dd class="mt-1 text-xl font-semibold text-gray-900">
              ${{ formatCurrency(projectData.total_invoiced) }}
            </dd>
            <dd class="text-xs text-gray-500 mt-0.5">All invoices</dd>
          </div>
          <div>
            <dt class="text-xs font-medium text-gray-500 uppercase tracking-wide">Total COGS</dt>
            <dd class="mt-1 text-xl font-semibold text-gray-900">
              ${{ formatCurrency(projectData.total_cost) }}
            </dd>
            <dd class="text-xs text-gray-500 mt-0.5">Cost of goods sold</dd>
          </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { fetchProjects } from '../../api/projects'
import { api } from '../../api/apiUtils'
import { 
  Chart, 
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  LineController,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import annotationPlugin from 'chartjs-plugin-annotation'

// Register Chart.js components
Chart.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  LineController,
  Title,
  Tooltip,
  Legend,
  Filler,
  annotationPlugin
)


console.log('Chart.js registered successfully')

const route = useRoute()
const router = useRouter()

interface Project {
  ID: number
  name: string
  billing_frequency?: string
  active_start?: string
  active_end?: string
  UpdatedAt?: string
  account?: {
    name: string
  }
}

interface BurndownDataPoint {
  date: string
  planned_budget_remaining: number
  actual_spent: number
  actual_cost: number
  invoiced_amount: number
}

interface ProjectProfitabilityData {
  project_id: number
  project_name: string
  billing_frequency: string
  project_active_start: string
  project_active_end: string
  budget_hours_per_period: number
  budget_dollars_per_period: number
  budget_cap_hours: number
  budget_cap_dollars: number
  total_budget_hours: number
  total_budget_dollars: number
  total_tracked_hours: number
  total_revenue: number
  total_cost: number
  total_profit: number
  profit_margin: number
  total_invoiced: number
  total_invoiced_accepted: number
  hours_completion: number
  dollars_completion: number
  is_ahead_of_budget: boolean
  is_on_budget: boolean
  is_behind_budget: boolean
  burndown_data: BurndownDataPoint[]
}

const projects = ref<Project[]>([])
const selectedProjectId = ref<number | null>(null)
const projectData = ref<ProjectProfitabilityData | null>(null)
const projectsLoading = ref(false)
const projectsError = ref<string | null>(null)
const dataLoading = ref(false)
const dataError = ref<string | null>(null)
const burndownChart = ref<HTMLCanvasElement | null>(null)
const searchQuery = ref('')
let chartInstance: InstanceType<typeof Chart> | null = null

// Check if a project has ended
function isProjectEnded(project: Project): boolean {
  if (!project.active_end) return false
  const now = new Date()
  const end = new Date(project.active_end)
  return now > end
}

// Check if a project is currently active based on dates
function isProjectActive(project: Project): boolean {
  if (!project.active_start || !project.active_end) return false
  const now = new Date()
  const start = new Date(project.active_start)
  const end = new Date(project.active_end)
  return now >= start && now <= end
}

// Calculate project status color for sidebar (placeholder - we don't have budget data here)
function getProjectStatusColor(project: Project): string {
  // For now, just show green if active, gray if not started
  if (isProjectActive(project)) {
    return 'bg-green-400'
  }
  const now = new Date()
  const start = project.active_start ? new Date(project.active_start) : null
  if (start && now < start) {
    return 'bg-blue-400' // Not started yet
  }
  return 'bg-gray-300'
}

// Get status title for sidebar
function getProjectStatusTitle(project: Project): string {
  if (isProjectActive(project)) return 'Active'
  const now = new Date()
  const start = project.active_start ? new Date(project.active_start) : null
  if (start && now < start) return 'Not Started'
  return 'Inactive'
}

// Calculate budget status based on time vs spend
// On-track: 15% behind up to 5% ahead (relative to time)
function getBudgetStatusLabel(): string {
  if (!projectData.value) return 'Unknown'
  
  const timeElapsed = getTimeElapsedPercent()
  const budgetUsed = projectData.value.dollars_completion
  const variance = budgetUsed - timeElapsed
  
  // variance > 0 means we're spending more than expected
  // variance < 0 means we're spending less than expected
  
  if (variance >= -5 && variance <= 15) {
    return 'On Track'
  } else if (variance < -5) {
    return 'Under Budget'
  } else {
    return 'Over Budget'
  }
}

// Get budget status CSS class
function getBudgetStatusClass(): string {
  if (!projectData.value) return 'bg-gray-100 text-gray-800'
  
  const timeElapsed = getTimeElapsedPercent()
  const budgetUsed = projectData.value.dollars_completion
  const variance = budgetUsed - timeElapsed
  
  if (variance >= -5 && variance <= 15) {
    return 'bg-blue-100 text-blue-800'
  } else if (variance < -5) {
    return 'bg-yellow-100 text-yellow-800' // Under-budget is yellow (caution)
  } else {
    return 'bg-red-100 text-red-800'
  }
}

// Calculate percentage of time elapsed in project
function getTimeElapsedPercent(): number {
  if (!projectData.value) return 0
  
  const start = new Date(projectData.value.project_active_start)
  const end = new Date(projectData.value.project_active_end)
  const now = new Date()
  
  if (now < start) return 0
  if (now > end) return 100
  
  const totalDuration = end.getTime() - start.getTime()
  const elapsed = now.getTime() - start.getTime()
  
  return (elapsed / totalDuration) * 100
}

// Filtered projects based on search
const filteredProjects = computed(() => {
  if (!searchQuery.value) return projects.value
  const query = searchQuery.value.toLowerCase()
  return projects.value.filter(p => 
    p.name.toLowerCase().includes(query) ||
    (p.account?.name || '').toLowerCase().includes(query)
  )
})

// Sort filtered projects: active first (by recency), then ended projects
// Format currency with thousand separators and no decimals
function formatCurrency(value: number): string {
  return Math.round(value).toLocaleString('en-US')
}

const sortedFilteredProjects = computed(() => {
  const active: Project[] = []
  const ended: Project[] = []
  
  filteredProjects.value.forEach(project => {
    if (isProjectEnded(project)) {
      ended.push(project)
    } else {
      active.push(project)
    }
  })
  
  // Sort active by UpdatedAt (most recent first)
  active.sort((a, b) => {
    const dateA = a.UpdatedAt ? new Date(a.UpdatedAt).getTime() : 0
    const dateB = b.UpdatedAt ? new Date(b.UpdatedAt).getTime() : 0
    return dateB - dateA
  })
  
  // Sort ended by end date (most recent first)
  ended.sort((a, b) => {
    const dateA = a.active_end ? new Date(a.active_end).getTime() : 0
    const dateB = b.active_end ? new Date(b.active_end).getTime() : 0
    return dateB - dateA
  })
  
  return [...active, ...ended]
})

// Load all projects on mount
onMounted(async () => {
  projectsLoading.value = true
  try {
    projects.value = await fetchProjects()
    
    // Check if there's a projectId in the route
    const projectIdParam = route.params.projectId
    if (projectIdParam) {
      const projectId = parseInt(projectIdParam as string, 10)
      if (!isNaN(projectId)) {
        selectProject(projectId)
      }
    }
  } catch (err: any) {
    projectsError.value = err.message || 'Failed to load projects'
  } finally {
    projectsLoading.value = false
  }
})

// Select a project
function selectProject(projectId: number) {
  selectedProjectId.value = projectId
  // Update URL for deep linking
  router.push({ name: 'project-analytics', params: { projectId: projectId.toString() } })
  loadProjectData()
}

// Load project profitability data
async function loadProjectData() {
  if (!selectedProjectId.value) {
    projectData.value = null
    return
  }

  dataLoading.value = true
  dataError.value = null

  try {
    const response = await api.get<ProjectProfitabilityData>(
      `/api/project-profitability?project_id=${selectedProjectId.value}`
    )

    projectData.value = response.data
    // Chart will render via watcher when canvas ref becomes available
  } catch (err: any) {
    dataError.value = err.response?.data?.message || err.message || 'Failed to load project data'
  } finally {
    dataLoading.value = false
  }
}

// Watch for canvas ref to become available and render chart
watch(burndownChart, (newVal) => {
  console.log('Canvas ref changed:', !!newVal)
  if (newVal && projectData.value?.burndown_data) {
    console.log('Canvas now available, rendering chart')
    renderBurndownChart()
  }
})

// Render the burndown chart using Chart.js
function renderBurndownChart() {
  console.log('renderBurndownChart called', {
    hasProjectData: !!projectData.value,
    hasBurndownData: !!projectData.value?.burndown_data,
    dataLength: projectData.value?.burndown_data?.length,
    hasCanvasRef: !!burndownChart.value
  })
  
  if (!projectData.value || !projectData.value.burndown_data || projectData.value.burndown_data.length === 0) {
    console.log('No data to render chart')
    return
  }
  
  if (!burndownChart.value) {
    console.log('No canvas ref available')
    return
  }

  // Destroy existing chart instance
  if (chartInstance) {
    console.log('Destroying existing chart')
    chartInstance.destroy()
    chartInstance = null
  }

  const ctx = burndownChart.value.getContext('2d')
  if (!ctx) {
    console.log('Could not get 2d context')
    return
  }

  const data = projectData.value.burndown_data
  const labels = data.map(d => formatDate(d.date))
  
  console.log('Creating chart with', data.length, 'data points')

  const datasets = []
  const totalBudget = projectData.value.total_budget_dollars

  // Ideal linear burn-up (from 0 to total budget)
  datasets.push({
    label: 'Budget Target',
    data: data.map(d => totalBudget - d.planned_budget_remaining),
    borderColor: 'rgb(156, 163, 175)', // gray
    backgroundColor: 'rgba(156, 163, 175, 0.1)',
    borderWidth: 2,
    borderDash: [5, 5],
    tension: 0.1,
    pointRadius: 0,
    fill: false
  })

  // Actual spent line (cumulative fees - the actual burn-up)
  const actualData = data.map(d => d.actual_spent)
  
  datasets.push({
    label: 'Actual Spent',
    data: actualData,
    borderColor: 'rgb(59, 130, 246)', // blue
    backgroundColor: 'rgba(59, 130, 246, 0.1)',
    borderWidth: 2,
    tension: 0.1,
    pointRadius: 1,
    fill: false,
    segment: {
      borderColor: (ctx: any) => {
        // Change color to red when line goes above budget
        const yValue = ctx.p1.parsed.y
        return yValue > totalBudget ? 'rgb(239, 68, 68)' : 'rgb(59, 130, 246)'
      }
    }
  })

  // Invoiced amount (stepwise - only changes when invoice is added)
  datasets.push({
    label: 'Invoiced',
    data: data.map(d => d.invoiced_amount),
    borderColor: 'rgb(147, 51, 234)', // purple
    backgroundColor: 'rgba(147, 51, 234, 0.1)',
    borderWidth: 2,
    stepped: 'before' as const, // Makes it stepwise
    pointRadius: 1,
    fill: false
  })

  console.log('Attempting to create Chart instance...')
  
  // Find today's position in the chart
  const today = new Date()
  const todayStr = today.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  const todayIndex = labels.indexOf(todayStr)
  
  console.log('Today indicator:', todayStr, 'at index', todayIndex)
  
  try {
    chartInstance = new Chart(ctx, {
      type: 'line',
      data: {
        labels,
        datasets
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: {
          mode: 'index',
          intersect: false
        },
        plugins: {
          annotation: {
            annotations: {
              ...(todayIndex >= 0 ? {
                todayLine: {
                  type: 'line',
                  xMin: todayIndex,
                  xMax: todayIndex,
                  borderColor: 'rgb(239, 68, 68)',
                  borderWidth: 2,
                  borderDash: [5, 5],
                  label: {
                    content: 'Today',
                    display: true,
                    position: 'start',
                    backgroundColor: 'rgb(239, 68, 68)',
                    color: 'white',
                    font: {
                      size: 10,
                      weight: 'bold'
                    },
                    padding: 4
                  }
                }
              } : {}),
              budgetLine: {
                type: 'line',
                yMin: totalBudget,
                yMax: totalBudget,
                borderColor: 'rgb(55, 65, 81)',
                borderWidth: 2,
                borderDash: [3, 3],
                label: {
                  content: 'Budget Limit',
                  display: true,
                  position: 'end',
                  backgroundColor: 'rgb(55, 65, 81)',
                  color: 'white',
                  font: {
                    size: 9,
                    weight: 'bold'
                  },
                  padding: 2
                }
              }
            }
          },
          title: {
            display: false
          },
          legend: {
            display: true,
            position: 'top',
            labels: {
              boxWidth: 12,
              padding: 15,
              font: {
                size: 11
              }
            }
          },
        tooltip: {
            callbacks: {
              label: function (context: any) {
                let label = context.dataset.label || ''
                if (label) {
                  label += ': '
                }
                if (context.parsed.y !== null) {
                  label += '$' + Math.round(context.parsed.y).toLocaleString('en-US')
                }
                return label
              }
            }
          }
        },
        scales: {
        y: {
          beginAtZero: true,
          ticks: {
            callback: function (value: any) {
              return '$' + Math.round(value).toLocaleString('en-US')
            },
            font: {
              size: 10
            }
          },
          grid: {
            color: 'rgba(0, 0, 0, 0.05)'
          }
        },
          x: {
            ticks: {
              maxRotation: 45,
              minRotation: 45,
              font: {
                size: 9
              }
            },
            grid: {
              display: false
            }
          }
        }
      }
    })
    
    console.log('Chart created successfully!', chartInstance)
  } catch (error) {
    console.error('Error creating chart:', error)
    alert('Chart error: ' + error)
  }
}

// Format date for display
function formatDate(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}
</script>

<style scoped>
.text-sage {
  color: #718172;
}

.bg-sage {
  background-color: #718172;
}

.border-sage {
  border-color: #718172;
}

.focus\:border-sage:focus {
  border-color: #718172;
}

.focus\:ring-sage:focus {
  --tw-ring-color: #718172;
}
</style>

