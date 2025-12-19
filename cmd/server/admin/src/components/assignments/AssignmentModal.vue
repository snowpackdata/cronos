<template>
  <TransitionRoot as="template" :show="isOpen">
    <Dialog class="relative z-20" @close="handleClose"> <!-- Increased z-index -->
      <!-- Backdrop -->
      <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0" enter-to="opacity-100" leave="ease-in duration-200" leave-from="opacity-100" leave-to="opacity-0">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
      </TransitionChild>

      <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
        <div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
          <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95" enter-to="opacity-100 translate-y-0 sm:scale-100" leave="ease-in duration-200" leave-from="opacity-100 translate-y-0 sm:scale-100" leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95">
            <DialogPanel class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6">
              <form @submit.prevent="handleSave">
                <div>
                  <DialogTitle as="h3" class="text-base font-semibold leading-6 text-gray-900">
                    {{ isEditing ? 'Edit Assignment' : 'Add New Assignment' }}
                  </DialogTitle>
                  <div class="mt-4 space-y-4">
                    <!-- Project Dropdown or Display -->
                    <div>
                      <label for="assignment-project" class="block text-sm font-medium leading-6 text-gray-900">Project</label>
                      <!-- Single project: just display it -->
                      <div 
                        v-if="projectsList.length === 1"
                        class="mt-1 block w-full rounded-md border-gray-300 py-1.5 pl-3 pr-10 text-gray-900 bg-gray-50 border sm:text-sm"
                      >
                        {{ projectsList[0].name }}{{ projectsList[0].account?.name ? ` (${projectsList[0].account.name})` : '' }}
                      </div>
                      <!-- Multiple projects: show dropdown -->
                      <select 
                        v-else
                        id="assignment-project"
                        name="assignment-project"
                        v-model.number="formData.project_id"
                        required
                        class="mt-1 block w-full rounded-md border-gray-300 py-1.5 pl-3 pr-10 text-gray-900 shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm"
                      >
                        <option :value="undefined" disabled>Select project...</option>
                        <option v-for="project in projectsList" :key="project.ID" :value="Number(project.ID)">
                          {{ project.name }}{{ project.account?.name ? ` (${project.account.name})` : '' }}
                        </option>
                      </select>
                    </div>

                    <!-- Staff Member Dropdown -->
                    <div>
                      <label for="assignment-staff" class="block text-sm font-medium leading-6 text-gray-900">Staff Member</label>
                      <select 
                        id="assignment-staff"
                        name="assignment-staff"
                        v-model.number="formData.employee_id"
                        required
                        class="mt-1 block w-full rounded-md border-gray-300 py-1.5 pl-3 pr-10 text-gray-900 shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm"
                      >
                        <option :value="undefined" disabled>Select staff...</option>
                        <!-- Current assignment's employee if not in staff list -->
                        <option 
                          v-if="assignmentData?.employee && assignmentData?.employee_id && !staffList.find(s => s.ID === assignmentData?.employee_id)" 
                          :value="Number(assignmentData.employee_id)"
                        >
                          {{ assignmentData.employee.first_name }} {{ assignmentData.employee.last_name }}{{ assignmentData.employee.title ? ` (${assignmentData.employee.title})` : '' }} (Current)
                        </option>
                        <option v-for="staff in staffList" :key="staff.ID" :value="Number(staff.ID)">
                          {{ staff.first_name }} {{ staff.last_name }}{{ staff.title ? ` (${staff.title})` : '' }}
                        </option>
                      </select>
                    </div>

                    <!-- Start Date -->
                    <div>
                      <label for="assignment-start-date" class="block text-sm font-medium leading-6 text-gray-900">Start Date</label>
                      <input 
                        type="date"
                        id="assignment-start-date"
                        name="assignment-start-date"
                        v-model="formData.start_date"
                        required
                        class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm py-1.5 px-2"
                      />
                    </div>

                    <!-- End Date -->
                    <div>
                      <label for="assignment-end-date" class="block text-sm font-medium leading-6 text-gray-900">End Date</label>
                      <input 
                        type="date"
                        id="assignment-end-date"
                        name="assignment-end-date"
                        v-model="formData.end_date"
                        required
                        class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm py-1.5 px-2"
                      />
                    </div>

                    <!-- Timeline Editor -->
                    <div v-if="formData.start_date && formData.end_date">
                      <TimelineEditor
                        :start-date="typeof formData.start_date === 'string' ? formData.start_date : formatDateForInput(formData.start_date)"
                        :end-date="typeof formData.end_date === 'string' ? formData.end_date : formatDateForInput(formData.end_date)"
                        :segments="formData.segments"
                        @update:segments="formData.segments = $event"
                      />
                    </div>

                    <!-- Fallback: Simple Commitment (deprecated, but shown if no segments) -->
                    <div v-if="(!formData.segments || formData.segments.length === 0) && formData.start_date && formData.end_date" class="text-xs text-gray-500 bg-yellow-50 p-2 rounded border border-yellow-200">
                      <p class="font-medium mb-1">Set weekly commitments above</p>
                      <p>Use the timeline editor to set variable commitments for different weeks.</p>
                    </div>
                  </div>
                </div>
                <div class="mt-5 sm:mt-6">
                  <div class="sm:grid sm:grid-flow-row-dense sm:grid-cols-2 sm:gap-3">
                    <button 
                      type="submit"
                      class="inline-flex w-full justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 sm:col-start-2"
                    >
                      {{ isEditing ? 'Update Assignment' : 'Add Assignment' }}
                    </button>
                    <button 
                      type="button"
                      class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:col-start-1 sm:mt-0"
                      @click="handleClose"
                    >
                      Cancel
                    </button>
                  </div>
                  <button 
                    v-if="isEditing && assignmentData?.ID"
                    type="button"
                    class="mt-3 inline-flex w-full justify-center rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500"
                    @click="handleDelete"
                  >
                    Delete Assignment
                  </button>
                </div>
              </form>
            </DialogPanel>
          </TransitionChild>
        </div>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

<script setup lang="ts">
import { ref, watch, computed, defineEmits } from 'vue';
import type { PropType } from 'vue';
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue';
import type { StaffingAssignment, Staff, Project } from '../../types/Project';
import TimelineEditor from './TimelineEditor.vue';
import type { CommitmentSegment } from './TimelineEditor.vue';

const emit = defineEmits(['close', 'save', 'delete']);

// Helper to format date strings (e.g., ISO) to YYYY-MM-DD for date inputs
// (Same as in ProjectDrawer)
const formatDateForInput = (dateString?: string | Date): string => {
  if (!dateString) return '';
  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
        if (typeof dateString === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(dateString)) { return dateString; }
        return ''; 
    }
    const year = date.getFullYear();
    const month = (date.getMonth() + 1).toString().padStart(2, '0');
    const day = date.getDate().toString().padStart(2, '0');
    return `${year}-${month}-${day}`;
  } catch (e) { return ''; }
};

// Define component props
const props = defineProps({
  isOpen: {
    type: Boolean,
    required: true,
  },
  assignmentData: {
    type: Object as PropType<StaffingAssignment | null>,
    default: null,
  },
  staffList: {
    type: Array as PropType<Staff[]>,
    required: true,
    default: () => [],
  },
  projectsList: {
    type: Array as PropType<Project[]>,
    default: () => [],
  },
  projectData: {
    type: Object as PropType<{ active_start: string; active_end: string } | null>,
    default: null,
  },
});

// Adjusted form data type to use employee_id, project_id and segments
type AssignmentFormData = Partial<Omit<StaffingAssignment, 'ID' | 'Employee'>> & {
  segments?: CommitmentSegment[];
  project_id?: number;
};
const formData = ref<AssignmentFormData>({});

// Computed property to check if editing
const isEditing = computed(() => !!props.assignmentData);

// Watcher adjusted to use employee_id
watch(() => props.isOpen, (isOpen) => {
  if (isOpen) {
    const assignment = props.assignmentData;
    console.log('AssignmentModal opened. Editing assignment:', assignment);
    console.log('Staff list:', props.staffList);
    if (assignment && assignment.ID) {
      // Editing: Populate form with segments
      let segments: CommitmentSegment[] | undefined = undefined;
      
      // Parse segments from commitment_schedule if available
      if (assignment.commitment_schedule) {
        try {
          const parsed = JSON.parse(assignment.commitment_schedule);
          segments = parsed.segments || [];
        } catch (e) {
          console.warn('Failed to parse commitment_schedule:', e);
        }
      }
      
      // If no segments but has commitment, create a simple segment
      if (!segments && assignment.commitment) {
        segments = [{
          start_date: formatDateForInput(assignment.start_date),
          end_date: formatDateForInput(assignment.end_date),
          commitment: Number(assignment.commitment)
        }];
      }
      
      formData.value = {
        employee_id: assignment.employee_id ? Number(assignment.employee_id) : undefined,
        project_id: assignment.project_id ? Number(assignment.project_id) : undefined,
        commitment: assignment.commitment !== undefined ? Number(assignment.commitment) : undefined,
        start_date: formatDateForInput(assignment.start_date),
        end_date: formatDateForInput(assignment.end_date),
        segments: segments,
      };
      console.log('Set formData for edit:', formData.value);
      console.log('Looking for employee_id:', formData.value.employee_id, 'in staff list with IDs:', props.staffList.map(s => s.ID));
    } else {
      // Adding: Reset form with project dates as defaults
      const defaultStartDate = props.projectData?.active_start 
        ? formatDateForInput(props.projectData.active_start)
        : formatDateForInput(new Date());
      const defaultEndDate = props.projectData?.active_end 
        ? formatDateForInput(props.projectData.active_end)
        : '';
      
      // Auto-select project if only one exists
      const autoProjectId = props.projectsList.length === 1 
        ? props.projectsList[0].ID 
        : (assignment?.project_id || undefined);
      
      formData.value = {
        employee_id: assignment?.employee_id || undefined,
        project_id: autoProjectId,
        commitment: undefined,
        start_date: defaultStartDate,
        end_date: defaultEndDate,
        segments: [],
      };
      console.log('Reset formData for add:', formData.value);
    }
  }
}, { immediate: true });

// Method to handle closing the modal
const handleClose = () => {
  emit('close');
};

// handleSave adjusted to use employee_id and project_id
const handleSave = () => {
  // Validation
  if (!formData.value.project_id) {
    alert('Please select a project.');
    return;
  }
  if (!formData.value.employee_id) {
    alert('Please select a staff member.');
    return;
  }
  if (!formData.value.start_date || !formData.value.end_date) {
    alert('Please provide both start and end dates.');
    return;
  }
  if (new Date(formData.value.end_date) < new Date(formData.value.start_date)) {
      alert('End date cannot be before start date.');
      return;
  }

  // Calculate average commitment from segments for backward compatibility
  let avgCommitment = 0;
  if (formData.value.segments && formData.value.segments.length > 0) {
    const totalCommitment = formData.value.segments.reduce((sum, seg) => sum + seg.commitment, 0);
    avgCommitment = Math.round(totalCommitment / formData.value.segments.length);
  }

  emit('save', {
    employee_id: formData.value.employee_id,
    project_id: formData.value.project_id,
    commitment: avgCommitment || undefined, // Fallback/average for legacy field
    start_date: formData.value.start_date,
    end_date: formData.value.end_date,
    segments: formData.value.segments,
  });
};

const handleDelete = () => {
  if (props.assignmentData?.ID) {
    emit('delete', props.assignmentData.ID);
  }
};

</script> 