<template>
  <TransitionRoot as="template" :show="isOpen">
    <Dialog class="relative z-10" @close="handleClose">
      <div class="fixed inset-0" />

      <div class="fixed inset-0 overflow-hidden">
        <div class="absolute inset-0 overflow-hidden">
          <div class="pointer-events-none fixed inset-y-0 right-0 flex max-w-full pl-10 sm:pl-16">
            <TransitionChild as="template" enter="transform transition ease-in-out duration-500 sm:duration-700" enter-from="translate-x-full" enter-to="translate-x-0" leave="transform transition ease-in-out duration-500 sm:duration-700" leave-from="translate-x-0" leave-to="translate-x-full">
              <DialogPanel class="pointer-events-auto w-screen max-w-2xl">
                <form class="flex h-full flex-col overflow-y-scroll bg-white shadow-xl" @submit.prevent="handleSubmit">
                  <div class="flex-1">
                    <!-- Header -->
                    <div class="bg-sage-dark px-4 py-6 sm:px-6" :class="{ 'bg-blue-700': isEditing }">
                      <div class="flex items-start justify-between space-x-3">
                        <div class="space-y-1">
                          <DialogTitle class="text-base font-semibold text-white">{{ isEditing ? 'Edit Project' : 'New Project' }}</DialogTitle>
                          <p class="text-sm text-gray-100">
                            {{ isEditing ? 'Update the project details below.' : 'Get started by filling in the information below to create your new project.' }}
                          </p>
                        </div>
                        <div class="flex h-7 items-center">
                          <button type="button" class="relative text-white hover:text-gray-200" @click="handleClose">
                            <span class="absolute -inset-2.5" />
                            <span class="sr-only">Close panel</span>
                            <XMarkIcon class="h-6 w-6" aria-hidden="true" />
                          </button>
                        </div>
                      </div>
                    </div>

                    <!-- Divider container -->
                    <div class="space-y-6 py-6 sm:space-y-0 sm:divide-y sm:divide-gray-200 sm:py-0">
                      <!-- Project name -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="project-name" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Project name</label>
                        </div>
                        <div class="sm:col-span-2">
                          <input 
                            type="text" 
                            name="project-name" 
                            id="project-name" 
                            v-model="project.name"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6" 
                          />
                        </div>
                      </div>

                      <!-- Account selection -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="project-account" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Account</label>
                        </div>
                        <div class="sm:col-span-2">
                          <select 
                            id="project-account" 
                            name="project-account" 
                            v-model="project.account_id"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6"
                          >
                            <option value="">Select an account</option>
                            <option v-for="acc in accounts" :key="acc.ID" :value="acc.ID">
                              {{ acc.name }}
                            </option>
                          </select>
                        </div>
                      </div>

                      <!-- Project Description -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="project-description" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Description</label>
                        </div>
                        <div class="sm:col-span-2">
                          <textarea
                            id="project-description"
                            name="project-description"
                            rows="4"
                            v-model="project.description"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                          ></textarea>
                          <p class="mt-2 text-xs text-gray-500">A brief description for this project.</p>
                        </div>
                      </div>

                      <!-- Project type -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="project-type" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Project Type</label>
                        </div>
                        <div class="sm:col-span-2">
                          <select 
                            id="project-type"
                            name="project-type"
                            v-model="project.project_type"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6"
                          >
                            <option value="PROJECT_TYPE_NEW">New Project</option>
                            <option value="PROJECT_TYPE_EXISTING">Existing Project</option>
                          </select>
                        </div>
                      </div>

                      <!-- Start date -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="project-start-date" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Start Date</label>
                        </div>
                        <div class="sm:col-span-2">
                          <input 
                            type="date" 
                            name="project-start-date" 
                            id="project-start-date" 
                            v-model="project.active_start"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6" 
                          />
                        </div>
                      </div>

                      <!-- End date -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="project-end-date" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">End Date</label>
                        </div>
                        <div class="sm:col-span-2">
                          <input 
                            type="date" 
                            name="project-end-date" 
                            id="project-end-date" 
                            v-model="project.active_end"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6" 
                          />
                        </div>
                      </div>

                      <!-- Internal project -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="project-internal" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Internal Project</label>
                        </div>
                        <div class="sm:col-span-2 flex items-center">
                          <input 
                            type="checkbox" 
                            name="project-internal" 
                            id="project-internal" 
                            v-model="project.internal"
                            class="h-4 w-4 rounded border-gray-300 text-sage focus:ring-sage" 
                          />
                          <label for="project-internal" class="ml-2 text-sm text-gray-600">
                            Mark this project as internal
                          </label>
                        </div>
                      </div>

                      <!-- Billing Frequency -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="billing-frequency" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Billing Frequency</label>
                        </div>
                        <div class="sm:col-span-2">
                          <select 
                            id="billing-frequency"
                            name="billing-frequency"
                            v-model="project.billing_frequency"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6"
                          >
                            <option value="BILLING_TYPE_MONTHLY">Monthly</option>
                            <option value="BILLING_TYPE_PROJECT">Project</option>
                            <option value="BILLING_TYPE_BIWEEKLY">Bi-Weekly</option>
                            <option value="BILLING_TYPE_WEEKLY">Weekly</option>
                            <option value="BILLING_TYPE_BIMONTHLY">Bi-Monthly</option>
                          </select>
                        </div>
                      </div>

                      <!-- Budget Hours -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="budget-hours" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Budget Hours</label>
                        </div>
                        <div class="sm:col-span-2">
                          <input 
                            type="number" 
                            name="budget-hours" 
                            id="budget-hours" 
                            v-model="project.budget_hours"
                            min="0"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6" 
                          />
                        </div>
                      </div>

                      <!-- Budget Dollars -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="budget-dollars" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Budget Dollars</label>
                        </div>
                        <div class="sm:col-span-2">
                          <input 
                            type="number" 
                            name="budget-dollars" 
                            id="budget-dollars" 
                            v-model="project.budget_dollars"
                            min="0"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6" 
                          />
                        </div>
                      </div>

                      <!-- Budget Cap Hours -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="budget-cap-hours" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Budget Cap Hours</label>
                        </div>
                        <div class="sm:col-span-2">
                          <input 
                            type="number" 
                            name="budget-cap-hours" 
                            id="budget-cap-hours" 
                            v-model.number="project.budget_cap_hours"
                            min="0"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6" 
                          />
                           <p class="mt-2 text-xs text-gray-500">Optional: Overall hour cap for the entire project lifecycle.</p>
                        </div>
                      </div>

                      <!-- Budget Cap Dollars -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="budget-cap-dollars" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Budget Cap Dollars</label>
                        </div>
                        <div class="sm:col-span-2">
                          <input 
                            type="number" 
                            name="budget-cap-dollars" 
                            id="budget-cap-dollars" 
                            v-model.number="project.budget_cap_dollars"
                            min="0"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6" 
                          />
                          <p class="mt-2 text-xs text-gray-500">Optional: Overall monetary cap for the entire project lifecycle.</p>
                        </div>
                      </div>

                      <!-- Account Executive -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="project-ae" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Account Executive</label>
                        </div>
                        <div class="sm:col-span-2">
                          <select 
                            id="project-ae"
                            name="project-ae"
                            v-model="project.ae_id"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6"
                          >
                            <option value="">Select an Account Executive</option>
                            <option v-for="ae in staffMembers" :key="ae.ID" :value="ae.ID">
                              {{ ae.first_name }} {{ ae.last_name }}
                            </option>
                          </select>
                        </div>
                      </div>

                      <!-- Sales Development Representative -->
                      <div class="space-y-2 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                        <div>
                          <label for="project-sdr" class="block text-sm/6 font-medium text-gray-900 sm:mt-1.5">Sales Development Representative</label>
                        </div>
                        <div class="sm:col-span-2">
                          <select 
                            id="project-sdr"
                            name="project-sdr"
                            v-model="project.sdr_id"
                            class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-sage sm:text-sm/6"
                          >
                            <option value="">Select a Sales Development Representative</option>
                            <option v-for="sdr in staffMembers" :key="sdr.ID" :value="sdr.ID">
                              {{ sdr.first_name }} {{ sdr.last_name }}
                            </option>
                          </select>
                        </div>
                      </div>

                      <!-- Staffing Assignments (Read-only section modified) -->
                      <div class="px-4 sm:px-6 sm:py-5">
                        <h3 class="text-sm font-medium leading-6 text-gray-900 mb-2">Staffing Assignments</h3>
                        <div v-if="project.staffing_assignments && project.staffing_assignments.length > 0">
                          <ul role="list" class="divide-y divide-gray-200 border-t border-b border-gray-200">
                            <li v-for="assignment in project.staffing_assignments" :key="assignment.ID" class="py-2.5">
                              <div class="flex items-center justify-between gap-x-2">
                                <div class="flex-1 min-w-0">
                                  <p class="text-sm font-medium text-gray-800 truncate">
                                    {{ assignment.employee?.first_name }} {{ assignment.employee?.last_name }}
                                    <span v-if="assignment.employee?.title" class="text-gray-500 text-xs font-normal">({{ assignment.employee?.title }})</span>
                                  </p>
                                  <div class="mt-1 flex flex-col sm:flex-row sm:flex-wrap sm:items-center gap-x-3 gap-y-1 text-xs text-gray-500">
                                      <p v-if="assignment.commitment !== undefined">
                                          <i class="far fa-clock mr-0.5"></i> {{ assignment.commitment }} hr/wk
                                      </p>
                                      <p v-if="assignment.start_date || assignment.end_date">
                                          <i class="far fa-calendar-alt mr-0.5"></i> {{ formatDateForDisplay(assignment.start_date) }} - {{ formatDateForDisplay(assignment.end_date) }}
                                      </p>
                                      <p :class="getStaffingAssignmentStatusBadgeColor(assignment)">
                                          Status: {{ getStaffingAssignmentStatusText(assignment) }}
                                      </p>
                                  </div>
                                </div>
                                <!-- Right side: Action Buttons -->
                                <div class="flex-shrink-0 flex items-center gap-x-1.5">
                                  <button @click.prevent="handleDeleteAssignment(assignment.ID)" type="button" class="p-1 text-gray-400 hover:text-red-600 focus:outline-none focus:ring-2 focus:ring-red-500 rounded">
                                    <span class="sr-only">Delete assignment</span>
                                    <i class="fas fa-trash-alt h-3.5 w-3.5"></i>
                                  </button>
                                </div>
                              </div>
                            </li>
                          </ul>
                        </div>
                        <div v-else class="mt-2">
                           <p class="text-sm text-gray-500 italic">No staff currently assigned to this project.</p>
                        </div>
                      </div>

                    </div>
                  </div>

                  <!-- Action buttons -->
                  <div class="shrink-0 border-t border-gray-200 px-4 py-5 sm:px-6">
                    <div class="flex justify-end space-x-3">
                      <button 
                        type="button" 
                        class="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50" 
                        @click="handleClose"
                      >
                        Cancel
                      </button>
                      <button 
                        type="submit" 
                        :class="[
                          'inline-flex justify-center rounded-md px-3 py-2 text-sm font-semibold text-white shadow-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2',
                          isEditing 
                            ? 'bg-blue-600 hover:bg-blue-700 focus-visible:outline-blue-600' 
                            : 'bg-sage hover:bg-sage-dark focus-visible:outline-sage'
                        ]"
                      >
                        {{ isEditing ? 'Update Project' : 'Create Project' }}
                      </button>
                    </div>
                  </div>
                </form>

              </DialogPanel>
            </TransitionChild>
          </div>
        </div>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue';
import { XMarkIcon } from '@heroicons/vue/24/outline';
import type { Project, StaffingAssignment, Staff } from '../../types/Project';
import { createEmptyProject } from '../../types/Project';
import { fetchAccounts } from '../../api/accounts';
import type { Account } from '../../types/Account';
import { getUsers } from '../../api/timesheet';
import { deleteProjectAssignment } from '../../api/projectAssignments';
import { fetchProjectById } from '../../api/projects';

const project = ref<Project>(createEmptyProject());
const accounts = ref<Account[]>([]);
const staffMembers = ref<Staff[]>([]);
const isEditing = computed(() => !!(project.value && project.value.ID && project.value.ID !== 0));

const props = defineProps<{ isOpen: boolean; projectData: Project | null; }>();
const emit = defineEmits(['close', 'save']);

const formatDateForInput = (dateString?: string | Date): string => {
  if (!dateString) return '';
  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
      if (typeof dateString === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(dateString)) {
        return dateString;
      }
      return '';
    }
    const year = date.getFullYear();
    const month = (date.getMonth() + 1).toString().padStart(2, '0');
    const day = date.getDate().toString().padStart(2, '0');
    return `${year}-${month}-${day}`;
  } catch (e) {
    return '';
  }
};

const formatDateForDisplay = (dateString?: string | Date | null): string => {
  if (!dateString) return 'N/A';
  if (typeof dateString === 'string' && dateString.startsWith('0001-01-01')) return 'N/A';
  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return 'N/A';
    return date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
  } catch (e) {
    return 'N/A';
  }
};

const getStaffingAssignmentStatusText = (assignment: StaffingAssignment): string => {
  const now = new Date();
  now.setHours(0, 0, 0, 0);
  const startDate = assignment.start_date ? new Date(assignment.start_date) : null;
  const endDate = assignment.end_date ? new Date(assignment.end_date) : null;
  if (startDate) startDate.setHours(0,0,0,0);
  if (endDate) endDate.setHours(0,0,0,0);

  if (startDate && now < startDate) return 'Upcoming';
  if (endDate && now > endDate) return 'Ended';
  if (startDate && !endDate) return 'Active (no end date)';
  if (startDate && endDate && now >= startDate && now <= endDate) return 'Active';
  if (!startDate && endDate && now <= endDate) return 'Active (no start date)';
  if (!startDate && !endDate) return 'Active (no dates)';
  return 'Inactive';
};

const getStaffingAssignmentStatusBadgeColor = (assignment: StaffingAssignment): string => {
    const status = getStaffingAssignmentStatusText(assignment);
    if (status.includes('Active')) return 'text-green-700';
    if (status === 'Upcoming') return 'text-blue-700';
    if (status === 'Ended') return 'text-gray-600';
    return 'text-gray-500';
};

const refreshProjectData = async () => {
    if (!project.value || !project.value.ID) return;
    console.log(`Refreshing data for project ID: ${project.value.ID}`);
    try {
        const updatedProjectData = await fetchProjectById(project.value.ID);
        project.value = {
            ...updatedProjectData,
            active_start: formatDateForInput(updatedProjectData.active_start),
            active_end: formatDateForInput(updatedProjectData.active_end),
            description: updatedProjectData.description === undefined ? '' : updatedProjectData.description,
            staffing_assignments: (updatedProjectData.staffing_assignments || []).map((sa: StaffingAssignment) => ({
                ...sa,
                start_date: sa.start_date, 
                end_date: sa.end_date,     
            })),
        };
         console.log("Project data refreshed:", project.value);
    } catch (error) {
        console.error("Failed to refresh project data:", error);
        alert("Failed to refresh project details after update. Please close and reopen the drawer.");
    }
};

watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    if (props.projectData) {
      const tempData = JSON.parse(JSON.stringify(props.projectData));

      const validBillingFrequencies = [
        'BILLING_TYPE_MONTHLY',
        'BILLING_TYPE_PROJECT',
        'BILLING_TYPE_BIWEEKLY',
        'BILLING_TYPE_WEEKLY',
        'BILLING_TYPE_BIMONTHLY'
      ];
      let currentBillingFrequency = tempData.billing_frequency;
      if (!currentBillingFrequency || !validBillingFrequencies.includes(currentBillingFrequency)) {
        // console.warn(`Invalid or missing billing_frequency: '${tempData.billing_frequency}', defaulting to BILLING_TYPE_MONTHLY.`);
        currentBillingFrequency = 'BILLING_TYPE_MONTHLY';
      }

      project.value = {
        ...tempData,
        active_start: formatDateForInput(tempData.active_start),
        active_end: formatDateForInput(tempData.active_end),
        description: tempData.description === undefined ? '' : tempData.description,
        budget_cap_hours: Number(tempData.budget_cap_hours) || 0,
        budget_cap_dollars: Number(tempData.budget_cap_dollars) || 0,
        billing_frequency: currentBillingFrequency, // Use validated or default value
        staffing_assignments: (tempData.staffing_assignments || []).map((sa: StaffingAssignment) => ({
            ...sa,
            start_date: sa.start_date, 
            end_date: sa.end_date,     
        })),
      };
      if (props.projectData.account && props.projectData.account.ID) { project.value.account_id = Number(props.projectData.account.ID); } else if (props.projectData.account_id) { project.value.account_id = Number(props.projectData.account_id); } else { project.value.account_id = 0; }
    } else {
      project.value = createEmptyProject();
    }
  }
}, { deep: true });

onMounted(async () => {
  try {
    const accResponse = await fetchAccounts();
    accounts.value = accResponse || [];
    const staffResponse = await getUsers();
    staffMembers.value = staffResponse || [];
  } catch (error) {
    console.error("Error fetching accounts or staff for drawer:", error);
  }
});

const handleClose = () => {
  emit('close');
};

const handleSubmit = () => {
  if (!project.value.name) {
    alert('Project name is required.');
    return;
  }
  if (!project.value.account_id || Number(project.value.account_id) === 0) {
     alert('Account is required.');
     return;
  }
  project.value.account_id = Number(project.value.account_id);

  if (!project.value.active_start) {
    alert('Start date is required.');
    return;
  }
  if (!project.value.active_end) {
    alert('End date is required.');
    return;
  }
  emit('save', { ...project.value });
};

const handleDeleteAssignment = async (assignmentId: number) => {
  if (!confirm('Are you sure you want to delete this staffing assignment?')) {
    return;
  }
  try {
    await deleteProjectAssignment(assignmentId);
    console.log('Assignment deleted successfully');
    await refreshProjectData(); 
  } catch (error) {
    console.error("Error deleting assignment:", error);
    alert('Failed to delete assignment.');
  }
};

</script> 