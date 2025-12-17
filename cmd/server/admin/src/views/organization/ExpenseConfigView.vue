<template>
  <div class="min-h-screen bg-gray-50 p-4">
    <div class="max-w-7xl mx-auto">
      <div class="mb-4">
        <h1 class="text-lg font-bold text-gray-900">Expense Configuration</h1>
        <p class="text-xs text-gray-500">Manage expense categories and tags</p>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <!-- Categories Column -->
        <div>
          <div class="mb-2 flex justify-between items-center">
            <h2 class="text-sm font-semibold text-gray-900">Categories</h2>
            <button
              @click="openCreateCategoryModal"
              class="px-2 py-1 text-xs font-medium text-white bg-blue-600 hover:bg-blue-700 rounded"
            >
              New Category
            </button>
          </div>

          <div class="bg-white shadow-sm rounded border border-gray-200">
            <table class="min-w-full divide-y divide-gray-200">
              <thead class="bg-gray-50">
                <tr>
                  <th class="px-2 py-1 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                  <th class="px-2 py-1 text-center text-xs font-medium text-gray-500 uppercase">Status</th>
                  <th class="px-2 py-1 text-center text-xs font-medium text-gray-500 uppercase">Actions</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-gray-200">
                <tr v-if="categories.length === 0">
                  <td colspan="3" class="px-2 py-4 text-center text-xs text-gray-500">
                    No categories found
                  </td>
                </tr>
                <tr v-for="category in categories" :key="category.ID" class="hover:bg-gray-50">
                  <td class="px-2 py-1 text-xs text-gray-900">
                    <div class="font-medium">{{ category.name }}</div>
                    <div v-if="category.description" class="text-gray-500 text-xs">{{ category.description }}</div>
                  </td>
                  <td class="px-2 py-1 text-center">
                    <span :class="category.active ? 'text-green-600' : 'text-gray-400'" class="text-xs font-medium">
                      {{ category.active ? 'Active' : 'Inactive' }}
                    </span>
                  </td>
                  <td class="px-2 py-1 text-center">
                    <div class="flex justify-center gap-1">
                      <button
                        @click="openEditCategoryModal(category)"
                        class="px-2 py-1 bg-sage text-white rounded hover:bg-sage-dark text-xs"
                        title="Edit"
                      >
                        <i class="fa fa-edit"></i>
                      </button>
                      <button
                        @click="confirmDeleteCategory(category)"
                        class="px-2 py-1 bg-red-600 text-white rounded hover:bg-red-700 text-xs"
                        title="Delete"
                      >
                        <i class="fa fa-trash"></i>
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Tags Column -->
        <div>
          <div class="mb-2 flex justify-between items-center">
            <h2 class="text-sm font-semibold text-gray-900">Tags</h2>
            <button
              @click="openCreateTagModal"
              class="px-2 py-1 text-xs font-medium text-white bg-blue-600 hover:bg-blue-700 rounded"
            >
              New Tag
            </button>
          </div>

          <div class="bg-white shadow-sm rounded border border-gray-200">
            <table class="min-w-full divide-y divide-gray-200">
              <thead class="bg-gray-50">
                <tr>
                  <th class="px-2 py-1 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                  <th class="px-2 py-1 text-right text-xs font-medium text-gray-500 uppercase">Budget</th>
                  <th class="px-2 py-1 text-center text-xs font-medium text-gray-500 uppercase">Status</th>
                  <th class="px-2 py-1 text-center text-xs font-medium text-gray-500 uppercase">Actions</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-gray-200">
                <tr v-if="tags.length === 0">
                  <td colspan="4" class="px-2 py-4 text-center text-xs text-gray-500">
                    No tags found
                  </td>
                </tr>
                <tr v-for="tag in tags" :key="tag.ID" class="hover:bg-gray-50">
                  <td class="px-2 py-1 text-xs text-gray-900">
                    <div class="font-medium">{{ tag.name }}</div>
                    <div v-if="tag.description" class="text-gray-500 text-xs">{{ tag.description }}</div>
                    <div v-if="tag.budget" class="mt-0.5">
                      <div class="w-full bg-gray-200 rounded-full h-1">
                        <div 
                          class="h-1 rounded-full transition-all"
                          :class="getBudgetBarClass(tag.budget_percentage)"
                          :style="`width: ${Math.min(tag.budget_percentage ?? 0, 100)}%`"
                        ></div>
                      </div>
                      <div class="flex justify-between text-xs mt-0.5">
                        <span class="text-gray-500">{{ formatCurrency(tag.total_spent || 0) }} / {{ formatCurrency(tag.budget) }}</span>
                        <span :class="getBudgetRemainingClass(tag)">{{ tag.budget_percentage }}%</span>
                      </div>
                    </div>
                  </td>
                  <td class="px-2 py-1 text-right text-xs text-gray-900">
                    {{ tag.budget ? formatCurrency(tag.budget) : '-' }}
                  </td>
                  <td class="px-2 py-1 text-center">
                    <span :class="tag.active ? 'text-green-600' : 'text-gray-400'" class="text-xs font-medium">
                      {{ tag.active ? 'Active' : 'Inactive' }}
                    </span>
                  </td>
                  <td class="px-2 py-1 text-center">
                    <div class="flex justify-center gap-1">
                      <button
                        @click="openEditTagModal(tag)"
                        class="px-2 py-1 bg-sage text-white rounded hover:bg-sage-dark text-xs"
                        title="Edit"
                      >
                        <i class="fa fa-edit"></i>
                      </button>
                      <button
                        @click="confirmDeleteTag(tag)"
                        class="px-2 py-1 bg-red-600 text-white rounded hover:bg-red-700 text-xs"
                        title="Delete"
                      >
                        <i class="fa fa-trash"></i>
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <!-- Category Modal -->
    <div v-if="showCategoryModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex items-center justify-center min-h-screen px-4">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 backdrop-blur-sm transition-opacity" @click="closeCategoryModal"></div>
        
        <div class="relative bg-white rounded shadow-xl max-w-lg w-full">
          <div class="bg-sage text-white px-4 py-2 flex justify-between items-center">
            <h3 class="text-xs font-semibold">{{ isEditingCategory ? 'Edit Category' : 'New Category' }}</h3>
            <button @click="closeCategoryModal" class="text-white hover:text-gray-200">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <div class="p-4">
            <div class="space-y-2">
              <div>
                <label class="block text-xs font-medium text-gray-700 mb-0.5">
                  Category Name <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="categoryForm.name"
                  type="text"
                  required
                  placeholder="e.g., Travel, Technology"
                  class="bg-white border border-gray-300 text-gray-900 text-xs rounded focus:ring-sage focus:border-sage block w-full p-1.5"
                />
              </div>

              <div>
                <label class="block text-xs font-medium text-gray-700 mb-0.5">Description</label>
                <textarea
                  v-model="categoryForm.description"
                  rows="2"
                  placeholder="Optional description"
                  class="bg-white border border-gray-300 text-gray-900 text-xs rounded focus:ring-sage focus:border-sage block w-full p-1.5"
                ></textarea>
              </div>

              <div v-if="isEditingCategory">
                <label class="flex items-center">
                  <input
                    v-model="categoryForm.active"
                    type="checkbox"
                    class="rounded border-gray-300 text-sage focus:ring-sage"
                  />
                  <span class="ml-2 text-xs text-gray-700">Active</span>
                </label>
              </div>

              <div v-if="categoryError" class="text-xs text-red-600 bg-red-50 p-2 rounded">
                {{ categoryError }}
              </div>
            </div>

            <div class="mt-3 flex justify-end gap-2">
              <button
                @click="closeCategoryModal"
                class="px-2.5 py-1 text-xs font-semibold text-gray-900 bg-white rounded shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                @click="saveCategory"
                :disabled="isSavingCategory"
                class="px-2.5 py-1 text-xs font-semibold text-white bg-sage rounded shadow-sm hover:bg-sage-dark disabled:opacity-50"
              >
                {{ isSavingCategory ? 'Saving...' : 'Save' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tag Modal -->
    <div v-if="showTagModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex items-center justify-center min-h-screen px-4">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 backdrop-blur-sm transition-opacity" @click="closeTagModal"></div>
        
        <div class="relative bg-white rounded shadow-xl max-w-lg w-full">
          <div class="bg-sage text-white px-4 py-2 flex justify-between items-center">
            <h3 class="text-xs font-semibold">{{ isEditingTag ? 'Edit Tag' : 'New Tag' }}</h3>
            <button @click="closeTagModal" class="text-white hover:text-gray-200">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <div class="p-4">
            <div class="space-y-2">
              <div>
                <label class="block text-xs font-medium text-gray-700 mb-0.5">
                  Tag Name <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="tagForm.name"
                  type="text"
                  required
                  placeholder="e.g., Offsite 2025, Q4 Campaign"
                  class="bg-white border border-gray-300 text-gray-900 text-xs rounded focus:ring-sage focus:border-sage block w-full p-1.5"
                />
              </div>

              <div>
                <label class="block text-xs font-medium text-gray-700 mb-0.5">Description</label>
                <textarea
                  v-model="tagForm.description"
                  rows="2"
                  placeholder="Optional description"
                  class="bg-white border border-gray-300 text-gray-900 text-xs rounded focus:ring-sage focus:border-sage block w-full p-1.5"
                ></textarea>
              </div>

              <div>
                <label class="block text-xs font-medium text-gray-700 mb-0.5">Budget (optional)</label>
                <div class="flex rounded border border-gray-300 bg-white focus-within:ring-sage focus-within:border-sage overflow-hidden">
                  <span class="flex items-center px-2 text-gray-500 text-xs bg-gray-50 border-r border-gray-300">$</span>
                  <input
                    v-model.number="budgetDollars"
                    type="number"
                    step="0.01"
                    min="0"
                    placeholder="0.00"
                    class="flex-1 text-gray-900 text-xs p-1.5 focus:outline-none border-0"
                  />
                </div>
                <p class="mt-1 text-xs text-gray-500">Leave empty for no budget tracking</p>
              </div>

              <div>
                <label class="flex items-center">
                  <input
                    v-model="tagForm.active"
                    type="checkbox"
                    class="rounded border-gray-300 text-sage focus:ring-sage"
                  />
                  <span class="ml-2 text-xs text-gray-700">Active (can be used for new expenses)</span>
                </label>
              </div>

              <div v-if="tagError" class="text-xs text-red-600 bg-red-50 p-2 rounded">
                {{ tagError }}
              </div>
            </div>

            <div class="mt-3 flex justify-end gap-2">
              <button
                @click="closeTagModal"
                class="px-2.5 py-1 text-xs font-semibold text-gray-900 bg-white rounded shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                @click="saveTag"
                :disabled="isSavingTag"
                class="px-2.5 py-1 text-xs font-semibold text-white bg-sage rounded shadow-sm hover:bg-sage-dark disabled:opacity-50"
              >
                {{ isSavingTag ? 'Saving...' : 'Save' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { getExpenseCategories, createExpenseCategory, updateExpenseCategory, deleteExpenseCategory } from '../../api/expenseCategories';
import { getExpenseTags, createExpenseTag, updateExpenseTag, deleteExpenseTag } from '../../api/expenseTags';
import type { ExpenseCategory } from '../../types/ExpenseCategory';
import type { ExpenseTag } from '../../types/ExpenseTag';

// Categories
const categories = ref<ExpenseCategory[]>([]);
const showCategoryModal = ref(false);
const isEditingCategory = ref(false);
const isSavingCategory = ref(false);
const categoryError = ref('');
const editingCategory = ref<ExpenseCategory | null>(null);

const categoryForm = ref({
  name: '',
  description: '',
  active: true
});

// Tags
const tags = ref<ExpenseTag[]>([]);
const showTagModal = ref(false);
const isEditingTag = ref(false);
const isSavingTag = ref(false);
const tagError = ref('');
const editingTag = ref<ExpenseTag | null>(null);

const tagForm = ref({
  name: '',
  description: '',
  active: true,
  budget: null as number | null
});

const budgetDollars = computed({
  get: () => tagForm.value.budget ? tagForm.value.budget / 100 : null,
  set: (val) => tagForm.value.budget = val ? Math.round(val * 100) : null
});

onMounted(() => {
  fetchCategories();
  fetchTags();
});

async function fetchCategories() {
  try {
    categories.value = await getExpenseCategories(false);
  } catch (error) {
    console.error('Failed to fetch expense categories:', error);
  }
}

async function fetchTags() {
  try {
    tags.value = await getExpenseTags(false);
  } catch (error) {
    console.error('Failed to fetch expense tags:', error);
  }
}

function formatCurrency(cents: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD'
  }).format(cents / 100);
}

function getBudgetRemainingClass(tag: ExpenseTag): string {
  if (!tag.remaining_budget) return 'text-gray-900';
  if (tag.remaining_budget < 0) return 'text-red-600 font-semibold';
  if (tag.budget_percentage && tag.budget_percentage > 90) return 'text-orange-600';
  return 'text-green-600';
}

function getBudgetBarClass(percentage: number | null | undefined): string {
  if (!percentage) return 'bg-green-500';
  if (percentage >= 100) return 'bg-red-600';
  if (percentage >= 90) return 'bg-orange-500';
  if (percentage >= 75) return 'bg-yellow-500';
  return 'bg-green-500';
}

// Category CRUD
function openCreateCategoryModal() {
  isEditingCategory.value = false;
  editingCategory.value = null;
  categoryForm.value = {
    name: '',
    description: '',
    active: true
  };
  categoryError.value = '';
  showCategoryModal.value = true;
}

function openEditCategoryModal(category: ExpenseCategory) {
  isEditingCategory.value = true;
  editingCategory.value = category;
  categoryForm.value = {
    name: category.name,
    description: category.description,
    active: category.active
  };
  categoryError.value = '';
  showCategoryModal.value = true;
}

function closeCategoryModal() {
  showCategoryModal.value = false;
  categoryError.value = '';
}

async function saveCategory() {
  if (!categoryForm.value.name.trim()) {
    categoryError.value = 'Category name is required';
    return;
  }

  isSavingCategory.value = true;
  categoryError.value = '';

  try {
    if (isEditingCategory.value && editingCategory.value) {
      await updateExpenseCategory(editingCategory.value.ID, categoryForm.value);
    } else {
      await createExpenseCategory({
        name: categoryForm.value.name,
        description: categoryForm.value.description
      });
    }
    await fetchCategories();
    closeCategoryModal();
  } catch (error: any) {
    categoryError.value = error.response?.data?.error || 'Failed to save category';
  } finally {
    isSavingCategory.value = false;
  }
}

async function confirmDeleteCategory(category: ExpenseCategory) {
  if (!confirm(`Delete category "${category.name}"? This cannot be undone.`)) {
    return;
  }

  try {
    await deleteExpenseCategory(category.ID);
    await fetchCategories();
  } catch (error: any) {
    alert(error.response?.data?.error || 'Failed to delete category');
  }
}

// Tag CRUD
function openCreateTagModal() {
  isEditingTag.value = false;
  editingTag.value = null;
  tagForm.value = {
    name: '',
    description: '',
    active: true,
    budget: null
  };
  tagError.value = '';
  showTagModal.value = true;
}

function openEditTagModal(tag: ExpenseTag) {
  isEditingTag.value = true;
  editingTag.value = tag;
  tagForm.value = {
    name: tag.name,
    description: tag.description,
    active: tag.active,
    budget: tag.budget
  };
  tagError.value = '';
  showTagModal.value = true;
}

function closeTagModal() {
  showTagModal.value = false;
  tagError.value = '';
}

async function saveTag() {
  if (!tagForm.value.name.trim()) {
    tagError.value = 'Tag name is required';
    return;
  }

  isSavingTag.value = true;
  tagError.value = '';

  try {
    if (isEditingTag.value && editingTag.value) {
      await updateExpenseTag(editingTag.value.ID, tagForm.value);
    } else {
      await createExpenseTag(tagForm.value);
    }
    await fetchTags();
    closeTagModal();
  } catch (error: any) {
    tagError.value = error.response?.data?.error || 'Failed to save tag';
  } finally {
    isSavingTag.value = false;
  }
}

async function confirmDeleteTag(tag: ExpenseTag) {
  if (!confirm(`Delete tag "${tag.name}"? This cannot be undone.`)) {
    return;
  }

  try {
    await deleteExpenseTag(tag.ID);
    await fetchTags();
  } catch (error: any) {
    alert(error.response?.data?.error || 'Failed to delete tag. It may still be in use by expenses.');
  }
}
</script>

