<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { fetchProjects, createProject, updateProject } from '../../api/projects';
import { getUsers } from '../../api/timesheet';
import type { Project } from '../../types/Project';
// @ts-ignore - Ignore type issues with Vue components for now
import ProjectDrawer from '../../components/projects/ProjectDrawer.vue';
// @ts-ignore - Ignore type issues with Vue components for now
import ProjectCard from '../../components/projects/ProjectCard.vue';
import AssetUploaderModal from '../../components/assets/AssetUploaderModal.vue';
import { createAsset } from '../../api/assets';
import type { Asset } from '../../types/Asset';

// State
const projects = ref<Project[]>([]);
const isLoading = ref(true);
const error = ref<string | null>(null);
const isProjectDrawerOpen = ref(false);
const selectedProject = ref<Project | null>(null);
const staffMembers = ref<any[]>([]);

// New state for Asset Uploader
const isAssetUploaderOpen = ref(false);
const selectedProjectIdForAsset = ref<number | null>(null);

// Filter state
type FilterType = 'all' | 'active' | 'ended' | 'internal';
const activeFilter = ref<FilterType>('all');

// Helper function to check if a project is currently active
const isProjectActive = (project: Project): boolean => {
  if (!project) return false;
  
  // Parse dates
  const startDate = project.active_start ? new Date(project.active_start) : null;
  const endDate = project.active_end ? new Date(project.active_end) : null;
  
  // Get current date at midnight
  const now = new Date();
  now.setHours(0, 0, 0, 0);
  
  // Project is active if:
  // 1. Current date is after or equal to the start date (if a start date exists)
  // 2. Current date is before or equal to the end date (if an end date exists)
  const isAfterStart = startDate ? now >= startDate : true;
  const isBeforeEnd = endDate ? now <= endDate : true;
  
  return isAfterStart && isBeforeEnd;
};

// Helper function to check if a project has ended
const isProjectEnded = (project: Project): boolean => {
  if (!project) return false;
  
  const endDate = project.active_end ? new Date(project.active_end) : null;
  const now = new Date();
  now.setHours(0, 0, 0, 0);
  
  return endDate ? now > endDate : false;
};

// Computed property to filter and sort projects
const sortedProjects = computed(() => {
  if (!projects.value || !Array.isArray(projects.value)) {
    return [];
  }
  
  // Filter based on active filter
  let filtered = [...projects.value];
  
  if (activeFilter.value === 'active') {
    filtered = filtered.filter(p => isProjectActive(p));
  } else if (activeFilter.value === 'ended') {
    filtered = filtered.filter(p => isProjectEnded(p));
  } else if (activeFilter.value === 'internal') {
    filtered = filtered.filter(p => p.internal);
  }
  
  // Sort active projects first, then inactive
  return filtered.sort((a, b) => {
    const isAActive = isProjectActive(a);
    const isBActive = isProjectActive(b);
    
    if (isAActive && !isBActive) return -1;
    if (!isAActive && isBActive) return 1;
    
    return 0;
  });
});

// Count projects by status
const projectCounts = computed(() => {
  if (!projects.value || !Array.isArray(projects.value)) {
    return { all: 0, active: 0, ended: 0, internal: 0 };
  }
  
  return {
    all: projects.value.length,
    active: projects.value.filter(p => isProjectActive(p)).length,
    ended: projects.value.filter(p => isProjectEnded(p)).length,
    internal: projects.value.filter(p => p.internal).length,
  };
});

// Fetch projects function
const fetchProjectsData = async () => {
  isLoading.value = true;
  error.value = null;
  
  try {
    // Use exported fetchProjects function
    const response = await fetchProjects();
    
    if (!response || !Array.isArray(response)) {
      console.error('Invalid response format - expected array but got:', typeof response);
      error.value = 'Invalid response format from API';
      projects.value = [];
      return;
    }
    
    // Map the response to ensure it has the expected structure
    projects.value = response.map((apiProject: any) => {
      const mappedAssets = apiProject.assets || [];
      
      return {
        ID: apiProject.ID,
        name: apiProject.name || '',
        account_id: apiProject.account_id || 0,
        account: apiProject.account || { ID: 0, name: 'Unknown' },
        active_start: apiProject.active_start || '',
        active_end: apiProject.active_end || '',
        budget_hours: Number(apiProject.budget_hours) || 0,
        budget_dollars: Number(apiProject.budget_dollars) || 0,
        budget_cap_hours: Number(apiProject.budget_cap_hours) || 0,
        budget_cap_dollars: Number(apiProject.budget_cap_dollars) || 0,
        internal: !!apiProject.internal,
        billing_frequency: apiProject.billing_frequency || '',
        project_type: apiProject.project_type || '',
        ae_id: apiProject.ae_id !== undefined ? Number(apiProject.ae_id) : undefined,
        sdr_id: apiProject.sdr_id !== undefined ? Number(apiProject.sdr_id) : undefined,
        description: apiProject.description || '',
        staffing_assignments: (apiProject.staffing_assignments || []).map((sa: any) => ({
          ...sa,
          employee_id: Number(sa.employee_id),
          commitment: sa.commitment !== undefined ? Number(sa.commitment) : undefined,
          start_date: sa.start_date || undefined,
          end_date: sa.end_date || undefined,
        })),
        billing_codes: apiProject.billing_codes || [],
        assets: mappedAssets,
        // We don't need to map CreatedAt, UpdatedAt, DeletedAt as they're optional
      };
    });
  } catch (err) {
    console.error('Error fetching projects:', err);
    error.value = 'Failed to load projects. Please try again.';
    projects.value = [];
  } finally {
    isLoading.value = false;
  }
};

// Fetch projects on component mount
onMounted(async () => {
  await fetchProjectsData();
  
  // Fetch staff members for AE and SDR mapping
  try {
    const staff = await getUsers();
    staffMembers.value = staff || [];
  } catch (err) {
    console.error('Error fetching staff members:', err);
  }
});

// Project drawer functions
const openProjectDrawer = (project: Project | null = null) => {
  selectedProject.value = project;
  isProjectDrawerOpen.value = true;
};

const closeProjectDrawer = () => {
  isProjectDrawerOpen.value = false;
  selectedProject.value = null;
};

// Function to edit a project
const editProject = (project: Project) => {
  openProjectDrawer(project);
};

// Save project
const saveProject = async (projectData: Project) => {
  try {
    
    if (projectData.ID && projectData.ID > 0) {
      // Use exported updateProject function
      await updateProject(Number(projectData.ID), projectData);
    } else {
      // Use exported createProject function
      await createProject(projectData);
    }
    
    // Refresh projects
    await fetchProjectsData();
    
    // Close drawer
    closeProjectDrawer();
  } catch (error) {
    console.error('Error saving project:', error);
    alert('Failed to save project. Please try again.');
  }
};

// Asset Uploader Modal Functions for Projects
const openAssetUploaderForProject = (projectId: number) => {
  selectedProjectIdForAsset.value = projectId;
  isAssetUploaderOpen.value = true;
};

const closeAssetUploaderModal = () => {
  isAssetUploaderOpen.value = false;
  selectedProjectIdForAsset.value = null;
  // Also reset selectedProject if it was tied to opening the asset uploader indirectly
  // For now, we assume it's independent or handled by ProjectDrawer logic already.
};

const handleSaveAssetForProject = async (assetData: Asset) => {
  if (!selectedProjectIdForAsset.value) {
    alert('No project selected for the asset.');
    return;
  }
  const dataToSave = { ...assetData, project_id: selectedProjectIdForAsset.value };

  try {
    await createAsset(dataToSave);
    // Asset created successfully - silently close modal
    closeAssetUploaderModal();
  } catch (error) {
    console.error('Error creating asset for project:', error);
    alert('Failed to create asset. Please try again.');
  }
};
</script>

<template>
  <div class="px-4 py-4 sm:px-6 lg:px-8">
    <div class="sm:flex sm:items-center">
      <div class="sm:flex-auto">
        <h1 class="text-base font-semibold leading-6 text-blue">Projects</h1>
        <p class="mt-1 text-xs text-gray">A list of all projects including their status, client, and dates.</p>
      </div>
      <div class="mt-3 sm:ml-16 sm:mt-0 sm:flex-none">
        <button
          type="button"
          @click="openProjectDrawer()"
          class="block rounded-md bg-sage px-2.5 py-1.5 text-center text-xs font-semibold text-white shadow-sm hover:bg-sage-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sage"
        >
          <i class="fas fa-plus-circle mr-1"></i> Create new project
        </button>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="isLoading" class="flex flex-col items-center justify-center p-6 bg-white rounded-lg shadow mt-4">
      <i class="fas fa-spinner fa-spin text-3xl text-teal mb-2"></i>
      <span class="text-sm text-gray-dark">Loading projects...</span>
    </div>
    
    <!-- Error state -->
    <div v-else-if="error" class="flex flex-col items-center justify-center p-6 bg-white rounded-lg shadow mt-4">
      <i class="fas fa-exclamation-circle text-3xl text-red mb-2"></i>
      <span class="text-sm text-gray-dark mb-1">{{ error }}</span>
      <button @click="fetchProjectsData" class="btn-secondary mt-3">
        <i class="fas fa-sync mr-1"></i> Retry
      </button>
    </div>
    
    <!-- Empty state -->
    <div v-else-if="projects.length === 0" class="flex flex-col items-center justify-center p-6 bg-white rounded-lg shadow mt-4">
      <i class="fas fa-project-diagram text-4xl text-teal mb-2"></i>
      <p class="text-base font-medium text-gray-dark">No projects found</p>
      <p class="text-sm text-gray mb-3">Projects will appear here once they are created</p>
    </div>
    
    <!-- Filter Buttons -->
    <div v-else class="mt-4 flex gap-2 flex-wrap">
      <button
        @click="activeFilter = 'all'"
        :class="[
          'px-3 py-1.5 text-xs font-medium rounded-md transition-colors',
          activeFilter === 'all'
            ? 'bg-sage text-white'
            : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
        ]"
      >
        All <span class="ml-1 text-xs opacity-75">({{ projectCounts.all }})</span>
      </button>
      <button
        @click="activeFilter = 'active'"
        :class="[
          'px-3 py-1.5 text-xs font-medium rounded-md transition-colors',
          activeFilter === 'active'
            ? 'bg-sage text-white'
            : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
        ]"
      >
        Active <span class="ml-1 text-xs opacity-75">({{ projectCounts.active }})</span>
      </button>
      <button
        @click="activeFilter = 'ended'"
        :class="[
          'px-3 py-1.5 text-xs font-medium rounded-md transition-colors',
          activeFilter === 'ended'
            ? 'bg-sage text-white'
            : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
        ]"
      >
        Ended <span class="ml-1 text-xs opacity-75">({{ projectCounts.ended }})</span>
      </button>
      <button
        @click="activeFilter = 'internal'"
        :class="[
          'px-3 py-1.5 text-xs font-medium rounded-md transition-colors',
          activeFilter === 'internal'
            ? 'bg-sage text-white'
            : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
        ]"
      >
        Internal <span class="ml-1 text-xs opacity-75">({{ projectCounts.internal }})</span>
      </button>
    </div>
    
    <!-- Projects List - Full Width -->
    <div v-if="!isLoading && projects.length > 0" class="mt-4 flow-root">
      <ul role="list" class="grid grid-cols-1 gap-4">
        <li v-for="project in sortedProjects" :key="project.ID">
          <ProjectCard 
            :project="project" 
            :staff-list="staffMembers"
            @edit="editProject"
            @add-asset="openAssetUploaderForProject"
            @project-updated="fetchProjectsData"
          />
        </li>
      </ul>
    </div>
    
    <!-- Project Drawer -->
    <ProjectDrawer
      :is-open="isProjectDrawerOpen"
      :project-data="selectedProject"
      @close="closeProjectDrawer"
      @save="saveProject"
    />

    <!-- Asset Uploader Modal for Projects -->
    <AssetUploaderModal 
      :is-open="isAssetUploaderOpen" 
      :project-id="selectedProjectIdForAsset"
      @close="closeAssetUploaderModal" 
      @save="handleSaveAssetForProject"
    />
  </div>
</template>

<style scoped>
.btn-secondary {
  display: inline-flex;
  align-items: center;
  padding: 0.375rem 0.75rem;
  border-radius: 0.375rem;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--color-gray-700);
  background-color: white;
  border: 1px solid var(--color-gray-300);
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background-color: var(--color-gray-50);
}
</style> 