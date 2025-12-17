<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
    <!-- Header -->
    <div class="mb-3">
      <h1 class="text-xl font-bold text-gray-900">Offline Journal Review</h1>
      <p class="mt-0.5 text-xs text-gray-500">
        Import and review external accounting entries from CSV
      </p>
    </div>

    <!-- CSV Upload Section -->
    <div class="bg-white shadow rounded p-2 mb-2">
      <h2 class="text-xs font-medium text-gray-900 mb-2">Upload CSV File</h2>
      
      <!-- Step 1: File Upload -->
      <div v-if="!csvPreview">
        <input
          ref="csvFileInput"
          type="file"
          accept=".csv"
          @change="handleCSVFileSelect"
          class="block w-full text-xs text-gray-500
            file:mr-2 file:py-1 file:px-2
            file:rounded file:border-0
            file:text-xs file:font-semibold
            file:bg-blue-50 file:text-blue-700
            hover:file:bg-blue-100"
        />
      </div>

      <!-- Step 2: Preview and Column Mapping -->
      <div v-else class="space-y-2">
        <div class="flex justify-between items-center">
          <h3 class="text-xs font-medium text-gray-900">Preview and Map Columns</h3>
          <button @click="cancelCSVPreview" class="text-xs text-gray-600 hover:text-gray-900">
            Cancel
          </button>
        </div>

        <!-- Column Mapping -->
        <div class="grid grid-cols-4 gap-2 bg-gray-50 p-2 rounded">
          <div>
            <label class="block text-xs text-gray-700 mb-0.5">Date Column</label>
            <select v-model.number="csvOptions.dateCol" class="w-full px-1 py-1 text-xs border rounded">
              <option v-for="(header, index) in csvPreview.headers" :key="index" :value="index">
                {{ index }}: {{ header }}
              </option>
            </select>
          </div>
          <div>
            <label class="block text-xs text-gray-700 mb-0.5">Description Column</label>
            <select v-model.number="csvOptions.descCol" class="w-full px-1 py-1 text-xs border rounded">
              <option v-for="(header, index) in csvPreview.headers" :key="index" :value="index">
                {{ index }}: {{ header }}
              </option>
            </select>
          </div>
          <div>
            <label class="block text-xs text-gray-700 mb-0.5">Amount Column</label>
            <select v-model.number="csvOptions.amountCol" class="w-full px-1 py-1 text-xs border rounded">
              <option v-for="(header, index) in csvPreview.headers" :key="index" :value="index">
                {{ index }}: {{ header }}
              </option>
            </select>
          </div>
          <div>
            <label class="block text-xs text-gray-700 mb-0.5">Date Format <span class="text-gray-500 font-normal">(YYYY-MM-DD, MM/DD/YYYY, etc.)</span></label>
            <input v-model="csvOptions.dateFormat" type="text" placeholder="YYYY-MM-DD or leave empty to auto-detect" class="w-full px-1 py-1 text-xs border rounded" />
          </div>
        </div>

        <div class="flex items-center justify-between">
          <label class="flex items-center text-xs text-gray-700">
            <input v-model="csvOptions.hasHeader" type="checkbox" class="mr-1" />
            First row is header
          </label>
        <button
            @click="uploadCSV"
            :disabled="isUploading"
            class="px-3 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-300"
        >
            {{ isUploading ? 'Uploading...' : 'Import Transactions' }}
        </button>
        </div>

        <!-- Preview Table -->
        <div class="border rounded overflow-x-auto max-h-48 overflow-y-auto">
          <table class="min-w-full text-xs">
            <thead class="bg-gray-50 sticky top-0">
              <tr>
                <th v-for="(header, index) in csvPreview.headers" :key="index" class="px-2 py-1 text-left text-xs font-medium text-gray-500">
                  {{ index }}: {{ header }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              <tr v-for="(row, rowIndex) in csvPreview.rows.slice(0, 5)" :key="rowIndex">
                <td v-for="(cell, cellIndex) in row" :key="cellIndex" class="px-2 py-1 text-xs text-gray-900">
                  {{ cell }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <p class="text-xs text-gray-500">Showing first 5 rows of {{ csvPreview.totalRows }} total</p>
      </div>

      <!-- Upload Result -->
      <div v-if="uploadResult" class="mt-2 p-2 rounded" :class="uploadResult.success ? 'bg-green-50' : 'bg-red-50'">
        <p :class="uploadResult.success ? 'text-green-800' : 'text-red-800'" class="text-xs">
          {{ uploadResult.message }}
        </p>
        <p v-if="uploadResult.success" class="text-xs text-green-700 mt-0.5">
          Imported: {{ uploadResult.imported }} | Skipped (duplicates): {{ uploadResult.skipped }}
        </p>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-white shadow rounded-lg p-3 mb-3">
      <h2 class="text-sm font-medium text-gray-900 mb-2">Filters</h2>
      
      <div class="grid grid-cols-1 md:grid-cols-4 gap-2">
        <div>
          <label class="block text-xs font-medium text-gray-700 mb-0.5">Start Date</label>
          <input
            v-model="startDate"
            type="date"
            class="w-full px-2 py-1 text-xs border border-gray-300 rounded"
          />
        </div>
        
        <div>
          <label class="block text-xs font-medium text-gray-700 mb-0.5">End Date</label>
          <input
            v-model="endDate"
            type="date"
            class="w-full px-2 py-1 text-xs border border-gray-300 rounded"
          />
        </div>
        
        <div>
          <label class="block text-xs font-medium text-gray-700 mb-0.5">Status</label>
          <select
            v-model="statusFilter"
            class="w-full px-2 py-1 text-xs border border-gray-300 rounded"
          >
            <option value="">All</option>
            <option value="pending_review">Pending Review</option>
            <option value="approved">Approved</option>
            <option value="posted">Posted</option>
            <option value="reconciled">Reconciled</option>
            <option value="duplicate">Duplicate</option>
            <option value="excluded">Excluded</option>
          </select>
        </div>
        
        <div class="flex items-end">
          <button
            @click="fetchData"
            class="w-full px-3 py-1 text-xs bg-gray-600 text-white rounded hover:bg-gray-700"
          >
            Apply Filters
          </button>
        </div>
      </div>
    </div>

    <!-- Stats -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-2 mb-4">
      <div class="bg-white shadow rounded p-2">
        <div class="text-xs text-gray-500">Pending Review</div>
        <div class="text-lg font-bold text-yellow-600">{{ stats.pending }}</div>
      </div>
      <div class="bg-white shadow rounded p-2">
        <div class="text-xs text-gray-500">Approved</div>
        <div class="text-lg font-bold text-green-600">{{ stats.approved }}</div>
      </div>
      <div class="bg-white shadow rounded p-2">
        <div class="text-xs text-gray-500">Duplicate</div>
        <div class="text-lg font-bold text-red-600">{{ stats.duplicate }}</div>
      </div>
      <div class="bg-white shadow rounded p-2">
        <div class="text-xs text-gray-500">Excluded</div>
        <div class="text-lg font-bold text-gray-600">{{ stats.excluded }}</div>
      </div>
    </div>

    <!-- Account Summary -->
    <div class="bg-white shadow rounded p-3 mb-3">
      <div class="flex items-center justify-between mb-2">
        <h2 class="text-sm font-semibold text-gray-900">Summary by Account</h2>
        <button 
          @click="showAccountSummary = !showAccountSummary"
          class="text-xs text-blue-600 hover:text-blue-800"
        >
          {{ showAccountSummary ? 'Hide' : 'Show' }}
        </button>
      </div>
      
      <div v-if="showAccountSummary" class="space-y-1">
        <div v-if="accountSummary.length === 0" class="text-xs text-gray-500 py-2">
          No categorized entries in this period
        </div>
        <div v-else>
          <div class="grid grid-cols-4 gap-2 text-2xs font-semibold text-gray-700 pb-1 border-b">
            <div>Account</div>
            <div class="text-right">Total Debits</div>
            <div class="text-right">Total Credits</div>
            <div class="text-right">Net Effect</div>
          </div>
          <div 
            v-for="summary in accountSummary" 
            :key="summary.account"
            class="grid grid-cols-4 gap-2 text-xs py-1 border-b border-gray-100"
          >
            <div class="font-medium text-gray-900">{{ summary.account }}</div>
            <div class="text-right font-mono tabular-nums text-gray-700">
              {{ summary.totalDebits > 0 ? formatCurrency(summary.totalDebits) : '-' }}
            </div>
            <div class="text-right font-mono tabular-nums text-gray-700">
              {{ summary.totalCredits > 0 ? formatCurrency(summary.totalCredits) : '-' }}
            </div>
            <div 
              class="text-right font-mono tabular-nums font-medium"
              :class="summary.netEffect > 0 ? 'text-green-700' : summary.netEffect < 0 ? 'text-red-700' : 'text-gray-900'"
            >
              {{ formatCurrency(Math.abs(summary.netEffect)) }}
              <span class="text-2xs ml-1">{{ summary.netEffect > 0 ? 'DR' : summary.netEffect < 0 ? 'CR' : '' }}</span>
            </div>
          </div>
          <div class="grid grid-cols-4 gap-2 text-xs font-bold pt-2 border-t-2 border-gray-300">
            <div class="text-gray-900">TOTAL</div>
            <div class="text-right font-mono tabular-nums text-gray-900">{{ formatCurrency(totalDebits) }}</div>
            <div class="text-right font-mono tabular-nums text-gray-900">{{ formatCurrency(totalCredits) }}</div>
            <div 
              class="text-right font-mono tabular-nums"
              :class="isBalanced ? 'text-green-700' : 'text-red-700'"
            >
              {{ isBalanced ? '✓ Balanced' : '⚠ Unbalanced' }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Bulk Actions -->
    <div v-if="selectedJournals.length > 0" class="bg-blue-50 border border-blue-200 rounded p-2 mb-3">
      <div class="flex items-center justify-between">
        <span class="text-xs text-blue-900 font-medium">
          {{ selectedJournals.length }} selected
        </span>
        <div class="space-x-1">
          <button
            v-if="hasSelectedPendingEntries"
            @click="bulkUpdate('approved')"
            class="px-2 py-1 bg-green-600 text-white text-xs rounded hover:bg-green-700"
          >
            <i class="fas fa-check mr-1"></i>
            Approve
          </button>
          <button
            v-if="hasSelectedApprovedEntries"
            @click="bulkPostToGL"
            class="px-2 py-1 bg-blue-600 text-white text-xs rounded hover:bg-blue-700"
          >
            <i class="fas fa-arrow-right mr-1"></i>
            Post to GL
          </button>
          <button
            v-if="hasSelectedDuplicateOrExcludedEntries"
            @click="bulkUpdate('pending_review')"
            class="px-2 py-1 bg-sky-600 text-white text-xs rounded hover:bg-sky-700"
          >
            <i class="fas fa-undo mr-1"></i>
            Reset to Pending
          </button>
          <button
            @click="bulkUpdate('duplicate')"
            class="px-2 py-1 bg-orange-600 text-white text-xs rounded hover:bg-orange-700"
          >
            <i class="fas fa-copy mr-1"></i>
            Duplicate
          </button>
          <button
            @click="bulkUpdate('excluded')"
            class="px-2 py-1 bg-gray-600 text-white text-xs rounded hover:bg-gray-700"
          >
            <i class="fas fa-ban mr-1"></i>
            Exclude
          </button>
          <button
            @click="clearSelection"
            class="px-2 py-1 bg-white text-gray-700 text-xs rounded border border-gray-300 hover:bg-gray-50"
          >
            Clear
          </button>
        </div>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white shadow rounded-lg overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full table-fixed divide-y divide-gray-200">
          <colgroup>
            <col class="w-8"> <!-- Checkbox -->
            <col class="w-20"> <!-- Date -->
            <col class="w-32"> <!-- Account -->
            <col class="w-28"> <!-- Sub-Account -->
            <col> <!-- Description (flexible) -->
            <col class="w-24"> <!-- Debit -->
            <col class="w-24"> <!-- Credit -->
            <col class="w-20"> <!-- Status -->
            <col class="w-16"> <!-- Actions -->
          </colgroup>
          <thead class="bg-gray-50">
            <tr>
              <th class="px-2 py-1 text-left">
                <input
                  type="checkbox"
                  v-model="selectAllCheckbox"
                  @change="toggleSelectAll"
                  class="rounded"
                />
              </th>
              <th class="px-2 py-1 text-left text-xs font-semibold text-gray-700 uppercase">
                Date
              </th>
              <th class="px-2 py-1 text-left text-xs font-semibold text-gray-700 uppercase">
                Account
              </th>
              <th class="px-2 py-1 text-left text-xs font-semibold text-gray-700 uppercase">
                Sub-Account
              </th>
              <th class="px-2 py-1 text-left text-xs font-semibold text-gray-700 uppercase">
                Description
              </th>
              <th class="px-1.5 py-1 text-right text-xs font-semibold text-gray-700 uppercase">
                Debit
              </th>
              <th class="px-1.5 py-1 text-right text-xs font-semibold text-gray-700 uppercase">
                Credit
              </th>
              <th class="px-1 py-1 text-center text-xs font-semibold text-gray-700 uppercase">
                Status
              </th>
              <th class="px-1 py-1 text-center text-xs font-semibold text-gray-700 uppercase">
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-if="isLoading">
              <td colspan="10" class="px-6 py-4 text-center text-gray-500">
                Loading...
              </td>
            </tr>
            <tr v-else-if="filteredJournals.length === 0">
              <td colspan="10" class="px-6 py-4 text-center text-gray-500">
                No offline journals found
              </td>
            </tr>
            <tr
              v-else
              v-for="journal in filteredJournals"
              :key="journal.ID"
              :class="[
                'hover:bg-gray-50',
                journal.account === 'UNCLASSIFIED' ? 'cursor-pointer' : ''
              ]"
              @click="journal.account === 'UNCLASSIFIED' ? openCategorizeModal(journal) : null"
            >
              <td class="px-2 py-1">
                <input
                  type="checkbox"
                  :checked="isSelected(journal.ID)"
                  @click.stop="toggleSelect(journal.ID)"
                  class="rounded"
                />
              </td>
              <td class="px-2 py-1 text-xs text-gray-900 truncate" :title="formatDate(journal.date)">
                {{ formatDate(journal.date) }}
              </td>
              <td class="px-2 py-1 text-xs font-medium text-gray-900 truncate" :title="journal.account">
                {{ journal.account }}
              </td>
              <td class="px-2 py-1 text-xs text-gray-700 truncate" :title="journal.sub_account || '-'">
                {{ journal.sub_account || '-' }}
              </td>
              <td class="px-2 py-1 text-xs text-gray-600">
                <div class="flex items-center gap-1.5 min-w-0">
                  <button 
                    @click.stop="openNotePopover(journal, $event)"
                    :class="[
                      'shrink-0',
                      journal.notes ? 'text-amber-600 hover:text-amber-800' : 'text-gray-400 hover:text-gray-600'
                    ]"
                    :title="journal.notes || 'Add note'"
                  >
                    <i :class="['fas', journal.notes ? 'fa-sticky-note' : 'fa-plus-circle', 'text-xs']"></i>
                  </button>
                  <span 
                    class="truncate" 
                    :title="journal.description"
                  >
                    {{ journal.description }}
                  </span>
                </div>
              </td>
              <td class="px-1.5 py-1 text-xs text-right text-gray-900 font-mono tabular-nums truncate">
                {{ journal.debit > 0 ? formatCurrency(journal.debit) : '' }}
              </td>
              <td class="px-1.5 py-1 text-xs text-right text-gray-900 font-mono tabular-nums truncate">
                {{ journal.credit > 0 ? formatCurrency(journal.credit) : '' }}
              </td>
              <td class="px-1 py-1 text-center truncate">
                <span :class="['px-1.5 py-0.5 text-[10px] font-semibold rounded whitespace-nowrap', getStatusColor(journal.status)]">
                  {{ formatStatus(journal.status) }}
                </span>
              </td>
              <td class="px-1 py-1 text-center">
                <div class="flex items-center justify-center gap-1">
                  <!-- Approve button for pending entries -->
                  <button
                    v-if="journal.status === 'pending_review'"
                    @click.stop="updateStatus(journal.ID, 'approved')"
                    class="text-green-600 hover:text-green-900"
                    title="Approve"
                  >
                    <i class="fas fa-check text-xs"></i>
                  </button>
                  <!-- Match to Expense button for approved payment entries -->
                  <button
                    v-if="journal.status === 'approved' && !journal.reconciled_expense_id && (journal.debit > 0 || journal.credit > 0)"
                    @click.stop="openReconcileModal(journal)"
                    class="text-purple-600 hover:text-purple-900"
                    title="Match to Expense"
                  >
                    <i class="fas fa-link text-xs"></i>
                  </button>
                  <!-- Reconciled indicator -->
                  <span
                    v-if="journal.reconciled_expense_id"
                    class="text-purple-600"
                    title="Reconciled with expense"
                  >
                    <i class="fas fa-check-circle text-xs"></i>
                  </span>
                  <!-- Reset button for duplicate/excluded entries -->
                  <button
                    v-if="journal.status === 'duplicate' || journal.status === 'excluded'"
                    @click.stop="updateStatus(journal.ID, 'pending_review')"
                    class="text-blue-600 hover:text-blue-900"
                    title="Reset to Pending Review"
                  >
                    <i class="fas fa-undo text-xs"></i>
                  </button>
                  <!-- Edit button (not for posted entries) -->
                  <button
                    v-if="journal.status !== 'posted' && journal.account !== 'UNCLASSIFIED'"
                    @click.stop="openEditModal(journal)"
                    class="text-sky-600 hover:text-sky-900"
                    title="Edit Entry"
                  >
                    <i class="fas fa-edit text-xs"></i>
                  </button>
                  <!-- Delete button (not for posted entries) -->
                  <button
                    v-if="journal.status !== 'posted'"
                    @click.stop="deleteOfflineJournal(journal.ID)"
                    class="text-red-600 hover:text-red-900"
                    title="Delete"
                  >
                    <i class="fas fa-trash-alt text-xs"></i>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Note Flyout Popover -->
    <Teleport to="body">
      <div
        v-if="notePopoverOpen"
        @click="closeNotePopover"
        class="pointer-events-none fixed inset-0 z-50"
      >
        <div
          @click.stop
          :style="notePopoverStyle"
          class="pointer-events-auto absolute w-72 transform rounded bg-white shadow-lg ring-1 ring-black ring-opacity-5"
        >
          <div class="p-2">
            <div class="flex items-start gap-2">
              <div class="shrink-0">
                <i class="fas fa-sticky-note text-amber-500 text-xs"></i>
              </div>
              <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between mb-1">
                  <p class="text-2xs font-medium text-gray-900">Note</p>
                  <button
                    @click="closeNotePopover"
                    class="inline-flex rounded text-gray-400 hover:text-gray-500 focus:outline-none"
                  >
                    <span class="sr-only">Close</span>
                    <i class="fas fa-times text-2xs"></i>
                  </button>
                </div>
                <textarea
                  ref="noteTextarea"
                  v-model="notePopoverText"
                  placeholder="Add a note..."
                  class="w-full px-1.5 py-1 text-2xs border border-gray-200 rounded focus:outline-none focus:ring-1 focus:ring-amber-500 focus:border-transparent resize-none"
                  rows="2"
                  @keydown.escape="closeNotePopover"
                ></textarea>
                <div class="flex justify-end gap-1 mt-1.5">
                  <button
                    @click="closeNotePopover"
                    class="px-1.5 py-0.5 text-2xs text-gray-700 hover:text-gray-900 focus:outline-none"
                  >
                    Cancel
                  </button>
                  <button
                    @click="saveNoteFromPopover"
                    class="px-2 py-0.5 text-2xs bg-amber-600 text-white rounded hover:bg-amber-700 focus:outline-none"
                  >
                    Save
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Confirm Post to GL Modal -->
    <div
      v-if="confirmPostModalOpen"
      @click="closeConfirmPostModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-gray-500/75"
    >
      <div
        @click.stop
        class="bg-white rounded-lg shadow-xl w-full max-w-md mx-4"
      >
        <div class="bg-blue-600 px-4 py-3 rounded-t-lg">
          <h3 class="text-sm font-semibold text-white">Confirm Post to General Ledger</h3>
        </div>
        <div class="p-4">
          <div class="flex items-start gap-3 mb-4">
            <div class="flex-shrink-0">
              <i class="fas fa-exclamation-triangle text-yellow-500 text-2xl"></i>
            </div>
            <div>
              <p class="text-sm text-gray-900 font-medium mb-2">
                Post {{ entriesToPost.length }} {{ entriesToPost.length === 1 ? 'entry' : 'entries' }} to the General Ledger?
              </p>
              <p class="text-xs text-gray-600">
                Once posted, these entries will be part of your official accounting records and cannot be edited. 
                Any corrections will require reversing journal entries.
              </p>
            </div>
          </div>
          <div class="flex justify-end gap-2">
            <button
              @click="closeConfirmPostModal"
              class="px-3 py-1.5 text-xs text-gray-700 border border-gray-300 rounded hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              @click="confirmPostToGL"
              class="px-3 py-1.5 text-xs bg-blue-600 text-white rounded hover:bg-blue-700"
            >
              <i class="fas fa-arrow-right mr-1"></i>
              Post to GL
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Reconcile to Expense Modal -->
    <div
      v-if="reconcileModalOpen"
      @click="closeReconcileModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-gray-500/75"
    >
      <div
        @click.stop
        class="bg-white rounded-lg shadow-xl w-full max-w-3xl mx-4"
      >
        <div class="bg-purple-600 px-4 py-3 rounded-t-lg flex items-center justify-between">
          <h3 class="text-sm font-semibold text-white">
            <i class="fas fa-link mr-2"></i>
            Match Transaction to Expense
          </h3>
          <button @click="closeReconcileModal" class="text-white hover:text-gray-200">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="p-4">
          <!-- Transaction Info -->
          <div v-if="selectedJournalForReconcile" class="bg-gray-50 rounded p-3 mb-4">
            <div class="text-xs text-gray-700 mb-2 font-semibold">Transaction Details:</div>
            <div class="grid grid-cols-3 gap-3 text-xs">
              <div>
                <span class="text-gray-600">Date:</span>
                <span class="ml-1 font-medium">{{ formatDate(selectedJournalForReconcile.date) }}</span>
              </div>
              <div>
                <span class="text-gray-600">Account:</span>
                <span class="ml-1 font-medium">{{ selectedJournalForReconcile.account }}</span>
              </div>
              <div>
                <span class="text-gray-600">Amount:</span>
                <span class="ml-1 font-medium font-mono">{{ formatCurrency(selectedJournalForReconcile.debit || selectedJournalForReconcile.credit) }}</span>
              </div>
            </div>
            <div class="mt-2 text-xs">
              <span class="text-gray-600">Description:</span>
              <span class="ml-1">{{ selectedJournalForReconcile.description }}</span>
            </div>
          </div>

          <!-- Search for Expenses -->
          <div class="mb-4">
            <label class="block text-xs font-medium text-gray-700 mb-1">Search for matching expense:</label>
            <input
              v-model="expenseSearchQuery"
              @input="searchExpensesForReconcile"
              type="text"
              placeholder="Search by description..."
              class="w-full px-3 py-2 text-xs border border-gray-300 rounded focus:ring-purple-500 focus:border-purple-500"
            />
          </div>

          <!-- Matching Expenses -->
          <div v-if="matchingExpenses.length > 0" class="border rounded max-h-80 overflow-y-auto">
            <table class="w-full text-xs">
              <thead class="bg-gray-50 sticky top-0">
                <tr>
                  <th class="px-2 py-1.5 text-left text-xs font-semibold text-gray-700">Date</th>
                  <th class="px-2 py-1.5 text-left text-xs font-semibold text-gray-700">Project</th>
                  <th class="px-2 py-1.5 text-left text-xs font-semibold text-gray-700">Description</th>
                  <th class="px-2 py-1.5 text-right text-xs font-semibold text-gray-700">Amount</th>
                  <th class="px-2 py-1.5 text-left text-xs font-semibold text-gray-700">Category</th>
                  <th class="px-2 py-1.5 text-center text-xs font-semibold text-gray-700">Action</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200">
                <tr v-for="expense in matchingExpenses" :key="expense.ID" class="hover:bg-gray-50">
                  <td class="px-2 py-1.5 text-gray-900">{{ formatDate(expense.date) }}</td>
                  <td class="px-2 py-1.5 text-gray-700">{{ expense.project?.name || 'Internal' }}</td>
                  <td class="px-2 py-1.5 text-gray-700">{{ expense.description }}</td>
                  <td class="px-2 py-1.5 text-right font-mono text-gray-900">{{ formatCurrency(expense.amount) }}</td>
                  <td class="px-2 py-1.5 text-gray-700">{{ expense.category?.name }}</td>
                  <td class="px-2 py-1.5 text-center">
                    <button
                      @click="reconcileExpense(expense.ID)"
                      class="px-2 py-1 bg-purple-600 text-white text-xs rounded hover:bg-purple-700"
                    >
                      <i class="fas fa-check mr-1"></i>
                      Match
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-else-if="expenseSearchQuery" class="text-xs text-gray-500 text-center py-4">
            No matching expenses found
          </div>
          <div v-else class="text-xs text-gray-500 text-center py-4">
            Start typing to search for expenses
          </div>
        </div>
      </div>
    </div>

    <!-- Edit Offline Journal Modal -->
    <div
      v-if="editModalOpen"
      @click="closeEditModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-gray-500/75"
    >
      <div
        @click.stop
        class="bg-white rounded-lg shadow-xl w-full max-w-lg mx-4"
      >
        <div class="bg-sage px-4 py-3 rounded-t-lg">
          <h3 class="text-sm font-semibold text-white">Edit Journal Entry</h3>
        </div>
        <div class="p-4 space-y-3">
          <div>
            <label class="block text-xs text-gray-700 mb-1">Account</label>
            <select v-model="editForm.account" class="w-full px-2 py-1.5 text-xs border rounded focus:ring-1 focus:ring-sage">
              <option value="">Select account...</option>
              <option v-for="account in availableAccounts" :key="account.account_code" :value="account.account_code">
                {{ account.account_name }}
              </option>
            </select>
          </div>
          <div>
            <label class="block text-xs text-gray-700 mb-1">Subaccount</label>
            <select v-model="editForm.subaccount" class="w-full px-2 py-1.5 text-xs border rounded focus:ring-1 focus:ring-sage">
              <option value="">None</option>
              <option v-for="sub in getSubaccountsForAccount(editForm.account)" :key="sub.code" :value="sub.code">
                {{ sub.name }}
              </option>
            </select>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs text-gray-700 mb-1">Debit</label>
              <input 
                v-model.number="editForm.debit" 
                type="number" 
                step="0.01"
                class="w-full px-2 py-1.5 text-xs border rounded focus:ring-1 focus:ring-sage"
              />
            </div>
            <div>
              <label class="block text-xs text-gray-700 mb-1">Credit</label>
              <input 
                v-model.number="editForm.credit" 
                type="number" 
                step="0.01"
                class="w-full px-2 py-1.5 text-xs border rounded focus:ring-1 focus:ring-sage"
              />
            </div>
          </div>
          <div class="flex justify-end gap-2 pt-2">
            <button
              @click="closeEditModal"
              class="px-3 py-1.5 text-xs text-gray-700 border border-gray-300 rounded hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              @click="saveEdit"
              class="px-3 py-1.5 text-xs bg-sage text-white rounded hover:bg-sage-dark"
            >
              Save Changes
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Categorize Modal -->
    <div
      v-if="categorizeModalOpen"
      @click="closeCategorizeModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-gray-500/75"
    >
      <div
        @click.stop
        class="bg-white rounded-lg shadow-xl w-full max-w-2xl mx-4"
      >
        <div class="bg-sage px-4 py-3 rounded-t-lg">
          <h3 class="text-sm font-semibold text-white">Categorize Transaction</h3>
        </div>
        <div class="p-4">
          <div class="mb-3 bg-gray-50 p-3 rounded">
            <p class="text-xs text-gray-900 font-medium mb-1">{{ categorizeJournal?.description }}</p>
            <div class="flex items-center gap-4 text-2xs text-gray-600">
              <span><strong>Date:</strong> {{ categorizeJournal?.date?.split('T')[0] }}</span>
              <span><strong>Amount:</strong> {{ formatCurrency((categorizeJournal?.debit || 0) + (categorizeJournal?.credit || 0)) }}</span>
            </div>
            <p v-if="pairedEntry" class="text-2xs text-green-600 mt-1">
              <i class="fas fa-check-circle"></i> Paired entry found - categorizing both sides
            </p>
            <p v-else class="text-2xs text-orange-600 mt-1">
              <i class="fas fa-exclamation-triangle"></i> Warning: No paired entry found
            </p>
          </div>

          <div class="grid grid-cols-2 gap-4 mb-4">
            <!-- Debit Side -->
            <div class="border-2 border-green-200 bg-green-50 rounded p-3">
              <div class="flex items-center justify-between mb-2">
                <h4 class="text-xs font-semibold text-green-900">Debit (DR)</h4>
                <span class="text-xs font-mono text-green-700">{{ formatCurrency(categorizeJournal?.debit || pairedEntry?.debit || 0) }}</span>
              </div>
              <div class="text-2xs text-gray-600 mb-2 italic">
                {{ categorizeJournal?.debit ? 'Current entry' : 'Paired entry' }}
              </div>
              <div class="space-y-2">
                <div>
                  <label class="block text-2xs text-gray-700 mb-1">Account *</label>
                  <select v-model="categorizeForm.debitAccount" class="w-full px-2 py-1 text-xs border rounded focus:ring-1 focus:ring-sage">
                    <option value="">Select account...</option>
                    <option v-for="account in availableAccounts" :key="account.account_code" :value="account.account_code">
                      {{ account.account_name }}
                    </option>
                  </select>
                </div>
                <div>
                  <label class="block text-2xs text-gray-700 mb-1">Subaccount</label>
                  <select v-model="categorizeForm.debitSubaccount" class="w-full px-2 py-1 text-xs border rounded focus:ring-1 focus:ring-sage">
                    <option value="">None</option>
                    <option v-for="sub in getSubaccountsForAccount(categorizeForm.debitAccount)" :key="sub.code" :value="sub.code">
                      {{ sub.name }}
                    </option>
                  </select>
                </div>
              </div>
            </div>

            <!-- Credit Side -->
            <div class="border-2 border-blue-200 bg-blue-50 rounded p-3">
              <div class="flex items-center justify-between mb-2">
                <h4 class="text-xs font-semibold text-blue-900">Credit (CR)</h4>
                <span class="text-xs font-mono text-blue-700">{{ formatCurrency(categorizeJournal?.credit || pairedEntry?.credit || 0) }}</span>
              </div>
              <div class="text-2xs text-gray-600 mb-2 italic">
                {{ categorizeJournal?.credit ? 'Current entry' : 'Paired entry' }}
              </div>
              <div class="space-y-2">
                <div>
                  <label class="block text-2xs text-gray-700 mb-1">Account *</label>
                  <select v-model="categorizeForm.creditAccount" class="w-full px-2 py-1 text-xs border rounded focus:ring-1 focus:ring-sage">
                    <option value="">Select account...</option>
                    <option v-for="account in availableAccounts" :key="account.account_code" :value="account.account_code">
                      {{ account.account_name }}
                    </option>
                  </select>
                </div>
                <div>
                  <label class="block text-2xs text-gray-700 mb-1">Subaccount</label>
                  <select v-model="categorizeForm.creditSubaccount" class="w-full px-2 py-1 text-xs border rounded focus:ring-1 focus:ring-sage">
                    <option value="">None</option>
                    <option v-for="sub in getSubaccountsForAccount(categorizeForm.creditAccount)" :key="sub.code" :value="sub.code">
                      {{ sub.name }}
                    </option>
                  </select>
                </div>
              </div>
            </div>
          </div>

          <div class="flex justify-end gap-2">
            <button
              @click="closeCategorizeModal"
              class="px-3 py-1.5 text-xs text-gray-700 border border-gray-300 rounded hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              @click="saveCategorization"
              :disabled="!categorizeForm.debitAccount || !categorizeForm.creditAccount"
              class="px-3 py-1.5 text-xs bg-sage text-white rounded hover:bg-sage-dark disabled:bg-gray-300 disabled:cursor-not-allowed"
            >
              Save & Categorize
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import type { OfflineJournal } from '../../types/OfflineJournal';
import { formatCurrency, getStatusColor } from '../../types/OfflineJournal';
import {
  uploadCSVFile,
  getOfflineJournals,
  updateOfflineJournalStatus,
  bulkUpdateOfflineJournalStatus,
  categorizeCSVTransaction,
  editOfflineJournal,
  deleteOfflineJournal as deleteOfflineJournalAPI,
  postOfflineJournalsToGL,
} from '../../api/offlineJournals';
import { getChartOfAccounts } from '../../api/chartOfAccounts';
import { getSubaccounts } from '../../api/subaccounts';
import { searchExpensesForReconciliation, reconcileExpenseWithOfflineJournal } from '../../api/reconciliation';
import type { ChartOfAccount } from '../../types/ChartOfAccount';
import type { Subaccount } from '../../types/Subaccount';

const journals = ref<OfflineJournal[]>([]);
const isLoading = ref(false);
const startDate = ref('');
const endDate = ref('');
const statusFilter = ref('');
const selectedJournals = ref<number[]>([]);
const selectAllCheckbox = ref(false);
const showAccountSummary = ref(true);

// Notes editing - flyout popover
const notePopoverOpen = ref(false);
const notePopoverJournal = ref<OfflineJournal | null>(null);
const notePopoverText = ref('');
const notePopoverStyle = ref({});
const noteTextarea = ref<HTMLTextAreaElement | null>(null);

// Confirm post to GL modal
const confirmPostModalOpen = ref(false);
const entriesToPost = ref<number[]>([]);

// Edit modal for offline journals
const editModalOpen = ref(false);
const editingJournal = ref<OfflineJournal | null>(null);
const editForm = ref({
  account: '',
  subaccount: '',
  debit: 0,
  credit: 0,
});

// Reconcile to Expense modal
const reconcileModalOpen = ref(false);
const selectedJournalForReconcile = ref<OfflineJournal | null>(null);
const expenseSearchQuery = ref('');
const matchingExpenses = ref<any[]>([]);

// Categorization modal
const categorizeModalOpen = ref(false);
const categorizeJournal = ref<OfflineJournal | null>(null);
const categorizeForm = ref({
  debitAccount: '',
  debitSubaccount: '',
  creditAccount: '',
  creditSubaccount: '',
});
const availableAccounts = ref<ChartOfAccount[]>([]);
const availableSubaccounts = ref<Subaccount[]>([]);

// Computed property to find the paired entry
const pairedEntry = computed(() => {
  if (!categorizeJournal.value) return null;
  return journals.value.find(j => 
    j.ID !== categorizeJournal.value!.ID &&
    j.date === categorizeJournal.value!.date &&
    j.description === categorizeJournal.value!.description &&
    ((j.debit > 0 && categorizeJournal.value!.credit > 0) || 
     (j.credit > 0 && categorizeJournal.value!.debit > 0))
  );
});

// CSV upload
const csvFileInput = ref<HTMLInputElement | null>(null);
const selectedCSVFile = ref<File | null>(null);
const csvPreview = ref<{
  headers: string[];
  rows: string[][];
  totalRows: number;
} | null>(null);
const csvOptions = ref({
  dateCol: 0,
  descCol: 1,
  amountCol: 2,
  hasHeader: true,
  dateFormat: '',
});
const isUploading = ref(false);
const uploadResult = ref<{ success: boolean; message: string; imported?: number; skipped?: number } | null>(null);

// Reconcile to Expense functions
async function openReconcileModal(journal: OfflineJournal) {
  selectedJournalForReconcile.value = journal;
  expenseSearchQuery.value = '';
  reconcileModalOpen.value = true;
  
  // Auto-search for matching expenses by amount and date
  await searchExpensesForReconcile();
}

function closeReconcileModal() {
  reconcileModalOpen.value = false;
  selectedJournalForReconcile.value = null;
  expenseSearchQuery.value = '';
  matchingExpenses.value = [];
}

async function searchExpensesForReconcile() {
  if (!selectedJournalForReconcile.value) {
    matchingExpenses.value = [];
    return;
  }

  try {
    const amount = selectedJournalForReconcile.value.debit || selectedJournalForReconcile.value.credit;
    const date = formatDate(selectedJournalForReconcile.value.date);
    
    const results = await searchExpensesForReconciliation({
      query: expenseSearchQuery.value || undefined,
      date: date,
      amount: amount, // Already in cents
    });
    
    matchingExpenses.value = results;
  } catch (error) {
    console.error('Failed to search expenses:', error);
    matchingExpenses.value = [];
  }
}

async function reconcileExpense(expenseID: number) {
  if (!selectedJournalForReconcile.value) return;

  try {
    await reconcileExpenseWithOfflineJournal(expenseID, selectedJournalForReconcile.value.ID);
    closeReconcileModal();
    await fetchData();
  } catch (error) {
    console.error('Failed to reconcile expense:', error);
  }
}

// Initialize date range to current month
onMounted(() => {
  const now = new Date();
  const firstDay = new Date(now.getFullYear(), now.getMonth(), 1);
  const lastDay = new Date(now.getFullYear(), now.getMonth() + 1, 0);
  
  startDate.value = firstDay.toISOString().split('T')[0];
  endDate.value = lastDay.toISOString().split('T')[0];
  
  fetchData();
  loadAccountsAndSubaccounts();
});

// Fetch offline journals
async function fetchData() {
  isLoading.value = true;
  try {
    const params: any = {};
    if (startDate.value) params.start_date = startDate.value;
    if (endDate.value) params.end_date = endDate.value;
    if (statusFilter.value) params.status = statusFilter.value;
    
    journals.value = await getOfflineJournals(params);
  } catch (error) {
    console.error('Failed to fetch offline journals:', error);
  } finally {
    isLoading.value = false;
  }
}

// CSV file upload handlers
// Parse a CSV line properly handling quoted fields
function parseCSVLine(line: string): string[] {
  const result: string[] = [];
  let current = '';
  let inQuotes = false;
  
  for (let i = 0; i < line.length; i++) {
    const char = line[i];
    
    if (char === '"') {
      // Handle escaped quotes ("")
      if (inQuotes && line[i + 1] === '"') {
        current += '"';
        i++; // Skip next quote
      } else {
        inQuotes = !inQuotes;
      }
    } else if (char === ',' && !inQuotes) {
      // End of field
      result.push(current.trim());
      current = '';
    } else {
      current += char;
    }
  }
  
  // Add last field
  result.push(current.trim());
  
  return result;
}

async function handleCSVFileSelect(event: Event) {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  
  if (!file) return;
  
  selectedCSVFile.value = file;
  uploadResult.value = null;
  
  // Parse CSV for preview
  try {
    const text = await file.text();
    const lines = text.split('\n').filter(line => line.trim());
    
    if (lines.length === 0) {
      uploadResult.value = {
        success: false,
        message: 'CSV file is empty',
      };
      return;
    }
    
    // Parse first line as headers using proper CSV parsing
    const headers = parseCSVLine(lines[0]);
    
    // Parse remaining lines as data (show first 5 for preview)
    const rows: string[][] = [];
    for (let i = 1; i < Math.min(lines.length, 6); i++) {
      rows.push(parseCSVLine(lines[i]));
    }
    
    csvPreview.value = {
      headers,
      rows,
      totalRows: lines.length - 1,
    };
    
    // Auto-detect likely columns
    headers.forEach((header, index) => {
      const headerLower = header.toLowerCase();
      if (headerLower.includes('date') || headerLower.includes('trans')) {
        csvOptions.value.dateCol = index;
      }
      if (headerLower.includes('description') || headerLower.includes('desc') || headerLower.includes('memo') || headerLower.includes('name')) {
        csvOptions.value.descCol = index;
      }
      if (headerLower.includes('amount') || headerLower.includes('price') || headerLower.includes('debit') || headerLower.includes('credit')) {
        csvOptions.value.amountCol = index;
      }
    });
    
    console.log('Parsed CSV:', { 
      headerCount: headers.length, 
      rowCount: rows.length, 
      headers,
      detectedColumns: { 
        date: csvOptions.value.dateCol, 
        desc: csvOptions.value.descCol, 
        amount: csvOptions.value.amountCol 
      }
    });
    
  } catch (error: any) {
    console.error('CSV parse error:', error);
    uploadResult.value = {
      success: false,
      message: 'Failed to parse CSV: ' + error.message,
    };
  }
}

function cancelCSVPreview() {
  csvPreview.value = null;
  selectedCSVFile.value = null;
  if (csvFileInput.value) csvFileInput.value.value = '';
  uploadResult.value = null;
}

async function uploadCSV() {
  if (!selectedCSVFile.value) return;
  
  isUploading.value = true;
  uploadResult.value = null;
  
  try {
    const result = await uploadCSVFile(selectedCSVFile.value, {
      dateCol: csvOptions.value.dateCol,
      descCol: csvOptions.value.descCol,
      amountCol: csvOptions.value.amountCol,
      hasHeader: csvOptions.value.hasHeader,
      dateFormat: csvOptions.value.dateFormat || undefined,
    });
    
    uploadResult.value = {
      success: true,
      message: result.message,
      imported: result.imported,
      skipped: result.skipped,
    };
    
    // Refresh data
    await fetchData();
    
    // Clear preview and file input
    csvPreview.value = null;
    selectedCSVFile.value = null;
    if (csvFileInput.value) csvFileInput.value.value = '';
  } catch (error: any) {
    uploadResult.value = {
      success: false,
      message: error.response?.data || 'CSV upload failed',
    };
  } finally {
    isUploading.value = false;
  }
}

// Notes popover functions
function openNotePopover(journal: OfflineJournal, event: MouseEvent) {
  notePopoverJournal.value = journal;
  notePopoverText.value = journal.notes || '';
  
  // Calculate position to appear near the clicked icon
  const button = event.currentTarget as HTMLElement;
  const rect = button.getBoundingClientRect();
  
  // Position to the right of the button, or to the left if not enough space
  const popoverWidth = 288; // w-72 = 18rem = 288px
  const popoverHeight = 130; // Approximate height with smaller padding
  const margin = 8;
  
  const spaceOnRight = window.innerWidth - rect.right;
  const spaceOnLeft = rect.left;
  
  let left: number;
  if (spaceOnRight > popoverWidth + margin) {
    // Position to the right
    left = rect.right + margin;
  } else if (spaceOnLeft > popoverWidth + margin) {
    // Position to the left
    left = rect.left - popoverWidth - margin;
  } else {
    // Not enough space on either side, position on the side with more space
    // but ensure it stays within bounds
    if (spaceOnRight > spaceOnLeft) {
      left = Math.max(margin, window.innerWidth - popoverWidth - margin);
    } else {
      left = margin;
    }
  }
  
  // Position vertically to align with the button, but keep on screen
  let top = rect.top;
  if (top + popoverHeight > window.innerHeight - margin) {
    top = Math.max(margin, window.innerHeight - popoverHeight - margin);
  }
  if (top < margin) {
    top = margin;
  }
  
  notePopoverStyle.value = {
    left: `${left}px`,
    top: `${top}px`,
  };
  
  notePopoverOpen.value = true;
  
  // Auto-focus the textarea after opening
  setTimeout(() => {
    noteTextarea.value?.focus();
  }, 50);
}

function closeNotePopover() {
  notePopoverOpen.value = false;
  notePopoverJournal.value = null;
  notePopoverText.value = '';
}

async function saveNoteFromPopover() {
  if (!notePopoverJournal.value) return;
  
  try {
    const journal = notePopoverJournal.value;
    
    // Update the note via API (keeping current status)
    await updateOfflineJournalStatus(journal.ID, journal.status, notePopoverText.value);
    
    // Update local state
    journal.notes = notePopoverText.value;
    
    // Close popover
    closeNotePopover();
  } catch (error) {
    console.error('Failed to save note:', error);
    alert('Failed to save note');
  }
}

// Categorization functions
async function loadAccountsAndSubaccounts() {
  try {
    const [accounts, subaccounts] = await Promise.all([
      getChartOfAccounts(),
      getSubaccounts(),
    ]);
    availableAccounts.value = accounts.filter((a: ChartOfAccount) => a.is_active);
    availableSubaccounts.value = subaccounts.filter((s: Subaccount) => s.is_active);
  } catch (error) {
    console.error('Failed to load accounts:', error);
  }
}

function getSubaccountsForAccount(accountCode: string) {
  if (!accountCode) return [];
  return availableSubaccounts.value.filter(s => s.account_code === accountCode);
}

function openCategorizeModal(journal: OfflineJournal) {
  categorizeJournal.value = journal;
  categorizeForm.value = {
    debitAccount: '',
    debitSubaccount: '',
    creditAccount: '',
    creditSubaccount: '',
  };
  categorizeModalOpen.value = true;
}

function closeCategorizeModal() {
  categorizeModalOpen.value = false;
  categorizeJournal.value = null;
}

async function saveCategorization() {
  if (!categorizeJournal.value || !pairedEntry.value) {
    alert('Could not find paired transaction entry');
    return;
  }
  
  try {
    // Extract just the date portion (YYYY-MM-DD) from the full timestamp
    const dateOnly = categorizeJournal.value.date.split('T')[0];
    
    // We always send debit account as "from" and credit account as "to"
    await categorizeCSVTransaction({
      date: dateOnly,
      description: categorizeJournal.value.description,
      from_account: categorizeForm.value.debitAccount,
      from_subaccount: categorizeForm.value.debitSubaccount,
      to_account: categorizeForm.value.creditAccount,
      to_subaccount: categorizeForm.value.creditSubaccount,
    });

    closeCategorizeModal();
    await fetchData();
  } catch (error) {
    console.error('Failed to categorize transaction:', error);
    alert('Failed to categorize transaction');
  }
}

// Edit modal functions
function openEditModal(journal: OfflineJournal) {
  editingJournal.value = journal;
  editForm.value = {
    account: journal.account,
    subaccount: journal.sub_account || '',
    debit: journal.debit / 100, // Convert cents to dollars
    credit: journal.credit / 100,
  };
  editModalOpen.value = true;
}

function closeEditModal() {
  editModalOpen.value = false;
  editingJournal.value = null;
}

async function saveEdit() {
  if (!editingJournal.value) return;
  
  try {
    await editOfflineJournal(editingJournal.value.ID, {
      account: editForm.value.account,
      sub_account: editForm.value.subaccount,
      debit: editForm.value.debit,
      credit: editForm.value.credit,
    });
    
    closeEditModal();
    await fetchData();
  } catch (error) {
    console.error('Failed to save edit:', error);
    alert('Failed to save changes');
  }
}

async function deleteOfflineJournal(id: number) {
  try {
    await deleteOfflineJournalAPI(id);
    await fetchData();
  } catch (error) {
    console.error('Failed to delete entry:', error);
    alert('Failed to delete entry');
  }
}

// Update single journal status
async function updateStatus(id: number, status: string) {
  console.log('Updating status:', id, status);
  try {
    const journal = journals.value.find(j => j.ID === id);
    const notes = journal?.notes || '';
    await updateOfflineJournalStatus(id, status, notes);
    console.log('Status updated successfully');
    await fetchData();
  } catch (error) {
    console.error('Failed to update status:', error);
    alert('Failed to update status: ' + (error as any).message);
  }
}

// Bulk update
async function bulkUpdate(status: string) {
  if (selectedJournals.value.length === 0) return;
  
  console.log('Bulk updating:', selectedJournals.value, 'to status:', status);
  try {
    await bulkUpdateOfflineJournalStatus(selectedJournals.value, status);
    console.log('Bulk update successful');
    selectedJournals.value = [];
    await fetchData();
  } catch (error) {
    console.error('Failed to bulk update:', error);
    alert('Failed to bulk update: ' + (error as any).message);
  }
}

// Bulk post to GL - open confirmation modal
async function bulkPostToGL() {
  if (selectedJournals.value.length === 0) return;
  
  // Filter to only approved and categorized entries
  const approvedEntries = selectedJournals.value.filter(id => {
    const journal = journals.value.find(j => j.ID === id);
    return journal && journal.status === 'approved' && journal.account !== 'UNCLASSIFIED';
  });
  
  if (approvedEntries.length === 0) {
    return;
  }
  
  entriesToPost.value = approvedEntries;
  confirmPostModalOpen.value = true;
}

// Close confirmation modal
function closeConfirmPostModal() {
  confirmPostModalOpen.value = false;
  entriesToPost.value = [];
}

// Confirm and actually post to GL
async function confirmPostToGL() {
  try {
    await postOfflineJournalsToGL(entriesToPost.value);
    selectedJournals.value = [];
    closeConfirmPostModal();
    await fetchData();
  } catch (error) {
    console.error('Failed to post to GL:', error);
    alert('Failed to post to GL: ' + (error as any).message);
  }
}

// Selection handlers
function toggleSelect(id: number) {
  console.log('Toggle select:', id);
  const index = selectedJournals.value.indexOf(id);
  if (index > -1) {
    selectedJournals.value.splice(index, 1);
  } else {
    selectedJournals.value.push(id);
  }
  console.log('Selected journals:', selectedJournals.value);
}

function toggleSelectAll() {
  if (selectAllCheckbox.value) {
    // Checkbox is now checked, select all
    selectedJournals.value = filteredJournals.value.map(j => j.ID);
  } else {
    // Checkbox is now unchecked, deselect all
    selectedJournals.value = [];
  }
}

function clearSelection() {
  selectedJournals.value = [];
}

function isSelected(id: number): boolean {
  return selectedJournals.value.includes(id);
}

// Computed
const filteredJournals = computed(() => {
  return journals.value;
});

const stats = computed(() => {
  return {
    pending: journals.value.filter(j => j.status === 'pending_review').length,
    approved: journals.value.filter(j => j.status === 'approved').length,
    duplicate: journals.value.filter(j => j.status === 'duplicate').length,
    excluded: journals.value.filter(j => j.status === 'excluded').length,
  };
});

// Check if any selected entries are pending
const hasSelectedPendingEntries = computed(() => {
  return selectedJournals.value.some(id => {
    const journal = journals.value.find(j => j.ID === id);
    return journal && journal.status === 'pending_review';
  });
});

// Check if any selected entries are approved and categorized (ready to post)
const hasSelectedApprovedEntries = computed(() => {
  return selectedJournals.value.some(id => {
    const journal = journals.value.find(j => j.ID === id);
    return journal && journal.status === 'approved' && journal.account !== 'UNCLASSIFIED';
  });
});

// Check if any selected entries are duplicate or excluded (can be reset)
const hasSelectedDuplicateOrExcludedEntries = computed(() => {
  return selectedJournals.value.some(id => {
    const journal = journals.value.find(j => j.ID === id);
    return journal && (journal.status === 'duplicate' || journal.status === 'excluded');
  });
});

// Account summary - aggregate by account
const accountSummary = computed(() => {
  const summary = new Map<string, { totalDebits: number; totalCredits: number }>();
  
  // Only include categorized entries (not UNCLASSIFIED, not posted)
  const categorizedJournals = journals.value.filter(j => 
    j.account !== 'UNCLASSIFIED' && 
    j.status !== 'posted' &&
    j.status !== 'excluded'
  );
  
  categorizedJournals.forEach(journal => {
    if (!summary.has(journal.account)) {
      summary.set(journal.account, { totalDebits: 0, totalCredits: 0 });
    }
    const entry = summary.get(journal.account)!;
    entry.totalDebits += journal.debit;
    entry.totalCredits += journal.credit;
  });
  
  // Convert to array and calculate net effect
  return Array.from(summary.entries())
    .map(([account, amounts]) => ({
      account,
      totalDebits: amounts.totalDebits,
      totalCredits: amounts.totalCredits,
      netEffect: amounts.totalDebits - amounts.totalCredits // Positive = net debit, negative = net credit
    }))
    .sort((a, b) => a.account.localeCompare(b.account));
});

// Total debits across all accounts
const totalDebits = computed(() => {
  return accountSummary.value.reduce((sum, acc) => sum + acc.totalDebits, 0);
});

// Total credits across all accounts
const totalCredits = computed(() => {
  return accountSummary.value.reduce((sum, acc) => sum + acc.totalCredits, 0);
});

// Check if entries are balanced
const isBalanced = computed(() => {
  return Math.abs(totalDebits.value - totalCredits.value) < 100; // Allow for minor rounding (< $1)
});

// Formatters
function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US');
}

function formatStatus(status: string): string {
  const statusMap: Record<string, string> = {
    'pending_review': 'Pending',
    'approved': 'Approved',
    'duplicate': 'Duplicate',
    'excluded': 'Excluded',
  };
  return statusMap[status] || status;
}
</script>

