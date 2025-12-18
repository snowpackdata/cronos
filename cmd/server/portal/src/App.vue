<template>
  <div>
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
            <!-- Portal Title/Logo -->
            <span class="text-xl font-bold text-white border-b-2 border-sage pb-1">Client Portal</span> 
          </div>
          <div class="mt-5 flex flex-1 flex-col overflow-y-auto px-4">
            <nav class="flex-1 space-y-1">
              <ul role="list" class="flex flex-1 flex-col gap-y-7">
                <li>
                  <ul role="list" class="-mx-2 space-y-1">
                    <li v-for="item in navigationItems" :key="item.path">
                      <router-link :to="item.path" :class="[
                        $route.path === item.path 
                          ? 'bg-sage-dark text-white shadow-md' 
                          : 'text-gray-400 hover:bg-sage hover:text-white', 
                        'group flex gap-x-3 rounded-md p-2.5 text-sm font-semibold transition-all duration-200'
                      ]">
                        <i :class="['fas', item.icon, 'h-5 w-5 shrink-0']" aria-hidden="true" />
                        {{ item.name }}
                      </router-link>
                    </li>
                  </ul>
                </li>
              </ul>
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
           <!-- Portal Title/Logo -->
           <span class="text-xl font-bold text-white border-b-2 border-sage pb-1">Client Portal</span> 
        </div>
        <nav class="flex flex-1 flex-col">
          <ul role="list" class="flex flex-1 flex-col gap-y-7">
            <li>
              <ul role="list" class="-mx-2 space-y-1">
                <li v-for="item in navigationItems" :key="item.path">
                  <router-link :to="item.path" :class="[
                    $route.path === item.path 
                      ? 'bg-sage-dark text-white shadow-md' 
                      : 'text-gray-400 hover:bg-sage hover:text-white', 
                    'group flex gap-x-3 rounded-md p-2.5 text-sm font-semibold transition-all duration-200'
                  ]">
                     <i :class="['fas', item.icon, 'h-5 w-5 shrink-0']" aria-hidden="true" />
                    {{ item.name }}
                  </router-link>
                </li>
              </ul>
            </li>
          </ul>
        </nav>
      </div>
    </div>

    <div class="lg:pl-64">
      <div class="flex h-16 shrink-0 items-center gap-x-4 border-b border-gray bg-gray-800 px-4 shadow-md sm:gap-x-6 sm:px-6 lg:hidden">
        <button type="button" class="-m-2.5 p-2.5 text-gray-400 hover:text-white lg:hidden" @click.prevent="sidebarOpen = true">
          <span class="sr-only">Open sidebar</span>
          <i class="fas fa-bars h-6 w-6" aria-hidden="true" />
        </button>

        <!-- Separator -->
        <div class="h-6 w-px bg-gray-700 lg:hidden" aria-hidden="true" />

        <!-- Mobile title -->
        <div class="lg:hidden flex-1 text-center">
          <span class="text-xl font-bold text-white">Client Portal</span>
        </div>

        <div class="hidden lg:flex lg:flex-1 gap-x-4 self-stretch lg:gap-x-6">
          <div class="flex items-center gap-x-4 lg:gap-x-6">
            <!-- Separator -->
            <div class="h-6 w-px bg-gray-700" aria-hidden="true" />
          </div>
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
import { ref, onMounted } from 'vue';
import { useTenant } from './composables/useTenant';
import { getToken } from './api/index';
// import { useRoute } from 'vue-router'; // Not using dynamic title yet

// Load tenant information
const { loadTenant } = useTenant();

// Portal Navigation Items (Update icons as needed)
const navigationItems = [
  { name: 'Dashboard', path: '/dashboard', icon: 'fa-tachometer-alt' }, // Example icon
  { name: 'Invoices', path: '/invoices', icon: 'fa-file-invoice-dollar' }, // Example icon
  { name: 'Projects', path: '/projects', icon: 'fa-folder-open' }, // Example icon
  { name: 'Settings', path: '/settings', icon: 'fa-cog' } // Added Settings navigation item
];

// State for mobile sidebar
const sidebarOpen = ref(false);

// Load tenant on mount if authenticated
onMounted(async () => {
  const token = getToken();
  if (token) {
    await loadTenant();
  }
});

// Note: Excluded route watching for title from admin App.vue for simplicity

</script>

<style scoped>
/* Ensure clickable elements aren't covered by anything */
a, button {
  position: relative;
  z-index: 1;
}

/* Custom color classes - these might be redundant if defined in global style.css */
:root {
  --sage-green-primary: #58837e;
  --sage-green-dark: #476b67;
  --sage-green-light: #76a19c;
  --sage-green-pale: #e6efee;
}
</style> 