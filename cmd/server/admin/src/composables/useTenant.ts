import { ref, computed, readonly } from 'vue';
import { fetchTenant, type Tenant } from '../api/tenant';

// Global state - shared across all components
const tenant = ref<Tenant | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);
const initialized = ref(false);

/**
 * Composable for managing tenant information
 * Provides reactive access to current tenant data and branding
 */
export function useTenant() {
  // Computed properties for easy access to tenant info
  const tenantName = computed(() => tenant.value?.name ?? 'Cronos');
  const tenantSlug = computed(() => tenant.value?.slug ?? '');
  const logoUrl = computed(() => tenant.value?.branding?.logo_url ?? '/branding/cronos-logo.png');
  const primaryColor = computed(() => tenant.value?.branding?.primary_color ?? '#3B82F6');
  const secondaryColor = computed(() => tenant.value?.branding?.secondary_color ?? '#1E40AF');
  
  /**
   * Load tenant information from the API
   */
  async function loadTenant() {
    // Don't reload if already initialized
    if (initialized.value && tenant.value) {
      return;
    }
    
    loading.value = true;
    error.value = null;
    
    try {
      tenant.value = await fetchTenant();
      initialized.value = true;
    } catch (err: any) {
      error.value = err.message || 'Failed to load tenant information';
      console.error('Failed to fetch tenant:', err);
      
      // If tenant fetch fails, might indicate user is on wrong subdomain or not logged in
      if (err.response?.status === 404) {
        error.value = 'Tenant not found. Please check your subdomain.';
      }
    } finally {
      loading.value = false;
    }
  }
  
  /**
   * Refresh tenant information (force reload)
   */
  async function refreshTenant() {
    initialized.value = false;
    await loadTenant();
  }
  
  return {
    // State (readonly to prevent external mutations)
    tenant: readonly(tenant),
    loading: readonly(loading),
    error: readonly(error),
    
    // Computed helpers
    tenantName,
    tenantSlug,
    logoUrl,
    primaryColor,
    secondaryColor,
    
    // Actions
    loadTenant,
    refreshTenant,
  };
}

