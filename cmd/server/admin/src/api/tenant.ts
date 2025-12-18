import { api } from './apiUtils';

export interface Tenant {
  id: number;
  slug: string;
  name: string;
  domain?: string;
  plan: string;
  branding: {
    logo_url?: string;
    primary_color?: string;
    secondary_color?: string;
  };
  settings: Record<string, any>;
}

/**
 * Fetch current tenant information
 */
export async function fetchTenant(): Promise<Tenant> {
  const response = await api.get<Tenant>('/api/tenant');
  return response.data;
}

/**
 * Update tenant information
 */
export async function updateTenant(data: Partial<Tenant>): Promise<Tenant> {
  const response = await api.put<Tenant>('/api/tenant', data);
  return response.data;
}

