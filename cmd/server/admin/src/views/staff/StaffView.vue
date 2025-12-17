<template>
  <div class="px-4 py-6 sm:px-6 lg:px-8">
    <div class="sm:flex sm:items-center">
      <div class="sm:flex-auto">
        <h1 class="text-base font-semibold leading-6 text-gray-900">Team</h1>
        <p class="mt-2 text-sm text-gray-700">
          Manage your team members, their roles, capacity, and employment status.
        </p>
      </div>
      <div class="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
        <button
          type="button"
          @click="openStaffDrawer()"
          class="block rounded-md bg-sage px-3 py-2 text-center text-sm font-semibold text-white shadow-sm hover:bg-sage-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sage"
        >
          Add New Staff Member
        </button>
      </div>
    </div>

    <!-- Filter Controls -->
    <div class="mt-6 flex gap-4">
      <div class="flex items-center gap-2">
        <label for="employment-status-filter" class="text-sm font-medium text-gray-700">Filter by Status:</label>
        <select
          id="employment-status-filter"
          v-model="statusFilter"
          class="rounded-md border-gray-300 py-1.5 text-sm focus:border-sage focus:ring-sage"
        >
          <option value="all">All Staff</option>
          <option :value="EmploymentStatus.ACTIVE">Active</option>
          <option :value="EmploymentStatus.INACTIVE">Inactive</option>
          <option :value="EmploymentStatus.TERMINATED">Terminated</option>
        </select>
      </div>
    </div>

    <!-- Staff Table -->
    <div class="mt-8 flow-root">
      <div class="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
        <div class="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
          <div class="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
            <table class="min-w-full divide-y divide-gray-300">
              <thead class="bg-gray-50">
                <tr>
                  <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">Name</th>
                  <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Status</th>
                  <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Title</th>
                  <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Role</th>
                  <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Capacity</th>
                  <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Employment</th>
                  <th scope="col" class="relative py-3.5 pl-3 pr-4 sm:pr-6">
                    <span class="sr-only">Edit</span>
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200 bg-white">
                <tr v-for="staff in filteredStaff" :key="staff.ID" class="hover:bg-gray-50">
                  <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm sm:pl-6">
                    <div class="flex items-center">
                      <StaffAvatar :employee="staff" size="sm" />
                      <div class="ml-3">
                        <div class="font-medium text-gray-900">{{ staff.first_name }} {{ staff.last_name }}</div>
                        <div v-if="staff.start_date" class="text-gray-500">Started {{ formatDate(staff.start_date) }}</div>
                      </div>
                    </div>
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm">
                    <span :class="[
                      getEmploymentStatusClass(staff.employment_status || 'in-seat'),
                      'inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset'
                    ]">
                      {{ formatEmploymentStatus(staff.employment_status || 'in-seat') }}
                    </span>
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                    {{ staff.title || '—' }}
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm">
                    <span :class="[
                      staff.is_owner ? 'bg-purple-50 text-purple-700 ring-purple-600/20' : getUserRoleClass(staff.user?.role || 'STAFF'),
                      'inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset'
                    ]">
                      {{ staff.is_owner ? 'Partner' : formatUserRole(staff.user?.role || 'STAFF') }}
                    </span>
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                    <div class="flex items-center">
                      <i class="fas fa-clock mr-1.5 text-gray-400"></i>
                      {{ staff.capacity_weekly || 0 }}h/week
                    </div>
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm">
                    <span :class="[
                      getCompensationTypeClass(staff.compensation_type || mapSalariedToCompensationType(staff.is_salaried)),
                      'inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset'
                    ]">
                      <i :class="['fas mr-1.5', getCompensationTypeIcon(staff.compensation_type || mapSalariedToCompensationType(staff.is_salaried))]"></i>
                      {{ formatCompensationType(staff.compensation_type || mapSalariedToCompensationType(staff.is_salaried)) }}
                    </span>
                  </td>
                  <td class="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                    <button
                      @click="openStaffDrawer(staff)"
                      class="text-sage hover:text-sage-dark"
                    >
                      Edit<span class="sr-only">, {{ staff.first_name }} {{ staff.last_name }}</span>
                    </button>
                  </td>
                </tr>

                <!-- Empty State -->
                <tr v-if="filteredStaff.length === 0">
                  <td colspan="7" class="py-12 text-center text-sm text-gray-500">
                    <div class="flex flex-col items-center">
                      <i class="fas fa-users text-4xl text-gray-300 mb-4"></i>
                      <p class="text-lg font-medium text-gray-900 mb-2">No staff members found</p>
                      <p class="text-gray-500 mb-4">
                        {{ statusFilter === 'all'
                          ? 'Click "Add New Staff Member" to add one'
                          : `No ${formatEmploymentStatus(statusFilter)} staff members found`
                        }}
                      </p>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <!-- Staff Drawer -->
    <StaffDrawer
      :is-open="isStaffDrawerOpen"
      :staff-data="selectedStaff"
      @close="closeStaffDrawer"
      @save="saveStaff"
      @delete="handleDeleteFromDrawer"
    />

    <!-- Delete Confirmation Modal -->
    <ConfirmationModal
      :show="showDeleteModal"
      title="Delete Staff Member"
      message="Are you sure you want to delete this staff member? This action cannot be undone."
      @confirm="deleteStaff"
      @cancel="showDeleteModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { fetchStaff, createStaff, updateStaff, deleteStaff as deleteStaffAPI } from '../../api/staff';
import StaffDrawer from '../../components/staff/StaffDrawer.vue';
import ConfirmationModal from '../../components/ConfirmationModal.vue';
import StaffAvatar from '../../components/StaffAvatar.vue';
import type { Staff } from '../../types/Project';
import { EmploymentStatus, CompensationType, EmploymentStatusLabels, CompensationTypeLabels } from '../../types/constants';

// State
const staff = ref<Staff[]>([]);
const isStaffDrawerOpen = ref(false);
const selectedStaff = ref<Staff | null>(null);
const showDeleteModal = ref(false);
const staffToDelete = ref<Staff | null>(null);
const statusFilter = ref<'all' | 'EMPLOYMENT_STATUS_ACTIVE' | 'EMPLOYMENT_STATUS_INACTIVE' | 'EMPLOYMENT_STATUS_TERMINATED'>('all');

// Computed
const filteredStaff = computed(() => {
  if (statusFilter.value === 'all') {
    return staff.value;
  }
  return staff.value.filter(member =>
    (member.employment_status || EmploymentStatus.ACTIVE) === statusFilter.value
  );
});

// Load staff data
const loadStaff = async () => {
  try {
    const fetchedStaff = await fetchStaff();
    // Use the backend employment_status field directly, with fallback to mapping for backward compatibility
    const mappedStaff = fetchedStaff?.map(member => ({
      ...member,
      employment_status: member.employment_status || mapIsActiveToEmploymentStatus(member.is_active),
      compensation_type: member.compensation_type || mapSalariedToCompensationType(member.is_salaried)
    })) || [];
    staff.value = mappedStaff;
  } catch (error) {
    console.error('Failed to load staff:', error);
    staff.value = [];
  }
};

// Helper function to map backend is_active to frontend employment_status
const mapIsActiveToEmploymentStatus = (isActive?: boolean) => {
  // For now, we'll default to 'EMPLOYMENT_STATUS_ACTIVE' for active and 'EMPLOYMENT_STATUS_TERMINATED' for inactive
  // Later when backend supports employment_status, this mapping won't be needed
  return isActive ? EmploymentStatus.ACTIVE : EmploymentStatus.TERMINATED;
};

onMounted(loadStaff);

// Helper functions
const formatUserRole = (role: string) => {
  const roles: Record<string, string> = {
    'ADMIN': 'Administrator',
    'STAFF': 'Staff',
    'CLIENT': 'Client'
  };
  return roles[role] || role;
};

const getUserRoleClass = (role: string) => {
  const classes: Record<string, string> = {
    'ADMIN': 'bg-red-50 text-red-700 ring-red-600/20',
    'STAFF': 'bg-blue-50 text-blue-700 ring-blue-600/20',
    'CLIENT': 'bg-green-50 text-green-700 ring-green-600/20'
  };
  return classes[role] || 'bg-gray-50 text-gray-700 ring-gray-600/20';
};

const formatDate = (dateString: string | Date) => {
  if (!dateString) return 'Not set';
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  });
};

// Removed formatCurrency as it's no longer used in the table view

const formatEmploymentStatus = (status: string) => {
  return EmploymentStatusLabels[status as keyof typeof EmploymentStatusLabels] || status;
};

const getEmploymentStatusClass = (status: string) => {
  const classes: Record<string, string> = {
    [EmploymentStatus.ACTIVE]: 'bg-green-50 text-green-700 ring-green-600/20',
    [EmploymentStatus.INACTIVE]: 'bg-yellow-50 text-yellow-700 ring-yellow-600/20',
    [EmploymentStatus.TERMINATED]: 'bg-red-50 text-red-700 ring-red-600/20'
  };
  return classes[status] || 'bg-gray-50 text-gray-700 ring-gray-600/20';
};

// Helper function to map old salaried boolean to new compensation types (for backward compatibility)
const mapSalariedToCompensationType = (isSalaried?: boolean) => {
  return isSalaried ? CompensationType.SALARIED : CompensationType.FULLY_VARIABLE;
};

const formatCompensationType = (type: string) => {
  return CompensationTypeLabels[type as keyof typeof CompensationTypeLabels] || type;
};

const getCompensationTypeClass = (type: string) => {
  const classes: Record<string, string> = {
    [CompensationType.FULLY_VARIABLE]: 'bg-purple-50 text-purple-700 ring-purple-600/20',
    [CompensationType.SALARIED]: 'bg-blue-50 text-blue-700 ring-blue-600/20',
    [CompensationType.BASE_PLUS_VARIABLE]: 'bg-indigo-50 text-indigo-700 ring-indigo-600/20'
  };
  return classes[type] || 'bg-gray-50 text-gray-700 ring-gray-600/20';
};

const getCompensationTypeIcon = (type: string) => {
  const icons: Record<string, string> = {
    [CompensationType.FULLY_VARIABLE]: 'fa-chart-line',
    [CompensationType.SALARIED]: 'fa-calendar-alt',
    [CompensationType.BASE_PLUS_VARIABLE]: 'fa-plus-circle'
  };
  return icons[type] || 'fa-dollar-sign';
};

// Drawer functions
const openStaffDrawer = (staffMember: Staff | null = null) => {
  selectedStaff.value = staffMember;
  isStaffDrawerOpen.value = true;
};

const closeStaffDrawer = () => {
  isStaffDrawerOpen.value = false;
  selectedStaff.value = null;
};

// Save staff
const saveStaff = async (staffData: Staff) => {
  try {
    if (selectedStaff.value && selectedStaff.value.ID) {
      await updateStaff(selectedStaff.value.ID, staffData);
    } else {
      await createStaff(staffData);
    }
    await loadStaff();
    closeStaffDrawer();
  } catch (error) {
    console.error('Failed to save staff member:', error);
    alert('Failed to save staff member. Please try again.');
  }
};

// Delete staff
const deleteStaff = async () => {
  if (staffToDelete.value && staffToDelete.value.ID) {
    try {
      await deleteStaffAPI(staffToDelete.value.ID);
      await loadStaff();
      showDeleteModal.value = false;
      staffToDelete.value = null;
      // Close the staff drawer since the staff member no longer exists
      closeStaffDrawer();
    } catch (error) {
      console.error('Failed to delete staff member:', error);
    }
  }
};

const handleDeleteFromDrawer = (staffId: number) => {
  const staffMember = staff.value.find(s => s.ID === staffId);
  if (staffMember) {
    staffToDelete.value = staffMember;
    showDeleteModal.value = true;
  }
};
</script>

<style scoped>
.bg-sage {
  background-color: #58837e;
}
.bg-sage-dark {
  background-color: #476b67;
}
.text-sage {
  color: #58837e;
}
.text-sage-dark {
  color: #476b67;
}
.focus-visible\:outline-sage {
  outline-color: #58837e;
}
.focus\:border-sage {
  border-color: #58837e;
}
.focus\:ring-sage {
  ring-color: #58837e;
}
</style>
