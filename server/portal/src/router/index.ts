import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import Invoices from '../views/Invoices.vue'
import Projects from '../views/Projects.vue'
import Settings from '../views/Settings.vue'
import { refreshToken } from '../api/index'
// Placeholder for LoginView, you'll need to create this component
const LoginView = { template: '<div>Login Page - Implement Me</div>' }; 

const routes: Array<RouteRecordRaw> = [
  {
    path: '/login',
    name: 'Login',
    component: LoginView, // Using placeholder component
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    redirect: '/dashboard', // Redirects are fine like this
    meta: { requiresAuth: false } // The redirect itself doesn't require auth, the target does
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: Dashboard,
    meta: { requiresAuth: true }
  },
  {
    path: '/invoices',
    name: 'Invoices',
    component: Invoices,
    meta: { requiresAuth: true }
  },
  {
    path: '/projects',
    name: 'Projects',
    component: Projects,
    meta: { requiresAuth: true }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: Settings,
    meta: { requiresAuth: true }
  },
  // Catch-all for 404 - Optional but good practice
  // {
  //   path: '/:catchAll(.*)',
  //   name: 'NotFound',
  //   component: () => import('../views/NotFound.vue')
  // }
]

const router = createRouter({
  history: createWebHistory('/portal/'), // Important: Set base to /portal/
  routes
})

router.beforeEach(async (to, from, next) => {
  const isAuthenticated = !!localStorage.getItem('snowpack_token');
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth);

  if (requiresAuth && !isAuthenticated) {
    // If route requires auth and user is not authenticated, redirect to the global /login page.
    window.location.href = '/login'; 
    // It's good practice to call next(false) to cancel the current navigation after triggering a manual redirect,
    // though window.location.href often makes it a moot point by reloading.
    next(false); 
  } else if (to.name === 'Login' && isAuthenticated) {
    // If user is authenticated and tries to access SPA's own /login route (which now shouldn't be hit often if redirecting globally),
    // redirect to dashboard.
    next({ name: 'Dashboard' });
  } else if (isAuthenticated && requiresAuth) {
    // Check if token is close to expiring and refresh if needed
    try {
      const token = localStorage.getItem('snowpack_token');
      if (token) {
        const payload = JSON.parse(atob(token.split('.')[1]));
        const expirationTime = payload.exp * 1000; // Convert to milliseconds
        const currentTime = Date.now();
        const sevenDaysInMs = 7 * 24 * 60 * 60 * 1000;
        
        // If token expires within 7 days, try to refresh it
        if (expirationTime - currentTime < sevenDaysInMs) {
          const refreshSuccess = await refreshToken();
          if (!refreshSuccess) {
            // If refresh failed, redirect to login
            localStorage.removeItem('snowpack_token');
            window.location.href = '/login';
            next(false);
            return;
          }
        }
      }
    } catch (error) {
      console.error('Error checking token expiration:', error);
      // If we can't parse the token, it's invalid
      localStorage.removeItem('snowpack_token');
      window.location.href = '/login';
      next(false);
      return;
    }
    
    next();
  } else {
    // Otherwise, proceed with navigation.
    next();
  }
});

export default router 