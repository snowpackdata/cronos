<template>
  <!-- Drawer Backdrop -->
  <div v-if="isOpen" class="fixed inset-0 z-50 overflow-hidden" @click="closeDrawer">
    <div class="absolute inset-0 overflow-hidden">
      <div class="pointer-events-none fixed inset-y-0 right-0 flex max-w-full pl-10 sm:pl-16">
        <!-- Drawer Content -->
        <div
          class="pointer-events-auto w-screen max-w-2xl transform transition duration-300 ease-in-out"
          :class="isOpen ? 'translate-x-0' : 'translate-x-full'"
          @click.stop
        >
          <div class="flex h-full flex-col overflow-y-auto bg-white shadow-xl">
            <!-- Header -->
            <div class="px-4 py-6 bg-gray-50 border-b border-gray-200 sm:px-6">
              <div class="flex items-start justify-between">
                <h2 class="text-lg font-semibold leading-6 text-gray-900">
                  {{ isEditing ? 'Edit Staff Member' : 'Add New Staff Member' }}
                </h2>
                <div class="ml-3 flex h-7 items-center">
                  <button
                    type="button"
                    class="rounded-md bg-gray-50 text-gray-400 hover:text-gray-600 focus:ring-2 focus:ring-sage"
                    @click="closeDrawer"
                  >
                    <span class="sr-only">Close panel</span>
                    <i class="fas fa-times h-6 w-6" aria-hidden="true" />
                  </button>
                </div>
              </div>
            </div>

            <!-- Form Content -->
            <div class="flex-1 px-4 py-6 sm:px-6">
              <form @submit.prevent="handleSave" class="space-y-6">

                <!-- Basic Information Section -->
                <div class="border-b border-gray-200 pb-6">
                  <h3 class="text-base font-semibold leading-6 text-gray-900 mb-4">Basic Information</h3>

                  <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                    <div>
                      <label for="first_name" class="block text-sm font-medium leading-6 text-gray-900">
                        First Name <span class="text-red-500">*</span>
                      </label>
                      <input
                        type="text"
                        name="first_name"
                        id="first_name"
                        v-model="formData.first_name"
                        required
                        class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm sm:leading-6"
                        placeholder="John"
                      />
                    </div>

                    <div>
                      <label for="last_name" class="block text-sm font-medium leading-6 text-gray-900">
                        Last Name <span class="text-red-500">*</span>
                      </label>
                      <input
                        type="text"
                        name="last_name"
                        id="last_name"
                        v-model="formData.last_name"
                        required
                        class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm sm:leading-6"
                        placeholder="Doe"
                      />
                    </div>
                  </div>

                  <div class="mt-4">
                    <label for="title" class="block text-sm font-medium leading-6 text-gray-900">
                      Job Title
                    </label>
                    <input
                      type="text"
                      name="title"
                      id="title"
                      v-model="formData.title"
                      class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm sm:leading-6"
                      placeholder="Senior Consultant"
                    />
                  </div>

                  <div class="mt-4">
                    <label for="headshot" class="block text-sm font-medium leading-6 text-gray-900">
                      Headshot Photo
                    </label>
                    <div class="mt-2 flex items-center gap-4">
                      <div v-if="headshotPreview || props.staffData?.headshot_asset?.url" class="flex-shrink-0">
                        <img
                          :src="headshotPreview || props.staffData?.headshot_asset?.url"
                          alt="Headshot preview"
                          class="h-16 w-16 rounded-full object-cover border-2 border-gray-200"
                        />
                      </div>
                      <div v-else class="flex-shrink-0 h-16 w-16 rounded-full bg-sage-pale flex items-center justify-center border-2 border-gray-200">
                        <span class="text-sage font-medium text-lg">
                          {{ getInitials() }}
                        </span>
                      </div>
                      <div class="flex-1">
                        <input
                          type="file"
                          id="headshot"
                          ref="headshotInput"
                          accept="image/*"
                          @change="handleHeadshotChange"
                          class="block w-full text-sm text-gray-900 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-sage file:text-white hover:file:bg-sage-dark"
                        />
                        <p class="mt-1 text-xs text-gray-500">PNG, JPG, GIF up to 10MB</p>
                      </div>
                    </div>
                  </div>

                  <div class="mt-4">
                    <label for="email" class="block text-sm font-medium leading-6 text-gray-900">
                      Email Address <span class="text-red-500">*</span>
                    </label>
                    <input
                      type="email"
                      name="email"
                      id="email"
                      v-model="formData.email"
                      required
                      :disabled="false"
                      class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm sm:leading-6 disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
                      placeholder="john.doe@example.com"
                    />
                    <p class="mt-1 text-sm text-gray-500">
                      {{ isEditing ? 'Updates the user account email address' : 'User will log in with this email address' }}
                    </p>
                  </div>

                  <div v-if="!isEditing" class="mt-4">
                    <label for="user_role" class="block text-sm font-medium leading-6 text-gray-900">
                      User Role <span class="text-red-500">*</span>
                    </label>
                    <select
                      name="user_role"
                      id="user_role"
                      v-model="formData.user_role"
                      required
                      class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm sm:leading-6"
                    >
                      <option value="STAFF">Staff (Can log time and expenses)</option>
                      <option value="ADMIN">Admin (Full access to all features)</option>
                    </select>
                    <p class="mt-1 text-sm text-gray-500">System access level for this user</p>
                  </div>

                  <div v-if="!isEditing" class="mt-4">
                    <label for="password" class="block text-sm font-medium leading-6 text-gray-900">
                      Initial Password <span class="text-sm text-gray-500">(Optional)</span>
                    </label>
                    <input
                      type="password"
                      name="password"
                      id="password"
                      v-model="formData.password"
                      class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm sm:leading-6"
                      placeholder="Leave blank for default: ChangeMe123!"
                    />
                    <p class="mt-1 text-sm text-gray-500">If not provided, default password is "ChangeMe123!" - user should change on first login</p>
                  </div>

                  <div class="mt-4">
                    <label for="employment_status" class="block text-sm font-medium leading-6 text-gray-900">
                      Employment Status
                    </label>
                    <select
                      name="employment_status"
                      id="employment_status"
                      v-model="formData.employment_status"
                      class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm sm:leading-6"
                    >
                      <option :value="EmploymentStatus.ACTIVE">Active (Currently Working)</option>
                      <option :value="EmploymentStatus.INACTIVE">Inactive (Available Contractor/Consultant)</option>
                      <option :value="EmploymentStatus.TERMINATED">Terminated (Former Employee)</option>
                    </select>
                    <p class="mt-1 text-sm text-gray-500">Current employment relationship status</p>
                  </div>

                  <div class="mt-4">
                    <label for="is_active" class="flex items-center">
                      <input
                        type="checkbox"
                        name="is_active"
                        id="is_active"
                        v-model="formData.is_active"
                        class="h-4 w-4 rounded border-gray-300 text-sage focus:ring-sage"
                      />
                      <span class="ml-2 text-sm font-medium text-gray-900">Active in System</span>
                    </label>
                    <p class="mt-1 text-sm text-gray-500">Can access the system and log time entries</p>
                  </div>

                  <div class="mt-4">
                    <label for="is_owner" class="flex items-center">
                      <input
                        type="checkbox"
                        name="is_owner"
                        id="is_owner"
                        v-model="formData.is_owner"
                        class="h-4 w-4 rounded border-gray-300 text-sage focus:ring-sage"
                      />
                      <span class="ml-2 text-sm font-medium text-gray-900">Company Owner/Partner</span>
                    </label>
                    <p class="mt-1 text-sm text-gray-500">Payouts will be classified as equity distributions instead of payroll expenses</p>
                  </div>
                </div>

                <!-- Employment Details Section -->
                <div class="border-b border-gray-200 pb-6">
                  <h3 class="text-base font-semibold leading-6 text-gray-900 mb-4">Employment Details</h3>

                  <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                    <div>
                      <label for="start_date" class="block text-sm font-medium leading-6 text-gray-900">
                        Start Date
                      </label>
                      <input
                        type="date"
                        name="start_date"
                        id="start_date"
                        v-model="formData.start_date"
                        class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm sm:leading-6"
                      />
                    </div>

                    <div v-if="formData.employment_status === EmploymentStatus.TERMINATED">
                      <label for="end_date" class="block text-sm font-medium leading-6 text-gray-900">
                        End Date
                      </label>
                      <input
                        type="date"
                        name="end_date"
                        id="end_date"
                        v-model="formData.end_date"
                        class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm sm:leading-6"
                      />
                      <p class="mt-1 text-sm text-gray-500">Date when employment ended</p>
                    </div>
                  </div>

                  <div class="mt-4">
                    <label for="capacity_weekly" class="block text-sm font-medium leading-6 text-gray-900">
                      Weekly Capacity (Hours)
                    </label>
                    <input
                      type="number"
                      name="capacity_weekly"
                      id="capacity_weekly"
                      v-model="formData.capacity_weekly"
                      min="0"
                      max="80"
                      class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm sm:leading-6"
                      placeholder="40"
                    />
                    <p class="mt-1 text-sm text-gray-500">Available hours per week for assignments</p>
                  </div>
                </div>

                <!-- Compensation Section -->
                <div class="border-b border-gray-200 pb-6">
                  <h3 class="text-base font-semibold leading-6 text-gray-900 mb-4">Compensation</h3>

                  <div class="mb-4">
                    <label for="compensation_type" class="block text-sm font-medium leading-6 text-gray-900">
                      Compensation Model
                    </label>
                    <select
                      name="compensation_type"
                      id="compensation_type"
                      v-model="formData.compensation_type"
                      class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm sm:leading-6"
                    >
                      <option :value="CompensationType.FULLY_VARIABLE">Fully Variable ({{ CompensationTypeDescriptions[CompensationType.FULLY_VARIABLE] }})</option>
                      <option :value="CompensationType.SALARIED">Salaried ({{ CompensationTypeDescriptions[CompensationType.SALARIED] }})</option>
                      <option :value="CompensationType.BASE_PLUS_VARIABLE">Base + Variable ({{ CompensationTypeDescriptions[CompensationType.BASE_PLUS_VARIABLE] }})</option>
                    </select>
                    <p class="mt-1 text-sm text-gray-500">How this person's compensation is structured</p>
                  </div>

                  <!-- Compensation inputs based on type -->
                  <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                    <div v-if="formData.compensation_type === CompensationType.SALARIED">
                      <label for="salary_annualized" class="block text-sm font-medium leading-6 text-gray-900">
                        Annual Salary ($)
                      </label>
                      <input
                        type="number"
                        name="salary_annualized"
                        id="salary_annualized"
                        v-model="formData.salary_annualized"
                        min="0"
                        class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm sm:leading-6"
                        placeholder="75000"
                      />
                    </div>

                    <div v-if="formData.compensation_type === CompensationType.BASE_PLUS_VARIABLE">
                      <label for="base_salary" class="block text-sm font-medium leading-6 text-gray-900">
                        Base Salary ($)
                      </label>
                      <input
                        type="number"
                        name="base_salary"
                        id="base_salary"
                        v-model="formData.base_salary"
                        min="0"
                        class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm sm:leading-6"
                        placeholder="50000"
                      />
                      <p class="mt-1 text-sm text-gray-500">Base annual salary before variable bonuses</p>
                    </div>

                    <!-- Hourly Rate Structure - for Fully Variable and Base + Variable -->
                    <div v-if="formData.compensation_type === CompensationType.FULLY_VARIABLE || formData.compensation_type === CompensationType.BASE_PLUS_VARIABLE">
                      <div class="mb-4">
                        <label class="block text-sm font-medium leading-6 text-gray-900 mb-2">
                          Hourly Rate Structure
                          <span v-if="formData.compensation_type === CompensationType.BASE_PLUS_VARIABLE" class="text-sm text-gray-500">(for billable component)</span>
                        </label>
                        <div class="space-y-2">
                          <label class="flex items-center">
                            <input
                              type="radio"
                              name="hourly_rate_type"
                              :value="false"
                              v-model="formData.is_fixed_hourly"
                              class="h-4 w-4 border-gray-300 text-sage focus:ring-sage"
                            />
                            <span class="ml-2 text-sm text-gray-700">Variable by Project (rate changes per project)</span>
                          </label>
                          <label class="flex items-center">
                            <input
                              type="radio"
                              name="hourly_rate_type"
                              :value="true"
                              v-model="formData.is_fixed_hourly"
                              class="h-4 w-4 border-gray-300 text-sage focus:ring-sage"
                            />
                            <span class="ml-2 text-sm text-gray-700">Fixed Hourly Rate (same rate for all projects)</span>
                          </label>
                        </div>
                      </div>

                      <div v-if="formData.is_fixed_hourly">
                        <label for="hourly_rate" class="block text-sm font-medium leading-6 text-gray-900">
                          <span v-if="formData.compensation_type === CompensationType.FULLY_VARIABLE">Fixed Hourly Rate ($)</span>
                          <span v-else>Fixed Billable Rate ($)</span>
                        </label>
                        <input
                          type="number"
                          name="hourly_rate"
                          id="hourly_rate"
                          v-model="hourlyRateDisplay"
                          min="0"
                          step="0.01"
                          class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm sm:leading-6"
                          placeholder="50.00"
                        />
                        <p class="mt-1 text-sm text-gray-500">
                          <span v-if="formData.compensation_type === CompensationType.FULLY_VARIABLE">Fixed rate used for all projects</span>
                          <span v-else>Fixed billable rate used for all projects (in addition to base salary)</span>
                        </p>
                      </div>

                      <div v-if="!formData.is_fixed_hourly">
                        <p class="text-sm text-gray-500">
                          <span v-if="formData.compensation_type === CompensationType.FULLY_VARIABLE">Hourly rates will be set per project/billing code</span>
                          <span v-else>Billable rates will be set per project/billing code (in addition to base salary)</span>
                        </p>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Advanced Settings Section -->
                <div>
                  <h3 class="text-base font-semibold leading-6 text-gray-900 mb-4">Advanced Settings</h3>

                  <div>
                    <label for="entry_pay_eligible_state" class="block text-sm font-medium leading-6 text-gray-900">
                      Pay Eligible State
                    </label>
                    <select
                      name="entry_pay_eligible_state"
                      id="entry_pay_eligible_state"
                      v-model="formData.entry_pay_eligible_state"
                      class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-sage sm:text-sm sm:leading-6"
                    >
                      <option value="">Select when employee gets paid...</option>
                      <option value="ENTRY_STATE_DRAFT">When entries are in Draft</option>
                      <option value="ENTRY_STATE_APPROVED">When entries are Approved</option>
                      <option value="ENTRY_STATE_SENT">When entries are Sent to Client</option>
                      <option value="ENTRY_STATE_PAID">When entries are Paid by Client</option>
                    </select>
                    <p class="mt-1 text-sm text-gray-500">
                      Defines at what stage of the billing process this employee becomes eligible for payment
                    </p>
                  </div>
                </div>
              </form>
            </div>

            <!-- Footer with action buttons -->
            <div class="border-t border-gray-200 px-4 py-4 sm:px-6">
              <div class="flex justify-between">
                <div>
                  <button
                    v-if="isEditing"
                    type="button"
                    @click="handleDelete"
                    class="rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-red-600"
                  >
                    Delete Staff Member
                  </button>
                </div>
                <div class="flex gap-3">
                  <button
                    type="button"
                    @click="closeDrawer"
                    class="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    @click="handleSave"
                    :disabled="!isFormValid"
                    class="rounded-md bg-sage px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-sage-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sage disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {{ isEditing ? 'Update' : 'Create' }} Staff Member
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import type { Staff } from '../../types/Project';
import { EmploymentStatus, CompensationType, CompensationTypeDescriptions } from '../../types/constants';

// Props
interface Props {
  isOpen: boolean;
  staffData?: Staff | null;
}

const props = withDefaults(defineProps<Props>(), {
  staffData: null
});

// Emits
const emit = defineEmits<{
  close: [];
  save: [staffData: Staff];
  delete: [staffId: number];
}>();

// Form data
const formData = ref<Partial<Staff & { is_owner: boolean; user_role?: string; password?: string }>>({
  first_name: '',
  last_name: '',
  title: '',
  email: '',
  is_active: true,
  is_owner: false,
  employment_status: EmploymentStatus.ACTIVE,
  start_date: '',
  end_date: '',
  capacity_weekly: 40,
  compensation_type: CompensationType.FULLY_VARIABLE,
  is_salaried: false,
  salary_annualized: 0,
  base_salary: 0,
  is_variable_hourly: true,
  is_fixed_hourly: false,
  hourly_rate: 0,
  entry_pay_eligible_state: 'ENTRY_STATE_PAID',
  user_role: 'STAFF',  // Default to STAFF role
  password: ''  // Optional initial password
});

// Headshot handling
const headshotInput = ref<HTMLInputElement | null>(null);
const headshotFile = ref<File | null>(null);
const headshotPreview = ref<string | null>(null);

const handleHeadshotChange = (event: Event) => {
  const target = event.target as HTMLInputElement;
  if (target.files && target.files[0]) {
    headshotFile.value = target.files[0];
    // Create preview URL
    const reader = new FileReader();
    reader.onload = (e) => {
      headshotPreview.value = e.target?.result as string;
    };
    reader.readAsDataURL(target.files[0]);
  }
};

const getInitials = () => {
  const first = formData.value.first_name?.[0] || '';
  const last = formData.value.last_name?.[0] || '';
  return (first + last).toUpperCase();
};

// Hourly rate display (no conversion needed - will be converted to cents in backend)
const hourlyRateDisplay = computed({
  get: () => formData.value.hourly_rate || 0,
  set: (value: number) => {
    formData.value.hourly_rate = Math.round(value);
  }
});

// Computed properties
const isEditing = computed(() => !!props.staffData?.ID);

const isFormValid = computed(() => {
  return !!(formData.value.first_name?.trim() && formData.value.last_name?.trim());
});

// Watch for employment status changes to auto-set end date
watch(() => formData.value.employment_status, (newStatus, oldStatus) => {
  if (newStatus === EmploymentStatus.TERMINATED && oldStatus !== EmploymentStatus.TERMINATED) {
    // When changing to terminated, set end date to today if not already set
    if (!formData.value.end_date) {
      formData.value.end_date = new Date().toISOString().split('T')[0];
    }
  } else if (newStatus !== EmploymentStatus.TERMINATED && oldStatus === EmploymentStatus.TERMINATED) {
    // When changing away from terminated, clear end date
    formData.value.end_date = '';
  }
});

// Watch for changes in staffData prop
watch(() => props.staffData, (newStaffData) => {
  if (newStaffData) {
    // Convert dates to YYYY-MM-DD format for date inputs
    const formatDateForInput = (dateValue: string | Date | undefined) => {
      if (!dateValue) return '';
      const date = new Date(dateValue);
      // Check if date is invalid or is a zero date (0001-01-01)
      if (isNaN(date.getTime()) || date.getFullYear() <= 1) {
        return '';
      }
      return date.toISOString().split('T')[0];
    };

    // Format end date with smart defaults for terminated employees
    let formattedEndDate = formatDateForInput(newStaffData.end_date);
    if (newStaffData.employment_status === EmploymentStatus.TERMINATED && !formattedEndDate) {
      // For terminated employees with no valid end date, default to today
      formattedEndDate = new Date().toISOString().split('T')[0];
    }

    formData.value = {
      ...newStaffData,
      start_date: formatDateForInput(newStaffData.start_date),
      end_date: formattedEndDate,
      // Map email from user relationship (since that's where the actual email is stored)
      email: newStaffData.user?.email || newStaffData.email || '',
      // Map salary fields properly based on compensation type
      base_salary: newStaffData.compensation_type === CompensationType.BASE_PLUS_VARIABLE
        ? (newStaffData.base_salary || newStaffData.salary_annualized || 0)
        : 0,
      salary_annualized: newStaffData.compensation_type === CompensationType.SALARIED
        ? (newStaffData.salary_annualized || 0)
        : 0
    };
  } else {
    // Reset form for new staff member
    formData.value = {
      first_name: '',
      last_name: '',
      title: '',
      email: '',
      is_active: true,
      is_owner: false,
      employment_status: EmploymentStatus.ACTIVE,
      start_date: new Date().toISOString().split('T')[0], // Today's date
      end_date: '',
      capacity_weekly: 40,
      compensation_type: CompensationType.FULLY_VARIABLE,
      is_salaried: false,
      salary_annualized: 0,
      base_salary: 0,
      is_variable_hourly: true,
      is_fixed_hourly: false,
      hourly_rate: 0,
      entry_pay_eligible_state: 'ENTRY_STATE_PAID'
    };
  }
}, { immediate: true });

// Methods
const closeDrawer = () => {
  emit('close');
};

const handleSave = () => {
  if (!isFormValid.value) return;

  const staffToSave = {
    ...formData.value,
    // Include headshot file if selected
    headshot: headshotFile.value,
    // Ensure we're sending the right data types
    capacity_weekly: Number(formData.value.capacity_weekly) || 0,
    salary_annualized: Number(formData.value.salary_annualized) || 0,
    base_salary: Number(formData.value.base_salary) || 0,
    hourly_rate: Number(formData.value.hourly_rate) || 0,
    // Map compensation_type to backend fields for compatibility
    is_salaried: formData.value.compensation_type === CompensationType.SALARIED,
    // is_variable_hourly = true means hourly rate can change per project
    // is_fixed_hourly = true means they have a fixed hourly rate
    // Both Fully Variable and Base + Variable can have variable or fixed rates
    is_variable_hourly: (formData.value.compensation_type === CompensationType.FULLY_VARIABLE || formData.value.compensation_type === CompensationType.BASE_PLUS_VARIABLE) && !formData.value.is_fixed_hourly,
    is_fixed_hourly: (formData.value.compensation_type === CompensationType.FULLY_VARIABLE || formData.value.compensation_type === CompensationType.BASE_PLUS_VARIABLE) && formData.value.is_fixed_hourly
  } as Staff;

  emit('save', staffToSave);
};

const handleDelete = () => {
  if (props.staffData?.ID) {
    emit('delete', props.staffData.ID);
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
.focus\:ring-sage {
  --tw-ring-color: #58837e;
}
.focus-visible\:outline-sage {
  outline-color: #58837e;
}
.text-sage {
  color: #58837e;
}
</style>
