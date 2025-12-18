<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Mobile sidebar -->
    <div class="fixed inset-0 z-50 lg:hidden" v-show="sidebarOpen" @click="sidebarOpen = false">
      <div class="fixed inset-0 bg-gray-900/80" />
      <div class="fixed inset-y-0 left-0 flex max-w-xs w-full" @click.stop>
        <div class="relative mr-16 flex w-full max-w-xs flex-1 flex-col bg-gray-900 pt-5 pb-4">
          <div class="absolute right-0 top-0 -mr-12 pt-2">
            <button type="button" class="ml-1 flex h-10 w-10 items-center justify-center rounded-full focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white" @click="sidebarOpen = false">
              <span class="sr-only">Close sidebar</span>
              <i class="fas fa-times h-2 w-2 text-white" aria-hidden="true" />
            </button>
          </div>
          <div class="flex-shrink-0 flex items-center px-4">
            <span class="text-xl font-bold text-white border-b-2 border-sage pb-1">Cronos</span>
          </div>
          <div class="mt-5 flex flex-1 flex-col overflow-y-auto px-4">
            <nav class="flex flex-1 flex-col">
              <ul role="list" class="flex flex-1 flex-col gap-y-1">
                <li v-for="item in navigationSections" :key="item.name">
                  <!-- Direct link (no sub-items) -->
                  <router-link v-if="item.path" :to="item.path" :class="[
                        $route.path === item.path
                          ? 'bg-sage-dark text-white shadow-md'
                          : 'text-gray-400 hover:bg-sage hover:text-white',
                    'group flex gap-x-2 rounded-md px-2 py-2 text-xs font-medium transition-all duration-200'
                      ]">
                    <i :class="['fas', item.icon, 'h-4 w-4 shrink-0']" aria-hidden="true" />
                        {{ item.name }}
                  </router-link>
                  
                  <!-- Expandable section (has sub-items) -->
                  <div v-else :class="[
                    'relative rounded-md transition-all duration-200',
                    expandedSections.has(item.name) ? 'bg-black/20' : ''
                  ]">
                    <div v-if="expandedSections.has(item.name)" class="absolute left-0 top-0 bottom-0 w-0.5 bg-sage rounded-full"></div>
                    <button
                      @click="toggleSection(item.name)"
                      :class="[
                        'w-full flex items-center gap-x-2 px-2 py-2 text-xs font-medium rounded-md transition-all duration-200',
                        expandedSections.has(item.name) ? 'text-white' : 'text-gray-400 hover:text-white hover:bg-black/30'
                      ]"
                    >
                      <i :class="[
                        'fas fa-chevron-right h-3 w-3 transition-transform duration-200',
                        expandedSections.has(item.name) ? 'rotate-90' : ''
                      ]" />
                      <i :class="['fas', item.icon, 'h-4 w-4']" />
                      <span>{{ item.name }}</span>
                    </button>
                    <ul v-show="expandedSections.has(item.name)" class="ml-7 mt-1 space-y-0.5 pb-1">
                      <li v-for="subItem in (item as any).items" :key="subItem.path">
                        <router-link :to="subItem.path" :class="[
                          $route.path === subItem.path
                            ? 'bg-sage-dark text-white shadow-md'
                            : 'text-gray-400 hover:bg-sage hover:text-white',
                          'group flex gap-x-2 rounded-md px-2 py-1.5 text-xs font-medium transition-all duration-200'
                        ]">
                          <i :class="['fas', subItem.icon, 'h-3 w-3 shrink-0']" aria-hidden="true" />
                          {{ subItem.name }}
                      </router-link>
                    </li>
                  </ul>
                  </div>
                </li>
              </ul>
              <!-- Logout button at bottom of mobile sidebar -->
              <div class="mt-auto pt-4 border-t border-gray-700">
                <button
                  @click="logout"
                  class="w-full flex items-center gap-x-2 px-2 py-2 text-xs font-medium text-gray-400 hover:bg-red-600 hover:text-white rounded-md transition-all duration-200"
                >
                  <i class="fas fa-sign-out-alt h-4 w-4 shrink-0"></i>
                  <span>Logout</span>
                </button>
              </div>
            </nav>
          </div>
        </div>
      </div>
    </div>

    <!-- Static sidebar for desktop -->
    <div class="hidden lg:fixed lg:inset-y-0 lg:z-50 lg:flex lg:w-64 lg:flex-col">
      <!-- Sidebar component, for desktop -->
      <div class="flex grow flex-col gap-y-5 overflow-y-auto bg-gray-900 shadow-xl px-6 pb-4">
        <div class="flex h-16 shrink-0 items-center">
          <span class="text-xl font-bold text-white border-b-2 border-sage pb-1">Cronos</span>
        </div>
        <nav class="flex flex-1 flex-col">
          <ul role="list" class="flex flex-1 flex-col gap-y-1">
            <li v-for="item in navigationSections" :key="item.name">
              <!-- Direct link (no sub-items) -->
              <router-link v-if="item.path" :to="item.path" :class="[
                    $route.path === item.path
                      ? 'bg-sage-dark text-white shadow-md'
                      : 'text-gray-400 hover:bg-sage hover:text-white',
                'group flex gap-x-2 rounded-md px-2 py-2 text-xs font-medium transition-all duration-200'
                  ]">
                <i :class="['fas', item.icon, 'h-4 w-4 shrink-0']" aria-hidden="true" />
                    {{ item.name }}
              </router-link>
              
              <!-- Expandable section (has sub-items) -->
              <div v-else :class="[
                'relative rounded-md transition-all duration-200',
                expandedSections.has(item.name) ? 'bg-black/20' : ''
              ]">
                <div v-if="expandedSections.has(item.name)" class="absolute left-0 top-0 bottom-0 w-0.5 bg-sage rounded-full"></div>
                <button
                  @click="toggleSection(item.name)"
                  :class="[
                    'w-full flex items-center gap-x-2 px-2 py-2 text-xs font-medium rounded-md transition-all duration-200',
                    expandedSections.has(item.name) ? 'text-white' : 'text-gray-400 hover:text-white hover:bg-black/30'
                  ]"
                >
                  <i :class="[
                    'fas fa-chevron-right h-3 w-3 transition-transform duration-200',
                    expandedSections.has(item.name) ? 'rotate-90' : ''
                  ]" />
                  <i :class="['fas', item.icon, 'h-4 w-4']" />
                  <span>{{ item.name }}</span>
                </button>
                <ul v-show="expandedSections.has(item.name)" class="ml-7 mt-1 space-y-0.5 pb-1">
                  <li v-for="subItem in (item as any).items" :key="subItem.path">
                    <router-link :to="subItem.path" :class="[
                      $route.path === subItem.path
                        ? 'bg-sage-dark text-white shadow-md'
                        : 'text-gray-400 hover:bg-sage hover:text-white',
                      'group flex gap-x-2 rounded-md px-2 py-1.5 text-xs font-medium transition-all duration-200'
                    ]">
                      <i :class="['fas', subItem.icon, 'h-3 w-3 shrink-0']" aria-hidden="true" />
                      {{ subItem.name }}
                  </router-link>
                </li>
              </ul>
              </div>
            </li>
          </ul>
          <!-- Logout button at bottom of sidebar -->
          <div class="mt-auto pt-4 border-t border-gray-700">
            <button
              @click="logout"
              class="w-full flex items-center gap-x-2 px-2 py-2 text-xs font-medium text-gray-400 hover:bg-red-600 hover:text-white rounded-md transition-all duration-200"
            >
              <i class="fas fa-sign-out-alt h-4 w-4 shrink-0"></i>
              <span>Logout</span>
            </button>
          </div>
        </nav>
      </div>
    </div>

    <div class="lg:pl-64">
      <!-- Top navbar - Mobile only (no logout button) -->
      <div class="flex h-16 shrink-0 items-center gap-x-4 border-b border-gray bg-gray-800 px-4 shadow-md sm:gap-x-6 sm:px-6 lg:hidden">
        <button type="button" class="-m-2.5 p-2.5 text-gray-400 hover:text-white lg:hidden" @click.prevent="sidebarOpen = true">
          <span class="sr-only">Open sidebar</span>
          <i class="fas fa-bars h-6 w-6" aria-hidden="true" />
        </button>

        <!-- Separator -->
        <div class="h-6 w-px bg-gray-700 lg:hidden" aria-hidden="true" />

        <!-- Mobile title -->
        <div class="lg:hidden flex-1 text-center">
          <span class="text-xl font-bold text-white">Cronos</span>
        </div>
      </div>

      <main class="py-10 bg-gray-50 min-h-screen lg:py-0 lg:pt-6 lg:min-h-[100vh]">
        <div class="px-4 sm:px-6 lg:px-8">
          <router-view />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { jwtDecode } from 'jwt-decode';
import { useTenant } from './composables/useTenant';
import { getToken } from './api/apiUtils';

// Add TypeScript declaration for import.meta.env
declare interface ImportMeta {
  readonly env: {
    readonly MODE: string
    readonly BASE_URL: string
    readonly DEV: boolean
    readonly PROD: boolean
    // Add any other environment variables used in your app
    readonly [key: string]: string | boolean | undefined
  }
}

interface DecodedToken {
  Email: string;
  IsStaff: boolean;
  UID: number;
  Role: string;
  exp: number;
  iat: number;
}

const route = useRoute();

// Load tenant information
const { loadTenant } = useTenant();

// Handle token from URL query parameter (from login redirect)
const handleTokenFromURL = () => {
  const token = route.query.token as string;
  
  if (token) {
    console.log('Token found in URL, storing and cleaning...');
    
    // Store token in localStorage
    localStorage.setItem('snowpack_token', token);
    
    // Also store in cookie for server-side access
    const expiryDate = new Date();
    expiryDate.setDate(expiryDate.getDate() + 30);
    const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
    const domainParam = isLocalhost ? '' : `; domain=${window.location.hostname}`;
    const secure = window.location.protocol === 'https:' ? '; Secure' : '';
    document.cookie = `x-access-token=${token}; expires=${expiryDate.toUTCString()}; path=/${domainParam}${secure}; SameSite=Lax`;
    
    // Remove token from URL - use window.history for immediate effect
    const cleanUrl = window.location.pathname + window.location.hash;
    window.history.replaceState({}, '', cleanUrl);
    
    console.log('Token stored and URL cleaned');
  }
};

// Load tenant on mount if authenticated
onMounted(async () => {
  handleTokenFromURL();
  const token = getToken();
  if (token) {
    await loadTenant();
  }
});

// Logout function
const logout = () => {
  // Clear localStorage
  localStorage.removeItem('snowpack_token');
  
  // Clear cookie
  document.cookie = 'x-access-token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
  
  // Redirect to login (no subdomain)
  const currentProtocol = window.location.protocol;
  const currentPort = window.location.port;
  const portPart = currentPort ? `:${currentPort}` : '';
  window.location.href = `${currentProtocol}//localhost${portPart}/login`;
};

// Define admin navigation structure with mix of direct links and expandable sections
const adminNavigationSections = [
  // Direct links (no expansion)
  { name: 'Timesheet', path: '/timesheet', icon: 'fa-clock' },
  { name: 'Expenses', path: '/expenses', icon: 'fa-receipt' },
  // Expandable sections
  {
    name: 'Accounting',
    icon: 'fa-book',
    items: [
  { name: 'Accounts Receivable', path: '/accounts-receivable', icon: 'fa-file-invoice-dollar' },
  { name: 'Accounts Payable', path: '/accounts-payable', icon: 'fa-file-invoice' },
  { name: 'General Ledger', path: '/accounting', icon: 'fa-book' },
      { name: 'Chart of Accounts', path: '/chart-of-accounts', icon: 'fa-list-alt' },
      { name: 'Offline Journals', path: '/offline-journals', icon: 'fa-file-import' },
    ]
  },
  {
    name: 'Organization',
    icon: 'fa-building',
    items: [
  { name: 'Accounts', path: '/accounts', icon: 'fa-building' },
  { name: 'Projects', path: '/projects', icon: 'fa-bars-progress' },
  { name: 'Team', path: '/staff', icon: 'fa-users' },
  { name: 'Billing Codes', path: '/billing-codes', icon: 'fa-barcode' },
  { name: 'Rates', path: '/rates', icon: 'fa-percent' },
      { name: 'Expenses', path: '/expense-config', icon: 'fa-tags' },
      { name: 'Settings', path: '/settings', icon: 'fa-cog' },
    ]
  },
  {
    name: 'Planning',
    icon: 'fa-calendar-alt',
    items: [
      { name: 'Timesheet Admin', path: '/timesheet-admin', icon: 'fa-toolbox' },
      { name: 'Capacity', path: '/capacity', icon: 'fa-chart-gantt' },
      { name: 'Expense Approvals', path: '/expense-approvals', icon: 'fa-receipt' },
      { name: 'Recurring Compensation', path: '/recurring-entries', icon: 'fa-repeat' },
    ]
  }
];

// Define staff-only navigation (direct links only)
const staffNavigationSections = [
  { name: 'Timesheet', path: '/timesheet', icon: 'fa-clock' },
  { name: 'Expenses', path: '/expenses', icon: 'fa-receipt' }
];

// Computed navigation sections based on user role
// Watch route.path to trigger re-evaluation after router stores token
const navigationSections = computed(() => {
  // Access route.path to make this reactive to navigation
  route.path;
  
  const token = localStorage.getItem('snowpack_token');
  if (!token) return staffNavigationSections;

  try {
    const decodedToken = jwtDecode<DecodedToken>(token);
    return decodedToken.Role === 'ADMIN' ? adminNavigationSections : staffNavigationSections;
  } catch (error) {
    console.error('Error decoding token:', error);
    return staffNavigationSections;
  }
});

// State for mobile sidebar
const sidebarOpen = ref(false);

// State for expanded sections (only for sections with sub-items) - all collapsed by default
const expandedSections = ref<Set<string>>(new Set());

// Toggle section expansion
function toggleSection(sectionName: string) {
  if (expandedSections.value.has(sectionName)) {
    expandedSections.value.delete(sectionName);
  } else {
    expandedSections.value.add(sectionName);
  }
}

// Dynamically set page title based on current route
const pageTitle = computed(() => {
  if (route.name) {
    return String(route.name);
  }
  return 'Dashboard';
});

// Set document title when route changes
watch(() => route.name, () => {
  document.title = `Cronos - ${pageTitle.value}`;
}, { immediate: true });

// Simple function to test the authentication token
const testToken = () => {
  // Don't run the token test on login pages
  const currentPath = window.location.pathname;
  if (currentPath.includes('/login') || currentPath.includes('/register')) {
    return;
  }

  const token = localStorage.getItem('snowpack_token');
  const isDevelopment = import.meta.env.DEV ||
                     window.location.hostname === 'localhost' ||
                     window.location.hostname === '127.0.0.1';

  // In development mode, we can continue even without a token
  // Our middleware will handle this case
  if (!token) {
    // Only redirect to login in production - in development we have auto-login middleware
    if (!isDevelopment) {
      window.location.href = '/login';
      return;
    } else {
      return;
    }
  }

  // Check token validity by making a request to the API
  const apiUrl = '/api/projects';

  // Use the standard API pattern with proxy
  fetch(apiUrl, {
    headers: {
      'x-access-token': token
    }
  })
  .then(response => {

    if (response.ok) {
      // Check if the response is JSON before trying to parse it
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        return response.json();
      } else {
        // If in development, just continue with an empty object
        if (import.meta.env.DEV) {
          return response.text().then(text => {
            return {};
          });
        } else {
          // In production, treat non-JSON as an error
          throw new Error(`Expected JSON response but got ${contentType}`);
        }
      }
    } else {
      // Only clear token on 401 unauthorized, not on 403 forbidden
      // 403 could mean the token is valid but the user lacks permission
      if (response.status === 401) {
        localStorage.removeItem('snowpack_token');
      }

      return response.text().then(text => {
        throw new Error(`API call failed with status ${response.status}`);
      });
    }
  })
  .then(data => {
    // Token is valid, nothing to do
  })
  .catch(error => {
    // Error handled silently
  });
};

// Call the test function on app initialization
setTimeout(testToken, 1000);
</script>

<style scoped>
/* Ensure clickable elements aren't covered by anything */
a, button {
  position: relative;
  z-index: 1;
}

/* Custom color classes for the muted green */
:root {
  --sage-green-primary: #58837e;
  --sage-green-dark: #476b67;
  --sage-green-light: #76a19c;
  --sage-green-pale: #e6efee;
}
</style>
