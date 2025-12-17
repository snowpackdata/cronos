import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import { jwtDecode } from 'jwt-decode'; // Corrected import
import { refreshToken } from '../api/apiUtils';

// Placeholder for LoginView, you'll need to create this component
const LoginView = { template: '<div>Admin Login Page - Implement Me</div>' };

interface DecodedToken {
  Email: string;
  IsStaff: boolean;
  UID: number;
  Role: string; // Add role field
  // Add other claims you expect, like exp, iat, etc.
  exp: number;
  iat: number;
}

// Define routes - Use dynamic imports to avoid typecheck errors
const routes: Array<RouteRecordRaw> = [
  {
    path: '/login',
    name: 'AdminLogin',
    component: LoginView,
    meta: { requiresAuth: false, title: 'Login' }
  },
  {
    path: '/',
    redirect: '/timesheet',
    meta: { requiresAuth: false } // Redirect itself doesn't require auth
  },
  {
    path: '/timesheet',
    name: 'timesheet',
    component: () => import('../views/timesheet/TimesheetView.vue'),
    meta: {
      title: 'Timesheet',
      requiresAuth: true
    }
  },
  {
    path: '/expenses',
    name: 'expenses',
    component: () => import('../views/expenses/ExpensesView.vue'),
    meta: {
      title: 'My Expenses',
      requiresAuth: true
    }
  },
  {
    path: '/timesheet-admin',
    name: 'timesheet-admin',
    component: () => import('../views/timesheet-admin/TimesheetAdminView.vue'),
    meta: {
      title: 'Timesheet Admin',
      requiresAuth: true
    }
  },
  {
    path: '/accounts-receivable',
    name: 'accounts-receivable',
    component: () => import('../views/invoices/AccountsReceivableView.vue'),
    meta: {
      title: 'Accounts Receivable',
      requiresAuth: true
    }
  },
  {
    path: '/accounts-payable',
    name: 'accounts-payable',
    component: () => import('../views/bills/AccountsPayableView.vue'),
    meta: {
      title: 'Accounts Payable',
      requiresAuth: true
    }
  },
  // Keep these routes for backward compatibility but redirect to the new routes
  {
    path: '/invoices',
    redirect: '/accounts-receivable',
    meta: { requiresAuth: false }
  },
  {
    path: '/bills',
    redirect: '/accounts-payable',
    meta: { requiresAuth: false }
  },
  {
    path: '/projects',
    name: 'projects',
    component: () => import('../views/projects/ProjectsView.vue'),
    meta: {
      title: 'Projects',
      requiresAuth: true
    }
  },
  {
    path: '/billing-codes',
    name: 'billing-codes',
    // @ts-ignore - Vue component type declaration
    component: () => import('../views/billing-codes/BillingCodesView.vue'),
    meta: {
      title: 'Billing Codes',
      requiresAuth: true
    }
  },
  {
    path: '/rates',
    name: 'rates',
    component: () => import('../views/rates/RatesView.vue'),
    meta: {
      title: 'Rates',
      requiresAuth: true
    }
  },
  {
    path: '/accounts',
    name: 'accounts',
    // @ts-ignore - Vue component type declaration
    component: () => import('../views/accounts/AccountsView.vue'),
    meta: {
      title: 'Accounts',
      requiresAuth: true
    }
  },
  {
    path: '/staff',
    name: 'team',
    component: () => import('../views/staff/StaffView.vue'),
    meta: {
      title: 'Team',
      requiresAuth: true
    }
  },
  {
    path: '/settings',
    name: 'settings',
    component: () => import('../views/settings/SettingsView.vue'),
    meta: {
      title: 'System Settings',
      requiresAuth: true
    }
  },
  {
    path: '/capacity',
    name: 'capacity',
    component: () => import('../views/capacity/CapacityView.vue'),
    meta: {
      title: 'Capacity Management',
      requiresAuth: true
    }
  },
  {
    path: '/expense-approvals',
    name: 'expense-approvals',
    component: () => import('../views/planning/ExpenseApprovalView.vue'),
    meta: {
      title: 'Expense Approvals',
      requiresAuth: true
    }
  },
  {
    path: '/recurring-entries',
    name: 'recurring-entries',
    component: () => import('../views/planning/RecurringEntriesView.vue'),
    meta: {
      title: 'Recurring Compensation',
      requiresAuth: true
    }
  },
  {
    path: '/expense-config',
    name: 'expense-config',
    component: () => import('../views/organization/ExpenseConfigView.vue'),
    meta: {
      title: 'Expense Configuration',
      requiresAuth: true
    }
  },
  {
    path: '/accounting',
    name: 'accounting',
    component: () => import('../views/accounting/AccountingView.vue'),
    meta: {
      title: 'General Ledger',
      requiresAuth: true
    }
  },
  {
    path: '/offline-journals',
    name: 'offline-journals',
    component: () => import('../views/accounting/OfflineJournalsView.vue'),
    meta: {
      title: 'Offline Journal Review',
      requiresAuth: true
    }
  },
  {
    path: '/chart-of-accounts',
    name: 'chart-of-accounts',
    component: () => import('../views/accounting/ChartOfAccountsView.vue'),
    meta: {
      title: 'Chart of Accounts',
      requiresAuth: true
    }
  }
];

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes
});

// Update page title and handle authentication
router.beforeEach(async (to, _from, next) => {
  document.title = `${to.meta.title || 'Admin'} | Cronos`;

  const token = localStorage.getItem('snowpack_token');
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth);
  let userIsStaff = false;
  let userRole = '';

  if (token) {
    try {
      const decodedToken = jwtDecode<DecodedToken>(token);
      // Check if token is expired (optional but good practice)
      if (decodedToken.exp * 1000 < Date.now()) {
        localStorage.removeItem('snowpack_token'); // Clear expired token
        window.location.href = '/login';
        next(false);
        return;
      }

      // Check if token is close to expiring (within 7 days) and refresh if needed
      const currentTime = Date.now();
      const sevenDaysInMs = 7 * 24 * 60 * 60 * 1000;

      if (decodedToken.exp * 1000 - currentTime < sevenDaysInMs) {
        const refreshSuccess = await refreshToken();
        if (!refreshSuccess) {
          // If refresh failed, redirect to login
          localStorage.removeItem('snowpack_token');
          window.location.href = '/login';
          next(false);
          return;
        }
      }

      // IMPORTANT: Confirm 'IsStaff' is the correct claim name from your JWT
      if (decodedToken && decodedToken.IsStaff === true) {
        userIsStaff = true;
        userRole = decodedToken.Role || '';
      }
    } catch (error) {
      console.error('Error decoding token:', error);
      localStorage.removeItem('snowpack_token'); // Clear invalid token
      // No need to redirect here yet, will be caught by requiresAuth check below
    }
  }

  if (requiresAuth) {
    if (token && userIsStaff) {
      // Check if user is staff (not admin) and trying to access restricted routes
      const staffAllowedPaths = ['/timesheet', '/expenses'];
      if (userRole === 'STAFF' && !staffAllowedPaths.includes(to.path)) {
        // Staff users can only access timesheet and expenses - redirect them
        next({ name: 'timesheet' });
        return;
      }
      // User has a token and is staff/admin, allow access
      next();
    } else {
      // No token, or token exists but user is not staff, or token is invalid
      localStorage.removeItem('snowpack_token'); // Ensure bad/non-staff token is cleared
      window.location.href = '/login';
      next(false);
    }
  } else if (to.name === 'AdminLogin' && token && userIsStaff) {
    // If authenticated staff user tries to access admin login page, redirect to default admin page
    next({ name: 'timesheet' });
  } else {
    // For non-auth routes, or if it's the login page and user is not staff/no token
    next();
  }
});

export default router;
