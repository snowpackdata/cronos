<script setup lang="ts">
import { ref, watch } from 'vue';
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue';
import { XMarkIcon, CalendarIcon } from '@heroicons/vue/24/outline';
import googleCalendarAPI, { type CalendarEvent } from '../../api/googleCalendar';

// Props
const props = defineProps<{
  show: boolean;
  startDate: string;
  endDate: string;
}>();

// Emits
const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'select', event: CalendarEvent): void;
}>();

// State
const events = ref<CalendarEvent[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);
const selectedEventId = ref<string | null>(null);

// Load calendar events when modal opens
watch(() => props.show, (newShow) => {
  if (newShow) {
    loadEvents();
  }
});

// Fetch events from API
const loadEvents = async () => {
  loading.value = true;
  error.value = null;
  
  try {
    events.value = await googleCalendarAPI.getCalendarEvents(
      props.startDate,
      props.endDate
    );
  } catch (err: any) {
    error.value = err.response?.data?.message || 'Failed to load calendar events';
    console.error('Error loading calendar events:', err);
  } finally {
    loading.value = false;
  }
};

// Format date/time for display
const formatDateTime = (isoString: string): string => {
  const date = new Date(isoString);
  return date.toLocaleString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
    hour12: true
  });
};

// Strip HTML tags from text
const stripHtml = (html: string): string => {
  const tmp = document.createElement('div');
  tmp.innerHTML = html;
  return tmp.textContent || tmp.innerText || '';
};

// Calculate duration in hours
const getDuration = (start: string, end: string): string => {
  const startDate = new Date(start);
  const endDate = new Date(end);
  const hours = (endDate.getTime() - startDate.getTime()) / (1000 * 60 * 60);
  return hours.toFixed(1);
};

// Handle event selection
const selectEvent = (eventId: string) => {
  selectedEventId.value = eventId;
};

// Handle use event button click
const useSelectedEvent = () => {
  console.log('useSelectedEvent called, selectedEventId:', selectedEventId.value);
  const event = events.value.find(e => e.id === selectedEventId.value);
  console.log('Found event:', event);
  if (event) {
    console.log('Emitting select event with:', event);
    emit('select', event);
  } else {
    console.error('No event found with id:', selectedEventId.value);
  }
};

// Close modal
const close = () => {
  emit('close');
};
</script>

<template>
  <TransitionRoot as="template" :show="show">
    <Dialog as="div" class="relative z-50" @close="close">
      <TransitionChild
        as="template"
        enter="ease-out duration-300"
        enter-from="opacity-0"
        enter-to="opacity-100"
        leave="ease-in duration-200"
        leave-from="opacity-100"
        leave-to="opacity-0"
      >
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
      </TransitionChild>

      <div class="fixed inset-0 z-10 overflow-y-auto">
        <div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
          <TransitionChild
            as="template"
            enter="ease-out duration-300"
            enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
            enter-to="opacity-100 translate-y-0 sm:scale-100"
            leave="ease-in duration-200"
            leave-from="opacity-100 translate-y-0 sm:scale-100"
            leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
          >
            <DialogPanel class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-2xl sm:p-6">
              <div class="absolute right-0 top-0 pr-4 pt-4">
                <button
                  type="button"
                  class="rounded-md bg-white text-gray-400 hover:text-gray-500 focus:outline-none"
                  @click="close"
                >
                  <span class="sr-only">Close</span>
                  <XMarkIcon class="h-6 w-6" aria-hidden="true" />
                </button>
              </div>
              
              <div class="sm:flex sm:items-start">
                <div class="mx-auto flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-blue-100 sm:mx-0">
                  <CalendarIcon class="h-5 w-5 text-blue-600" aria-hidden="true" />
                </div>
                <div class="mt-3 text-center sm:ml-3 sm:mt-0 sm:text-left w-full">
                  <DialogTitle as="h3" class="text-base font-semibold leading-6 text-gray-900">
                    Import from Google Calendar
                  </DialogTitle>
                  <p class="text-xs text-gray-500 mt-0.5">
                    Select a calendar event to import as a timesheet entry
                  </p>
                </div>
              </div>

              <div class="mt-4">
                <!-- Loading state -->
                <div v-if="loading" class="text-center py-8">
                  <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                  <p class="mt-2 text-sm text-gray-500">Loading calendar events...</p>
                </div>

                <!-- Error state -->
                <div v-else-if="error" class="rounded-md bg-red-50 p-4">
                  <p class="text-sm text-red-800">{{ error }}</p>
                </div>

                <!-- No events state -->
                <div v-else-if="events.length === 0" class="text-center py-8">
                  <CalendarIcon class="mx-auto h-12 w-12 text-gray-400" />
                  <p class="mt-2 text-sm text-gray-500">No calendar events found for this period</p>
                </div>

                <!-- Events list -->
                <div v-else class="space-y-2 max-h-96 overflow-y-auto">
                  <div
                    v-for="event in events"
                    :key="event.id"
                    @click="selectEvent(event.id)"
                    class="cursor-pointer rounded-lg border p-3 transition-colors"
                    :class="[
                      selectedEventId === event.id
                        ? 'border-blue-500 bg-blue-50'
                        : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
                    ]"
                  >
                    <div class="flex justify-between items-start gap-3">
                      <div class="flex-1 min-w-0">
                        <h4 class="font-semibold text-gray-900 text-sm truncate">{{ event.summary }}</h4>
                        <p v-if="event.description && stripHtml(event.description).trim()" class="mt-1 text-xs text-gray-600 line-clamp-1">
                          {{ stripHtml(event.description) }}
                        </p>
                        <div class="mt-1.5 text-xs text-gray-500">
                          {{ formatDateTime(event.start) }} - {{ formatDateTime(event.end).split(',').pop()?.trim() }}
                        </div>
                      </div>
                      <div class="flex-shrink-0">
                        <span class="inline-flex items-center rounded-full bg-blue-100 px-2 py-0.5 text-xs font-semibold text-blue-700">
                          {{ getDuration(event.start, event.end) }}h
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Actions -->
              <div class="mt-5 flex justify-end gap-2">
                <button
                  type="button"
                  class="inline-flex justify-center rounded-md bg-white px-3 py-1.5 text-sm font-medium text-gray-700 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                  @click="close"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  :disabled="!selectedEventId"
                  class="inline-flex justify-center rounded-md bg-blue-600 px-3 py-1.5 text-sm font-medium text-white shadow-sm hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                  @click="useSelectedEvent"
                >
                  Use This Event
                </button>
              </div>
            </DialogPanel>
          </TransitionChild>
        </div>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

