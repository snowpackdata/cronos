import type { Project } from '../types/Project';
import { 
  fetchAll, 
  fetchById, 
  remove, 
  createWithFormData, 
  updateWithFormData
} from './apiUtils';

/**
 * Validates project data before sending to the API
 * @param project - Project data to validate
 * @returns Object with validation result and any error messages
 */
function validateProject(project: Partial<Project>): { isValid: boolean; errors: string[] } {
  const errors: string[] = [];
  
  // Check required fields
  if (!project.name) errors.push('Project name is required');
  if (!project.account_id && (!project.account || !project.account.ID)) errors.push('Account is required');
  if (!project.active_start) errors.push('Start date is required');
  if (!project.active_end) errors.push('End date is required');
  if (!project.billing_frequency) errors.push('Billing frequency is required');
  
  // Validate date range
  if (project.active_start && project.active_end) {
    const startDate = new Date(project.active_start);
    const endDate = new Date(project.active_end);
    if (endDate < startDate) {
      errors.push('End date must be after start date');
    }
  }
  
  return { 
    isValid: errors.length === 0,
    errors 
  };
}

/**
 * Prepares project data for the backend by transforming it to the expected format
 * @param project - Project data to transform
 * @returns Prepared project data as FormData
 */
function prepareProjectForApi(project: Project): FormData {
  const formData = new FormData();
  
  // Set required fields
  formData.set("name", project.name);
  
  // Get the account ID from either direct ID or account object
  const accountId = project.account_id || (project.account ? project.account.ID : 0);
  formData.set("account_id", accountId.toString());
  
  formData.set("budget_hours", project.budget_hours.toString());
  formData.set("budget_dollars", project.budget_dollars.toString());
  formData.set("active_start", project.active_start);
  formData.set("active_end", project.active_end);
  formData.set("internal", project.internal.toString());
  formData.set("project_type", project.project_type);
  
  // Add billing_frequency
  if (project.billing_frequency) {
    formData.set("billing_frequency", project.billing_frequency);
  } else {
    formData.set("billing_frequency", "BILLING_TYPE_MONTHLY");
  }

  // Add description if it exists
  if (project.description !== undefined) { // Check for undefined to allow sending empty string
    formData.set("description", project.description);
  }
  
  // Add budget cap fields if they have values (assuming they are optional and can be number or undefined)
  // Temporary cast to any to bypass linter error until Project type is globally updated and recognized
  const projectWithCaps = project as any;
  if (projectWithCaps.budget_cap_hours !== undefined && projectWithCaps.budget_cap_hours !== null) {
    formData.set("budget_cap_hours", projectWithCaps.budget_cap_hours.toString());
  }
  if (projectWithCaps.budget_cap_dollars !== undefined && projectWithCaps.budget_cap_dollars !== null) {
    formData.set("budget_cap_dollars", projectWithCaps.budget_cap_dollars.toString());
  }
  
  // Only add optional fields if they have values
  if (project.ae_id) {
    formData.set("ae_id", project.ae_id.toString());
  }
  
  if (project.sdr_id) {
    formData.set("sdr_id", project.sdr_id.toString());
  }
  
  
  return formData;
}

/**
 * API service for project-related operations
 */
const projectsAPI = {
  /**
   * Get all projects
   * @returns Promise with array of projects
   */
  async getProjects(): Promise<Project[]> {
    return fetchAll<Project>('projects');
  },

  /**
   * Get a single project by ID
   * @param id - Project ID
   * @returns Promise with project data
   */
  async getProject(id: number): Promise<Project> {
    return fetchById<Project>('projects', id);
  },

  /**
   * Create a new project with validation
   * @param project - Project data
   * @returns Promise with created project
   * @throws Error if validation fails
   */
  async createProject(project: Project): Promise<Project> {
    // Validate project data
    const validation = validateProject(project);
    if (!validation.isValid) {
      throw new Error(`Validation failed: ${validation.errors.join(', ')}`);
    }
    
    try {
      // Prepare data and use the createWithFormData utility
      const formData = prepareProjectForApi(project);
      return createWithFormData<Project>('projects', formData);
    } catch (error) {
      console.error('Failed to create project:', error);
      throw error;
    }
  },

  /**
   * Update an existing project with validation
   * @param project - Project data to update
   * @returns Promise with updated project
   * @throws Error if validation fails
   */
  async updateProject(project: Project): Promise<Project> {
    // Validate project data
    const validation = validateProject(project);
    if (!validation.isValid) {
      throw new Error(`Validation failed: ${validation.errors.join(', ')}`);
    }
    
    try {
      // Prepare data and use the updateWithFormData utility
      const formData = prepareProjectForApi(project);
      return updateWithFormData<Project>('projects', project.ID, formData);
    } catch (error) {
      console.error('Failed to update project:', error);
      throw error;
    }
  },

  /**
   * Delete a project
   * @param id - ID of project to delete
   * @returns Promise with deletion result
   */
  async deleteProject(id: number): Promise<any> {
    try {
      return remove('projects', id);
    } catch (error) {
      console.error('Failed to delete project:', error);
      throw error;
    }
  },
  
  /**
   * Get projects for a specific account
   * @param accountId - ID of the account
   * @returns Promise with array of projects
   */
  async getAccountProjects(accountId: number): Promise<Project[]> {
    return fetchAll<Project>(`accounts/${accountId}/projects`);
  },
  
  /**
   * Get active projects
   * @returns Promise with array of active projects
   */
  async getActiveProjects(): Promise<Project[]> {
    return fetchAll<Project>('active_projects');
  },

  /**
   * Get billing codes for a project
   * @param id - Project ID
   * @returns Promise with billing codes
   */
  async getBillingCodes(id: number): Promise<any[]> {
    return fetchAll(`projects/${id}/billing_codes`);
  },

  /**
   * Fetches projects for a specific account
   * @param accountId Account ID to filter projects
   * @returns Promise with projects data
   */
  async getProjectsByAccount(accountId: number): Promise<Project[]> {
    return fetchAll<Project>(`projects?account_id=${accountId}`);
  }
};

/**
 * Fetches all projects from the API
 * @returns Promise with array of projects
 */
export const fetchProjects = async () => {
  try {
    return await fetchAll<Project>('projects');
  } catch (error) {
    console.error('Error fetching projects:', error);
    throw error;
  }
};

/**
 * Fetches a single project by ID from the API
 * @param id - Project ID
 * @returns Promise with project data
 */
export const fetchProjectById = async (id: number) => {
  try {
    return await fetchById<Project>('projects', id);
  } catch (error) {
    console.error(`Error fetching project with ID ${id}:`, error);
    throw error;
  }
};

/**
 * Creates a new project with validation
 * @param projectData Project data to create
 * @returns Promise with created project data
 * @throws Error if validation fails
 */
export const createProject = async (projectData: Partial<Project>) => {
  return projectsAPI.createProject(projectData as Project);
};

/**
 * Updates an existing project with validation
 * @param id Project ID
 * @param projectData Updated project data
 * @returns Promise with updated project data
 * @throws Error if validation fails
 */
export const updateProject = async (id: number, projectData: Partial<Project>) => {
  const fullProject = { ...projectData, ID: id } as Project;
  return projectsAPI.updateProject(fullProject);
};

/**
 * Deletes a project
 * @param id Project ID to delete
 * @returns Promise with deletion status
 */
export const deleteProject = async (id: number) => {
  return projectsAPI.deleteProject(id);
};

// Export the API object explicitly to match the accounts pattern
export { projectsAPI };

export default projectsAPI;

// Add getProjectAnalytics function to fetch analytics data for a specific project
export const getProjectAnalytics = async (projectId: number) => {
  try {
    // Import getToken from apiUtils
    const { getToken } = await import('./apiUtils');
    const token = getToken();
    
    // Prepare headers - only add token if it exists
    const headers: Record<string, string> = {};
    if (token) {
      headers['x-access-token'] = token;
    }
    
    // Add the token to request headers
    const response = await fetch(`/api/projects/${projectId}/analytics`, {
      headers
    });
    
    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Failed to fetch project analytics: ${error}`);
    }
    return await response.json();
  } catch (error) {
    console.error('Error fetching project analytics:', error);
    throw error;
  }
}; 