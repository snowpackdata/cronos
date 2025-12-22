<template>
  <div class="px-4 py-6 sm:px-6 lg:px-8">
    <!-- Combined Header Bar: Breadcrumb + Action Button -->
    <div class="flex items-center justify-between mb-4">
      <!-- Breadcrumb Navigation -->
      <nav class="flex" aria-label="Breadcrumb">
        <ol role="list" class="flex items-center space-x-2">
          <li>
            <div>
              <a
                href="#"
                @click.prevent="navigateToLevel('accounts')"
                :class="[
                  'text-sm font-medium flex items-center gap-1.5',
                  currentLevel === 'accounts' ? 'text-gray-700' : 'text-gray-500 hover:text-gray-700'
                ]"
              >
                <i class="fas fa-building text-xs"></i>
                Accounts
              </a>
            </div>
          </li>
          <li v-if="selectedAccount">
            <div class="flex items-center">
              <svg class="h-5 w-5 flex-shrink-0 text-gray-400" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path fill-rule="evenodd" d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z" clip-rule="evenodd" />
              </svg>
              <a
                href="#"
                @click.prevent="navigateToLevel('projects')"
                :class="[
                  'ml-2 text-sm font-medium flex items-center gap-1.5',
                  currentLevel === 'projects' ? 'text-gray-700' : 'text-gray-500 hover:text-gray-700'
                ]"
              >
                <i class="fas fa-building text-xs"></i>
                {{ selectedAccount.name }}
              </a>
            </div>
          </li>
          <li v-if="selectedProject">
            <div class="flex items-center">
              <svg class="h-5 w-5 flex-shrink-0 text-gray-400" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path fill-rule="evenodd" d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z" clip-rule="evenodd" />
              </svg>
              <a
                href="#"
                @click.prevent="navigateToLevel('billing-codes')"
                :class="[
                  'ml-2 text-sm font-medium flex items-center gap-1.5',
                  currentLevel === 'billing-codes' ? 'text-gray-700' : 'text-gray-500 hover:text-gray-700'
                ]"
              >
                <i class="fas fa-folder text-xs"></i>
                {{ selectedProject.name }}
              </a>
            </div>
          </li>
        </ol>
      </nav>

      <!-- Action Button -->
      <div>
        <button
          v-if="currentLevel === 'accounts'"
          type="button"
          @click="openAccountDrawer()"
          class="block rounded-md bg-sage px-2.5 py-1.5 text-center text-xs font-semibold text-white shadow-sm hover:bg-sage-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sage"
        >
          <i class="fas fa-plus-circle mr-1"></i> Create new account
        </button>
        <button
          v-else-if="currentLevel === 'projects'"
          type="button"
          @click="openProjectDrawer()"
          class="block rounded-md bg-sage px-2.5 py-1.5 text-center text-xs font-semibold text-white shadow-sm hover:bg-sage-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sage"
        >
          <i class="fas fa-plus-circle mr-1"></i> Create new project
        </button>
        <button
          v-else-if="currentLevel === 'billing-codes'"
          type="button"
          @click="openBillingCodeDrawer()"
          class="block rounded-md bg-sage px-2.5 py-1.5 text-center text-xs font-semibold text-white shadow-sm hover:bg-sage-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sage"
        >
          <i class="fas fa-plus-circle mr-1"></i> Create new billing code
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="flex flex-col items-center justify-center p-10 bg-white rounded-lg shadow mt-6">
      <i class="fas fa-spinner fa-spin text-4xl text-teal mb-4"></i>
      <span class="text-gray-dark">Loading...</span>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="flex flex-col items-center justify-center p-10 bg-white rounded-lg shadow mt-6">
      <i class="fas fa-exclamation-circle text-4xl text-red mb-4"></i>
      <span class="text-gray-dark mb-2">{{ error }}</span>
      <button @click="loadData" class="btn-secondary mt-4">
        <i class="fas fa-sync mr-2"></i> Retry
      </button>
    </div>

    <!-- Content Area -->
    <div v-else class="mt-4">
      <!-- Accounts List (Full Width) -->
      <div v-if="currentLevel === 'accounts'" class="flow-root">
        <ul role="list" class="grid grid-cols-1 gap-2">
          <li v-for="account in accounts" :key="account.ID" 
              class="cursor-pointer" 
              @click="selectAccount(account)">
            <AccountCard
              :account="account"
              @edit="(acc) => openAccountDrawer(acc)"
              @invite-client="(acc) => openInviteClientModal(acc)"
              @add-asset="(id) => openAssetUploaderModal(id)"
              @asset-deleted="handleAssetDeleted"
            />
          </li>
          <li v-if="accounts.length === 0" class="col-span-full py-5">
            <div class="flex flex-col items-center justify-center p-10 bg-white rounded-lg shadow">
              <i class="fas fa-building text-5xl text-gray-300 mb-4"></i>
              <p class="text-lg font-medium text-gray-dark">No client accounts found</p>
              <p class="text-gray mb-4">Click "Create new account" to add one</p>
            </div>
          </li>
        </ul>
      </div>

      <!-- Two-Column Layout: Account Detail + Projects -->
      <div v-else-if="currentLevel === 'projects' && selectedAccount" class="grid grid-cols-1 lg:grid-cols-5 gap-6">
        <!-- Left: Account Detail (2 columns = ~40%) -->
        <div class="lg:col-span-2">
          <!-- Section Header -->
          <div class="mb-2">
            <h2 class="text-xs font-semibold text-gray-500 uppercase tracking-wide">Account Details</h2>
          </div>
          <div class="bg-white shadow rounded-lg overflow-hidden sticky top-0 max-h-[calc(100vh-120px)] flex flex-col">
            <!-- Account Header - Dark Background -->
            <div class="bg-gray-900 px-3 py-2 flex-shrink-0">
              <div class="flex items-start justify-between">
                <div class="flex-1 min-w-0">
                  <h2 class="text-sm font-semibold text-white truncate">{{ selectedAccount.name }}</h2>
                  <p v-if="selectedAccount.legal_name && selectedAccount.legal_name !== selectedAccount.name" class="text-xs text-gray-400 truncate">{{ selectedAccount.legal_name }}</p>
                </div>
                <button
                  @click="openAccountDrawer(selectedAccount)"
                  class="ml-2 rounded-md bg-sage hover:bg-sage-dark px-2 py-1 text-xs font-semibold text-white transition-colors"
                >
                  <i class="fas fa-pencil-alt mr-1"></i> Edit
                </button>
              </div>
            </div>

            <!-- Account Details - Dense Grid Layout -->
            <div class="p-3 space-y-2.5 flex-1 overflow-y-auto">
              <!-- Primary Info -->
              <div class="space-y-2 text-xs">
                <div v-if="selectedAccount.email" class="flex items-center gap-2">
                  <i class="fas fa-envelope text-gray-400 w-4"></i>
                  <span class="text-gray-900 font-medium truncate">{{ selectedAccount.email }}</span>
                </div>
                <div v-if="selectedAccount.website" class="flex items-center gap-2">
                  <i class="fas fa-globe text-gray-400 w-4"></i>
                  <a :href="selectedAccount.website" target="_blank" class="text-sage hover:text-sage-dark font-medium truncate">{{ selectedAccount.website }}</a>
                </div>
                <div v-if="selectedAccount.address" class="flex items-start gap-2">
                  <i class="fas fa-map-marker-alt text-gray-400 w-4 mt-0.5"></i>
                  <span class="text-gray-900 leading-tight">{{ selectedAccount.address }}</span>
                </div>
              </div>

              <!-- Financial Info -->
              <div class="pt-3 border-t border-gray-200">
                <h3 class="text-xs font-semibold text-gray-700 mb-2"><i class="fas fa-money-bill-wave mr-1"></i> Financial</h3>
                <dl class="grid grid-cols-2 gap-x-3 gap-y-2 text-xs">
                  <div>
                    <dt class="text-gray-500 mb-0.5"><i class="fas fa-calendar w-4"></i> Billing</dt>
                    <dd class="text-gray-900 font-medium text-xs">{{ formatBillingFrequency(selectedAccount.billing_frequency) }}</dd>
                  </div>
                  <div v-if="selectedAccount.projects_single_invoice !== undefined">
                    <dt class="text-gray-500 mb-0.5"><i class="fas fa-file-invoice w-4"></i> Invoicing</dt>
                    <dd class="text-gray-900 font-medium text-xs">{{ selectedAccount.projects_single_invoice ? 'Combined' : 'Separate' }}</dd>
                  </div>
                  <div v-if="selectedAccount.budget_dollars">
                    <dt class="text-gray-500 mb-0.5"><i class="fas fa-dollar-sign w-4"></i> Period Budget</dt>
                    <dd class="text-gray-900 font-medium">${{ selectedAccount.budget_dollars.toLocaleString() }}{{ formatBudgetPeriod(selectedAccount.billing_frequency) }}</dd>
                  </div>
                  <div v-if="selectedAccount.budget_hours">
                    <dt class="text-gray-500 mb-0.5"><i class="fas fa-clock w-4"></i> Hours Cap</dt>
                    <dd class="text-gray-900 font-medium">{{ selectedAccount.budget_hours }} hrs{{ formatBudgetPeriod(selectedAccount.billing_frequency) }}</dd>
                  </div>
                  <div v-if="selectedAccount.budget_amount" class="col-span-2">
                    <dt class="text-gray-500 mb-0.5"><i class="fas fa-coins w-4"></i> Budget Cap</dt>
                    <dd class="text-gray-900 font-medium">${{ selectedAccount.budget_amount.toLocaleString() }}{{ selectedAccount.budget_year ? ` (${selectedAccount.budget_year})` : '' }}</dd>
                  </div>
                </dl>
              </div>

              <!-- Client Users -->
              <div v-if="(selectedAccount.clients && selectedAccount.clients.length > 0) || (selectedAccount.client_users && selectedAccount.client_users.length > 0)" class="pt-3 border-t border-gray-200">
                <h3 class="text-xs font-semibold text-gray-700 mb-2 flex items-center justify-between">
                  <span><i class="fas fa-users mr-1"></i> Client Users</span>
                  <button
                    @click="openInviteClientModal(selectedAccount)"
                    class="text-sage hover:text-sage-dark"
                  >
                    <i class="fas fa-plus-circle"></i>
                  </button>
                </h3>
                <ul class="space-y-1.5">
                  <li v-for="client in getUniqueClients(selectedAccount)" :key="client.ID || client.email" 
                      class="text-xs bg-gray-50 rounded px-2 py-1.5 flex items-center justify-between">
                    <span class="flex items-center gap-2">
                      <i class="fas fa-user text-gray-400"></i>
                      <span class="text-gray-900 font-medium">{{ client.email }}</span>
                    </span>
                    <span v-if="client.role" class="text-gray-500 text-xs">{{ client.role }}</span>
                  </li>
                  <li v-if="getUniqueClients(selectedAccount).length === 0">
                    <p class="text-xs text-gray-400 italic">No clients invited yet</p>
                  </li>
                </ul>
              </div>

              <!-- Assets -->
              <div class="pt-3 border-t border-gray-200">
                <h3 class="text-xs font-semibold text-gray-700 mb-2 flex items-center justify-between">
                  <span><i class="fas fa-paperclip mr-1"></i> Assets</span>
                  <button
                    @click="openAssetUploaderModal(selectedAccount.ID)"
                    class="text-sage hover:text-sage-dark"
                  >
                    <i class="fas fa-plus-circle"></i>
                  </button>
                </h3>
                <ul v-if="selectedAccount.assets && selectedAccount.assets.length > 0" class="space-y-1.5">
                  <li v-for="asset in selectedAccount.assets" :key="asset.ID" 
                      class="text-xs bg-gray-50 hover:bg-gray-100 rounded px-2 py-1.5 flex items-center justify-between group cursor-pointer transition-colors"
                      @click="handleAssetClick(asset)">
                    <span class="flex items-center gap-2 min-w-0 flex-1">
                      <i :class="[getAssetIcon(asset.asset_type, asset.content_type), getAssetIconColor(asset.asset_type, asset.content_type)]"></i>
                      <span class="text-gray-900 font-medium truncate group-hover:text-sage">{{ asset.name }}</span>
                      <i class="fas fa-external-link-alt text-xs text-gray-400 group-hover:text-sage"></i>
                    </span>
                    <button
                      @click.stop="confirmDeleteAsset(asset.ID, asset.name, true)"
                      class="text-red-600 hover:text-red-800 p-1 opacity-0 group-hover:opacity-100 transition-opacity"
                      title="Delete asset"
                    >
                      <i class="fas fa-trash-alt"></i>
                    </button>
                  </li>
                </ul>
                <p v-else class="text-xs text-gray-400 italic">No assets uploaded yet</p>
              </div>

            </div>
          </div>
        </div>

        <!-- Right: Projects List (3 columns = ~60%) -->
        <div class="lg:col-span-3">
          <div class="mb-2">
            <h2 class="text-xs font-semibold text-gray-500 uppercase tracking-wide">Projects</h2>
          </div>
          <ul role="list" class="space-y-2">
            <li v-for="project in filteredProjects" :key="project.ID"
                class="cursor-pointer"
                @click="selectProject(project)">
              <ProjectCard
                :project="project"
                @edit="(proj) => openProjectDrawer(proj)"
              />
            </li>
            <li v-if="filteredProjects.length === 0">
              <div class="flex flex-col items-center justify-center p-10 bg-white rounded-lg shadow">
                <i class="fas fa-folder text-5xl text-gray-300 mb-4"></i>
                <p class="text-lg font-medium text-gray-dark">No projects found for this account</p>
                <p class="text-gray mb-4">Click "Create new project" to add one</p>
              </div>
            </li>
          </ul>
        </div>
      </div>

      <!-- Two-Column Layout: Project Detail + Billing Codes -->
      <div v-else-if="currentLevel === 'billing-codes' && selectedProject" class="grid grid-cols-1 lg:grid-cols-5 gap-6">
        <!-- Left: Project Detail (2 columns = ~40%) -->
        <div class="lg:col-span-2">
          <!-- Section Header -->
          <div class="mb-2">
            <h2 class="text-xs font-semibold text-gray-500 uppercase tracking-wide">Project Details</h2>
          </div>
          <div class="bg-white shadow rounded-lg overflow-hidden sticky top-0 max-h-[calc(100vh-120px)] flex flex-col">
            <!-- Project Header - Dark Background -->
            <div class="bg-gray-900 px-3 py-2 flex-shrink-0">
              <div class="flex items-start justify-between">
                <div class="flex-1 min-w-0">
                  <h2 class="text-sm font-semibold text-white truncate">{{ selectedProject.name }}</h2>
                  <p v-if="selectedProject.account?.name" class="text-xs text-gray-400 truncate">{{ selectedProject.account.name }}</p>
                </div>
                <button
                  @click="openProjectDrawer(selectedProject)"
                  class="ml-2 rounded-md bg-sage hover:bg-sage-dark px-2 py-1 text-xs font-semibold text-white transition-colors"
                >
                  <i class="fas fa-pencil-alt mr-1"></i> Edit
                </button>
              </div>
            </div>

            <!-- Project Details - Dense Layout -->
            <div class="p-3 space-y-2.5 flex-1 overflow-y-auto">
              <!-- Status Badges -->
              <div class="flex flex-wrap gap-2">
                <span :class="[
                  isProjectActive(selectedProject) ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800',
                  'inline-flex items-center rounded-md px-2 py-1 text-xs font-medium'
                ]">
                  <i :class="isProjectActive(selectedProject) ? 'fas fa-check-circle' : 'fas fa-stop-circle'" class="mr-1"></i>
                  {{ isProjectActive(selectedProject) ? 'Active' : 'Ended' }}
                </span>
                <span v-if="selectedProject.internal" class="inline-flex items-center rounded-md bg-blue-100 px-2 py-1 text-xs font-medium text-blue-800">
                  <i class="fas fa-home mr-1"></i> Internal
                </span>
                <span v-if="selectedProject.project_type" class="inline-flex items-center rounded-md bg-purple-100 px-2 py-1 text-xs font-medium text-purple-800">
                  <i class="fas fa-tag mr-1"></i> {{ selectedProject.project_type === 'PROJECT_TYPE_NEW' ? 'New' : 'Existing' }}
                </span>
              </div>

              <!-- Description -->
              <div v-if="selectedProject.description" class="text-xs text-gray-600 leading-relaxed">
                {{ selectedProject.description }}
              </div>

              <!-- Timeline -->
              <div class="pt-3 border-t border-gray-200">
                <h3 class="text-xs font-semibold text-gray-700 mb-2"><i class="fas fa-calendar mr-1"></i> Timeline</h3>
                <dl class="space-y-2 text-xs">
                  <div v-if="selectedProject.active_start" class="flex justify-between">
                    <dt class="text-gray-500">Start:</dt>
                    <dd class="text-gray-900 font-medium">{{ formatDate(selectedProject.active_start) }}</dd>
                  </div>
                  <div v-if="selectedProject.active_end" class="flex justify-between">
                    <dt class="text-gray-500">End:</dt>
                    <dd class="text-gray-900 font-medium">{{ formatDate(selectedProject.active_end) }}</dd>
                  </div>
                  <div v-if="selectedProject.active_start && selectedProject.active_end" class="flex justify-between pt-1 border-t border-gray-100">
                    <dt class="text-gray-500">Duration:</dt>
                    <dd class="text-gray-900 font-medium">{{ calculateDuration(selectedProject.active_start, selectedProject.active_end) }}</dd>
                  </div>
                </dl>
              </div>

              <!-- Financial Info -->
              <div class="pt-3 border-t border-gray-200">
                <h3 class="text-xs font-semibold text-gray-700 mb-2"><i class="fas fa-money-bill-wave mr-1"></i> Financial</h3>
                <dl class="grid grid-cols-2 gap-x-3 gap-y-2 text-xs">
                  <div>
                    <dt class="text-gray-500 mb-0.5"><i class="fas fa-calendar w-4"></i> Billing</dt>
                    <dd class="text-gray-900 font-medium text-xs">{{ formatBillingFrequency(selectedProject.billing_frequency) }}</dd>
                  </div>
                  <div v-if="selectedProject.budget_dollars">
                    <dt class="text-gray-500 mb-0.5"><i class="fas fa-dollar-sign w-4"></i> Period Budget</dt>
                    <dd class="text-gray-900 font-medium">${{ selectedProject.budget_dollars.toLocaleString() }}{{ formatBudgetPeriod(selectedProject.billing_frequency) }}</dd>
                  </div>
                  <div v-if="selectedProject.budget_hours">
                    <dt class="text-gray-500 mb-0.5"><i class="fas fa-clock w-4"></i> Hours</dt>
                    <dd class="text-gray-900 font-medium">{{ selectedProject.budget_hours }} hrs{{ formatBudgetPeriod(selectedProject.billing_frequency) }}</dd>
                  </div>
                  <div v-if="selectedProject.budget_cap_dollars" class="col-span-2">
                    <dt class="text-gray-500 mb-0.5"><i class="fas fa-coins w-4"></i> Budget Cap</dt>
                    <dd class="text-gray-900 font-medium">${{ selectedProject.budget_cap_dollars.toLocaleString() }}</dd>
                  </div>
                  <div v-if="selectedProject.budget_cap_hours" class="col-span-2">
                    <dt class="text-gray-500 mb-0.5"><i class="fas fa-hourglass-half w-4"></i> Hours Cap</dt>
                    <dd class="text-gray-900 font-medium">{{ selectedProject.budget_cap_hours }} hrs</dd>
                  </div>
                </dl>
              </div>

              <!-- Team -->
              <div v-if="selectedProject.ae_id || selectedProject.sdr_id" class="pt-3 border-t border-gray-200">
                <h3 class="text-xs font-semibold text-gray-700 mb-2"><i class="fas fa-briefcase mr-1"></i> Team</h3>
                <dl class="space-y-2 text-xs">
                  <div v-if="selectedProject.ae_id" class="flex items-center gap-2">
                    <i class="fas fa-user-tie text-gray-400 w-4"></i>
                    <span class="text-gray-500">AE:</span>
                    <span class="text-gray-900 font-medium">{{ getStaffName(selectedProject.ae_id) }}</span>
                  </div>
                  <div v-if="selectedProject.sdr_id" class="flex items-center gap-2">
                    <i class="fas fa-user-check text-gray-400 w-4"></i>
                    <span class="text-gray-500">SDR:</span>
                    <span class="text-gray-900 font-medium">{{ getStaffName(selectedProject.sdr_id) }}</span>
                  </div>
                </dl>
              </div>

              <!-- Assigned Staff -->
              <div class="pt-3 border-t border-gray-200">
                <h3 class="text-xs font-semibold text-gray-700 mb-2">
                  <i class="fas fa-user-friends mr-1"></i> Assigned Staff
                  <span class="ml-1 text-gray-500 font-normal">({{ selectedProject.staffing_assignments?.length || 0 }})</span>
                </h3>
                <ul v-if="selectedProject.staffing_assignments && selectedProject.staffing_assignments.length > 0" class="space-y-1.5">
                  <li v-for="assignment in selectedProject.staffing_assignments" :key="assignment.ID" 
                      class="text-xs bg-gray-50 rounded px-2 py-1.5 flex items-center justify-between">
                    <span class="flex items-center gap-2">
                      <i class="fas fa-user text-gray-400"></i>
                      <span class="text-gray-900 font-medium">{{ assignment.employee?.first_name }} {{ assignment.employee?.last_name }}</span>
                    </span>
                    <span class="text-gray-500 font-medium">{{ assignment.commitment }}h/wk</span>
                  </li>
                </ul>
                <p v-else class="text-xs text-gray-400 italic">No staff assigned yet</p>
              </div>

              <!-- Assets -->
              <div v-if="selectedProject.assets && selectedProject.assets.length > 0" class="pt-3 border-t border-gray-200">
                <h3 class="text-xs font-semibold text-gray-700 mb-2">
                  <i class="fas fa-paperclip mr-1"></i> Assets
                  <span class="ml-1 text-gray-500 font-normal">({{ selectedProject.assets.length }})</span>
                </h3>
                <ul class="space-y-1.5">
                  <li v-for="asset in selectedProject.assets" :key="asset.ID"
                      class="text-xs bg-gray-50 hover:bg-gray-100 rounded px-2 py-1.5 flex items-center justify-between group cursor-pointer transition-colors"
                      @click="handleAssetClick(asset)">
                    <span class="flex items-center gap-2 min-w-0 flex-1">
                      <i :class="[getAssetIcon(asset.asset_type, asset.content_type), getAssetIconColor(asset.asset_type, asset.content_type)]"></i>
                      <span class="text-gray-900 font-medium truncate group-hover:text-sage">{{ asset.name }}</span>
                      <i class="fas fa-external-link-alt text-xs text-gray-400 group-hover:text-sage"></i>
                    </span>
                    <button
                      @click.stop="confirmDeleteAsset(asset.ID, asset.name, false)"
                      class="text-red-600 hover:text-red-800 p-1 opacity-0 group-hover:opacity-100 transition-opacity"
                      title="Delete asset"
                    >
                      <i class="fas fa-trash-alt"></i>
                    </button>
                  </li>
                </ul>
              </div>
            </div>
          </div>
        </div>

        <!-- Right: Billing Codes List (3 columns = ~60%) -->
        <div class="lg:col-span-3">
          <div class="mb-2">
            <h2 class="text-xs font-semibold text-gray-500 uppercase tracking-wide">Billing Codes</h2>
          </div>
          <div class="bg-white shadow rounded-lg">
            <ul role="list" class="divide-y divide-gray-100">
              <li v-for="billingCode in filteredBillingCodes" :key="billingCode.ID" 
                  class="flex items-center justify-between gap-x-6 py-5 px-4 hover:bg-gray-50 cursor-pointer"
                  @click.stop="openBillingCodeDrawer(billingCode)">
                <div class="min-w-0 flex-1">
                  <div class="flex items-start gap-x-3">
                    <p class="text-sm/6 font-semibold text-gray-900">{{ billingCode.name }}</p>
                    <p :class="[
                      isBillingCodeActive(billingCode) ? 'text-green-700 bg-green-50 ring-green-600/20' : 'text-red-700 bg-red-50 ring-red-600/20',
                      'mt-0.5 whitespace-nowrap rounded-md px-1.5 py-0.5 text-xs font-medium ring-1 ring-inset'
                    ]">
                      {{ isBillingCodeActive(billingCode) ? 'Active' : 'Inactive' }}
                    </p>
                    <p class="mt-0.5 whitespace-nowrap rounded-md px-1.5 py-0.5 text-xs font-medium bg-indigo-50 text-indigo-700 ring-1 ring-inset ring-indigo-600/20">
                      {{ formatRateType(billingCode.type) }}
                    </p>
                  </div>
                  <div class="mt-1 flex items-center gap-x-2 text-xs/5 text-gray-500">
                    <p class="whitespace-nowrap">
                      <span class="font-medium">Code: {{ billingCode.code }}</span>
                    </p>
                    <svg viewBox="0 0 2 2" class="size-0.5 fill-current">
                      <circle cx="1" cy="1" r="1" />
                    </svg>
                    <p class="whitespace-nowrap">
                      Rate: <span class="font-medium">{{ getRateName(billingCode.rate_id) }}</span>
                    </p>
                    <svg viewBox="0 0 2 2" class="size-0.5 fill-current">
                      <circle cx="1" cy="1" r="1" />
                    </svg>
                    <p class="whitespace-nowrap">
                      <span class="font-medium">${{ getRateDetails(billingCode.rate_id, billingCode.internal_rate_id).value }}/hr</span>
                    </p>
                    <svg viewBox="0 0 2 2" class="size-0.5 fill-current">
                      <circle cx="1" cy="1" r="1" />
                    </svg>
                    <p class="whitespace-nowrap">
                      Margin: <span class="font-medium">{{ getRateDetails(billingCode.rate_id, billingCode.internal_rate_id).margin.toFixed(1) }}%</span>
                    </p>
                  </div>
                  <div v-if="billingCode.active_start || billingCode.active_end" class="mt-1 flex items-center gap-x-2 text-xs/5 text-gray-500">
                    <p v-if="billingCode.active_start" class="whitespace-nowrap">
                      Active from <time :datetime="billingCode.active_start">{{ formatDate(billingCode.active_start) }}</time>
                    </p>
                    <svg v-if="billingCode.active_start && billingCode.active_end" viewBox="0 0 2 2" class="size-0.5 fill-current">
                      <circle cx="1" cy="1" r="1" />
                    </svg>
                    <p v-if="billingCode.active_end" class="whitespace-nowrap">
                      to <time :datetime="billingCode.active_end">{{ formatDate(billingCode.active_end) }}</time>
                    </p>
                  </div>
                </div>
                <div class="flex flex-none items-center gap-x-4">
                  <button
                    @click.stop="openBillingCodeDrawer(billingCode)"
                    class="rounded-md bg-white px-2.5 py-1.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                  >
                    <i class="fas fa-pencil-alt mr-1"></i> Edit
                  </button>
                </div>
              </li>
              <li v-if="filteredBillingCodes.length === 0" class="py-5">
                <div class="flex flex-col items-center justify-center p-10">
                  <i class="fas fa-code text-5xl text-gray-300 mb-4"></i>
                  <p class="text-lg font-medium text-gray-dark">No billing codes found for this project</p>
                  <p class="text-gray mb-4">Click "Create new billing code" to add one</p>
                </div>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>

    <!-- Account Drawer -->
    <AccountDrawer
      :is-open="isAccountDrawerOpen"
      :account-data="editingAccount"
      @close="closeAccountDrawer"
      @save="saveAccount"
      @delete="handleDeleteAccount"
    />

    <!-- Project Drawer -->
    <ProjectDrawer
      :is-open="isProjectDrawerOpen"
      :project-data="editingProject"
      :staff-members="staffMembers"
      @close="closeProjectDrawer"
      @save="saveProject"
    />

    <!-- Billing Code Drawer -->
    <BillingCodeDrawer
      :is-open="isBillingCodeDrawerOpen"
      :billing-code-data="editingBillingCode"
      :project-id="selectedProject?.ID || null"
      @close="closeBillingCodeDrawer"
      @save="saveBillingCode"
      @delete="handleDeleteBillingCode"
    />

    <!-- Asset Uploader Modal -->
    <AssetUploaderModal 
      :is-open="isAssetUploaderOpen" 
      :account-id="selectedAccountIdForAsset"
      @close="closeAssetUploaderModal" 
      @save="handleSaveAsset"
    />

    <!-- Invite Client Modal -->
    <div v-if="showInviteClientModal" class="fixed inset-0 z-50 overflow-y-auto bg-gray-500 bg-opacity-75 transition-opacity" aria-labelledby="modal-title" role="dialog" aria-modal="true">
      <div class="flex min-h-full items-center justify-center p-4 text-center sm:p-0">
        <div class="relative transform overflow-hidden rounded-lg bg-white text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg">
          <div class="bg-white px-4 pb-4 pt-5 sm:p-6 sm:pb-4">
            <div class="sm:flex sm:items-start">
              <div class="mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-indigo-100 sm:mx-0 sm:h-10 sm:w-10">
                <i class="fas fa-user-plus text-indigo-600"></i>
              </div>
              <div class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left">
                <h3 class="text-base font-semibold leading-6 text-gray-900" id="modal-title">
                  Invite New Client to {{ accountToInviteClientTo?.name }}
                </h3>
                <div class="mt-2">
                  <p class="text-sm text-gray-500">
                    Enter the email address of the client you want to invite. They will receive an email to set up their account.
                  </p>
                  <input 
                    type="email" 
                    v-model="newClientEmail"
                    placeholder="client@example.com"
                    class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                  />
                </div>
              </div>
            </div>
          </div>
          <div class="bg-gray-50 px-4 py-3 sm:flex sm:flex-row-reverse sm:px-6">
            <button 
              type="button" 
              @click="handleSendInvite"
              :disabled="!newClientEmail.trim()"
              class="inline-flex w-full justify-center rounded-md bg-sage px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-sage-dark sm:ml-3 sm:w-auto disabled:opacity-50">
              Send Invite
            </button>
            <button 
              type="button" 
              @click="closeInviteClientModal"
              class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto">
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Delete Asset Confirmation Modal -->
  <div v-if="showDeleteAssetModal" class="relative z-50" aria-labelledby="modal-title" role="dialog" aria-modal="true">
    <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"></div>
    <div class="fixed inset-0 z-10 overflow-y-auto">
      <div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
        <div class="relative transform overflow-hidden rounded-lg bg-white text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg">
          <div class="bg-white px-4 pb-4 pt-5 sm:p-6 sm:pb-4">
            <div class="sm:flex sm:items-start">
              <div class="mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-red-100 sm:mx-0 sm:h-10 sm:w-10">
                <i class="fas fa-exclamation-triangle text-red-600"></i>
              </div>
              <div class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left">
                <h3 class="text-base font-semibold leading-6 text-gray-900" id="modal-title">Delete Asset</h3>
                <div class="mt-2">
                  <p class="text-sm text-gray-500">
                    Are you sure you want to delete <strong>{{ assetToDelete?.name }}</strong>? This action cannot be undone.
                  </p>
                </div>
              </div>
            </div>
          </div>
          <div class="bg-gray-50 px-4 py-3 sm:flex sm:flex-row-reverse sm:px-6">
            <button 
              type="button" 
              @click="handleDeleteAsset"
              class="inline-flex w-full justify-center rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 sm:ml-3 sm:w-auto">
              Delete
            </button>
            <button 
              type="button" 
              @click="cancelDeleteAsset"
              class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto">
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { fetchAccounts, getAccount, createAccount, updateAccount, deleteAccount as deleteAccountAPI, inviteUser } from '../../api/accounts';
import { fetchProjects, fetchProjectById, createProject, updateProject } from '../../api/projects';
import { getBillingCodes, createBillingCode, updateBillingCode, deleteBillingCode as deleteBillingCodeAPI } from '../../api/billingCodes';
import { fetchRates } from '../../api/rates';
import { getUsers } from '../../api/timesheet';
import { createAsset, deleteAsset, refreshAssetUrl } from '../../api/assets';
import type { Account } from '../../types/Account';
import type { Project } from '../../types/Project';
import type { BillingCode } from '../../types/BillingCode';
import type { Rate } from '../../types/Rate';
import type { Asset } from '../../types/Asset';
import AccountCard from '../../components/accounts/AccountCard.vue';
import ProjectCard from '../../components/projects/ProjectCard.vue';
import AccountDrawer from '../../components/accounts/AccountDrawer.vue';
import ProjectDrawer from '../../components/projects/ProjectDrawer.vue';
import BillingCodeDrawer from '../../components/billing-codes/BillingCodeDrawer.vue';
import AssetUploaderModal from '../../components/assets/AssetUploaderModal.vue';

const route = useRoute();
const router = useRouter();

// State
type NavigationLevel = 'accounts' | 'projects' | 'billing-codes';
const currentLevel = ref<NavigationLevel>('accounts');
const isLoading = ref(true);
const error = ref<string | null>(null);

// Data
const accounts = ref<Account[]>([]);
const projects = ref<Project[]>([]);
const billingCodes = ref<BillingCode[]>([]);
const rates = ref<Rate[]>([]);
const staffMembers = ref<any[]>([]);

// Selected items
const selectedAccount = ref<Account | null>(null);
const selectedProject = ref<Project | null>(null);
const selectedBillingCode = ref<BillingCode | null>(null);

// Drawer states
const isAccountDrawerOpen = ref(false);
const isProjectDrawerOpen = ref(false);
const isBillingCodeDrawerOpen = ref(false);
const editingAccount = ref<Account | null>(null);
const editingProject = ref<Project | null>(null);
const editingBillingCode = ref<BillingCode | null>(null);

// Asset uploader
const isAssetUploaderOpen = ref(false);
const selectedAccountIdForAsset = ref<number | null>(null);

// Invite client
const showInviteClientModal = ref(false);
const accountToInviteClientTo = ref<Account | null>(null);
const newClientEmail = ref('');

// Delete asset confirmation
const showDeleteAssetModal = ref(false);
const assetToDelete = ref<{ id: number; name: string; isAccount: boolean } | null>(null);

// Computed properties
const filteredProjects = computed(() => {
  if (!selectedAccount.value) return [];
  return projects.value.filter(p => p.account_id === selectedAccount.value!.ID);
});

const filteredBillingCodes = computed(() => {
  if (!selectedProject.value) return [];
  return billingCodes.value.filter(bc => bc.project_id === selectedProject.value!.ID);
});

// Navigation functions
const navigateToLevel = (level: NavigationLevel) => {
  currentLevel.value = level;
  
  if (level === 'accounts') {
    selectedAccount.value = null;
    selectedProject.value = null;
    selectedBillingCode.value = null;
    updateURL();
  } else if (level === 'projects') {
    selectedProject.value = null;
    selectedBillingCode.value = null;
    updateURL();
  } else if (level === 'billing-codes') {
    selectedBillingCode.value = null;
    updateURL();
  }
};

const selectAccount = async (account: Account) => {
  // If the account doesn't have nested data loaded (minimal mode), fetch full details
  if (!account.clients && !account.client_users && !account.assets) {
    try {
      const fullAccount = await getAccount(account.ID);
      // Update the account in the list with full data
      const index = accounts.value.findIndex(a => a.ID === account.ID);
      if (index !== -1) {
        accounts.value[index] = fullAccount;
      }
      selectedAccount.value = fullAccount;
    } catch (err) {
      console.error('Error loading full account details:', err);
      selectedAccount.value = account;
    }
  } else {
    selectedAccount.value = account;
  }
  
  currentLevel.value = 'projects';
  updateURL();

  // Load projects for this account if not already loaded
  if (projects.value.filter(p => p.account_id === account.ID).length === 0) {
    await loadProjectsForAccount(account.ID);
  }
};

const selectProject = async (project: Project) => {
  selectedProject.value = project;
  currentLevel.value = 'billing-codes';
  updateURL();

  // If the project has billing codes already loaded, use them
  if (project.billing_codes && project.billing_codes.length > 0) {
    // Add project's billing codes to our list if not already there
    const newBillingCodes = project.billing_codes.filter(bc =>
      !billingCodes.value.find(existing => existing.ID === bc.ID)
    );
    billingCodes.value = [...billingCodes.value, ...newBillingCodes];
  } else if (billingCodes.value.filter(bc => bc.project_id === project.ID).length === 0) {
    // Otherwise load them via API
    await loadBillingCodesForProject(project.ID);
  }
  
  // Load rates if not already loaded (for displaying rate names)
  if (rates.value.length === 0) {
    rates.value = await fetchRates();
  }
};

// Update URL with query parameters
const updateURL = () => {
  const query: Record<string, string> = {};
  
  if (selectedAccount.value) {
    query.accountId = selectedAccount.value.ID.toString();
  }
  if (selectedProject.value) {
    query.projectId = selectedProject.value.ID.toString();
  }
  if (selectedBillingCode.value) {
    query.billingCodeId = selectedBillingCode.value.ID.toString();
  }
  
  router.replace({ query });
};

// Parse URL query parameters on mount (and load necessary data)
const parseURLParams = async () => {
  const accountId = route.query.accountId ? parseInt(route.query.accountId as string) : null;
  const projectId = route.query.projectId ? parseInt(route.query.projectId as string) : null;
  const billingCodeId = route.query.billingCodeId ? parseInt(route.query.billingCodeId as string) : null;
  
  if (accountId) {
    const account = accounts.value.find(a => a.ID === accountId);
    if (account) {
      selectedAccount.value = account;
      currentLevel.value = 'projects';
      
      // Load projects for this account
      await loadProjectsForAccount(accountId);
      
      if (projectId) {
        const project = projects.value.find(p => p.ID === projectId);
        if (project) {
          selectedProject.value = project;
          currentLevel.value = 'billing-codes';

          // Use project's preloaded billing codes if available
          if (project.billing_codes && project.billing_codes.length > 0) {
            const newBillingCodes = project.billing_codes.filter(bc =>
              !billingCodes.value.find(existing => existing.ID === bc.ID)
            );
            billingCodes.value = [...billingCodes.value, ...newBillingCodes];
          } else {
            // Otherwise load them via API
            await loadBillingCodesForProject(projectId);
          }
          
          // Load rates if not already loaded
          if (rates.value.length === 0) {
            rates.value = await fetchRates();
          }

          if (billingCodeId) {
            const billingCode = billingCodes.value.find(bc => bc.ID === billingCodeId);
            if (billingCode) {
              selectedBillingCode.value = billingCode;
              // Optionally open the billing code drawer
              openBillingCodeDrawer(billingCode);
            }
          }
        }
      }
    }
  }
};

// Load initial data (only accounts)
const loadData = async () => {
  isLoading.value = true;
  error.value = null;

  try {
    const [accountsData, staffData] = await Promise.all([
      fetchAccounts(true), // Pass true for minimal load - don't preload nested data
      getUsers() // Load staff for assignment modals
    ]);

    accounts.value = accountsData;
    staffMembers.value = staffData;

    // Parse URL params after initial data is loaded
    await parseURLParams();
  } catch (err) {
    console.error('Error loading organization data:', err);
    error.value = 'Failed to load organization data. Please try again.';
  } finally {
    isLoading.value = false;
  }
};

// Load projects for a specific account
const loadProjectsForAccount = async (accountId: number) => {
  try {
    const allProjects = await fetchProjects();
    // Only add projects we don't already have
    const newProjects = allProjects.filter(p => 
      p.account_id === accountId && !projects.value.find(existing => existing.ID === p.ID)
    );
    projects.value = [...projects.value, ...newProjects];
  } catch (err) {
    console.error('Error loading projects:', err);
  }
};

// Load billing codes for a specific project
const loadBillingCodesForProject = async (projectId: number) => {
  try {
    const projectBillingCodes = await getBillingCodes(projectId);
    // Only add billing codes we don't already have
    const newBillingCodes = projectBillingCodes.filter(bc => 
      !billingCodes.value.find(existing => existing.ID === bc.ID)
    );
    billingCodes.value = [...billingCodes.value, ...newBillingCodes];
    
    // Load rates if not already loaded (for displaying rate names)
    if (rates.value.length === 0) {
      rates.value = await fetchRates();
    }
  } catch (err) {
    console.error('Error loading billing codes:', err);
  }
};

// Account drawer functions
const openAccountDrawer = (account: Account | null = null) => {
  editingAccount.value = account;
  isAccountDrawerOpen.value = true;
};

const closeAccountDrawer = () => {
  isAccountDrawerOpen.value = false;
  editingAccount.value = null;
};

const saveAccount = async (accountData: any) => {
  try {
    if (editingAccount.value?.ID) {
      const payload = { ...accountData, ID: editingAccount.value.ID };
      const updatedAccount = await updateAccount(payload);
      // Update the account in the list
      const index = accounts.value.findIndex(a => a.ID === editingAccount.value!.ID);
      if (index !== -1) {
        accounts.value[index] = updatedAccount;
      }
      // Update selected account if it's the one being edited
      if (selectedAccount.value?.ID === editingAccount.value.ID) {
        selectedAccount.value = updatedAccount;
      }
    } else {
      const newAccount = await createAccount(accountData);
      accounts.value.push(newAccount);
    }
    closeAccountDrawer();
  } catch (err) {
    console.error('Error saving account:', err);
    alert('Failed to save account. Please try again.');
  }
};

const handleDeleteAccount = async (accountId: number) => {
  if (!confirm('Are you sure you want to delete this account?')) return;
  
  try {
    await deleteAccountAPI(accountId);
    // Remove from local list
    accounts.value = accounts.value.filter(a => a.ID !== accountId);
    // Remove associated projects and billing codes
    projects.value = projects.value.filter(p => p.account_id !== accountId);
    closeAccountDrawer();
    navigateToLevel('accounts');
  } catch (err) {
    console.error('Error deleting account:', err);
    alert('Failed to delete account. Please try again.');
  }
};

// Project drawer functions
const openProjectDrawer = (project: Project | null = null) => {
  editingProject.value = project;
  isProjectDrawerOpen.value = true;
};

const closeProjectDrawer = () => {
  isProjectDrawerOpen.value = false;
  editingProject.value = null;
};

const saveProject = async (projectData: any) => {
  try {
    // If creating a new project and we have a selected account, set the account_id
    if (!editingProject.value && selectedAccount.value) {
      projectData.account_id = selectedAccount.value.ID;
    }
    
    if (editingProject.value?.ID) {
      const updatedProject = await updateProject(editingProject.value.ID, projectData);
      // Update the project in the list
      const index = projects.value.findIndex(p => p.ID === editingProject.value!.ID);
      if (index !== -1) {
        projects.value[index] = updatedProject;
      }
      // Update selected project if it's the one being edited
      if (selectedProject.value?.ID === editingProject.value.ID) {
        selectedProject.value = updatedProject;
      }
    } else {
      const newProject = await createProject(projectData);
      projects.value.push(newProject);
    }
    closeProjectDrawer();
  } catch (err) {
    console.error('Error saving project:', err);
    alert('Failed to save project. Please try again.');
  }
};

// Billing code drawer functions
const openBillingCodeDrawer = (billingCode: BillingCode | null = null) => {
  editingBillingCode.value = billingCode;
  isBillingCodeDrawerOpen.value = true;
};

const closeBillingCodeDrawer = () => {
  isBillingCodeDrawerOpen.value = false;
  editingBillingCode.value = null;
};

const saveBillingCode = async (billingCodeData: any) => {
  try {
    // If creating a new billing code and we have a selected project, set the project_id
    if (!editingBillingCode.value && selectedProject.value) {
      billingCodeData.project_id = selectedProject.value.ID;
    }
    
    if (editingBillingCode.value?.ID) {
      const payload = { ...billingCodeData, ID: editingBillingCode.value.ID };
      const updatedBillingCode = await updateBillingCode(payload);
      // Update the billing code in the list
      const index = billingCodes.value.findIndex(bc => bc.ID === editingBillingCode.value!.ID);
      if (index !== -1) {
        billingCodes.value[index] = updatedBillingCode;
      }
    } else {
      const newBillingCode = await createBillingCode(billingCodeData);
      billingCodes.value.push(newBillingCode);
    }
    closeBillingCodeDrawer();
  } catch (err) {
    console.error('Error saving billing code:', err);
    alert('Failed to save billing code. Please try again.');
  }
};

const handleDeleteBillingCode = async (billingCodeId: number) => {
  if (!confirm('Are you sure you want to delete this billing code?')) return;
  
  try {
    await deleteBillingCodeAPI(billingCodeId);
    // Remove from local list
    billingCodes.value = billingCodes.value.filter(bc => bc.ID !== billingCodeId);
    closeBillingCodeDrawer();
  } catch (err) {
    console.error('Error deleting billing code:', err);
    alert('Failed to delete billing code. Please try again.');
  }
};

// Asset uploader functions
const openAssetUploaderModal = (accountId: number) => {
  selectedAccountIdForAsset.value = accountId;
  isAssetUploaderOpen.value = true;
};

const closeAssetUploaderModal = () => {
  isAssetUploaderOpen.value = false;
  selectedAccountIdForAsset.value = null;
};

const handleSaveAsset = async (assetData: Asset) => {
  try {
    await createAsset(assetData);
    // Refresh just the account to get updated assets
    if (selectedAccountIdForAsset.value) {
      const accountsData = await fetchAccounts();
      const updatedAccount = accountsData.find(a => a.ID === selectedAccountIdForAsset.value);
      if (updatedAccount) {
        const index = accounts.value.findIndex(a => a.ID === selectedAccountIdForAsset.value);
        if (index !== -1) {
          accounts.value[index] = updatedAccount;
        }
        if (selectedAccount.value?.ID === selectedAccountIdForAsset.value) {
          selectedAccount.value = updatedAccount;
        }
      }
    }
    closeAssetUploaderModal();
  } catch (err) {
    console.error('Error saving asset:', err);
    alert('Failed to save asset. Please try again.');
  }
};


// Invite client functions
const openInviteClientModal = (account: Account) => {
  accountToInviteClientTo.value = account;
  newClientEmail.value = '';
  showInviteClientModal.value = true;
};

const closeInviteClientModal = () => {
  showInviteClientModal.value = false;
  accountToInviteClientTo.value = null;
  newClientEmail.value = '';
};

const handleSendInvite = async () => {
  if (!accountToInviteClientTo.value || !newClientEmail.value.trim()) return;
  
  try {
    await inviteUser(accountToInviteClientTo.value.ID, newClientEmail.value.trim());
    alert('Invitation sent successfully!');
    closeInviteClientModal();
    
    // Refresh just the account to show the new user
    const accountsData = await fetchAccounts();
    const updatedAccount = accountsData.find(a => a.ID === accountToInviteClientTo.value!.ID);
    if (updatedAccount) {
      const index = accounts.value.findIndex(a => a.ID === accountToInviteClientTo.value!.ID);
      if (index !== -1) {
        accounts.value[index] = updatedAccount;
      }
      if (selectedAccount.value?.ID === accountToInviteClientTo.value!.ID) {
        selectedAccount.value = updatedAccount;
      }
    }
  } catch (err) {
    console.error('Error sending invite:', err);
    alert('Failed to send invitation. Please try again.');
  }
};

const handleAssetDeleted = async () => {
  // Refresh accounts to get updated assets
  const accountsData = await fetchAccounts();
  accounts.value = accountsData;
  if (selectedAccount.value) {
    const updatedAccount = accountsData.find(a => a.ID === selectedAccount.value!.ID);
    if (updatedAccount) {
      selectedAccount.value = updatedAccount;
    }
  }
};

const confirmDeleteAsset = (assetId: number, assetName: string, isAccount: boolean) => {
  assetToDelete.value = { id: assetId, name: assetName, isAccount };
  showDeleteAssetModal.value = true;
};

const handleDeleteAsset = async () => {
  if (!assetToDelete.value) return;

  const { id, isAccount } = assetToDelete.value;
  
  try {
    await deleteAsset(id);
    
    // Close modal and reset state
    showDeleteAssetModal.value = false;
    assetToDelete.value = null;
    
    if (isAccount && selectedAccount.value) {
      // Refresh the account data by fetching full account details
      const refreshedAccount = await getAccount(selectedAccount.value.ID);
      // Force reactivity by creating a new object
      selectedAccount.value = { ...refreshedAccount };
      
      // Also update in the accounts list
      const accountIndex = accounts.value.findIndex(a => a.ID === selectedAccount.value!.ID);
      if (accountIndex !== -1) {
        accounts.value[accountIndex] = { ...refreshedAccount };
      }
    } else if (!isAccount && selectedProject.value) {
      // Refresh the project data
      const refreshedProject = await fetchProjectById(selectedProject.value.ID);
      // Force reactivity by creating a new object
      selectedProject.value = { ...refreshedProject };
      
      // Also update in the projects list
      const projectIndex = projects.value.findIndex(p => p.ID === selectedProject.value!.ID);
      if (projectIndex !== -1) {
        projects.value[projectIndex] = { ...refreshedProject };
      }
    }
  } catch (error) {
    console.error('Error deleting asset:', error);
    showDeleteAssetModal.value = false;
    assetToDelete.value = null;
    // You could add a toast notification here instead of alert
  }
};

const cancelDeleteAsset = () => {
  showDeleteAssetModal.value = false;
  assetToDelete.value = null;
};

const handleAssetClick = async (asset: any) => {
  try {
    let urlToOpen = asset.url;
    
    // Check if this is a GCS asset that needs a signed URL
    // GCS assets have gcs_object_path or the URL contains storage.googleapis.com
    const isGCSAsset = asset.gcs_object_path || (asset.url && asset.url.includes('storage.googleapis.com'));
    
    if (isGCSAsset) {
      // Always refresh GCS assets to get a valid signed URL
      try {
        const refreshed = await refreshAssetUrl(asset.ID);
        urlToOpen = refreshed.new_url;
        // Update the asset object with new values
        asset.url = refreshed.new_url;
        asset.expires_at = refreshed.new_expires_at;
      } catch (refreshError) {
        console.error('Error refreshing asset URL:', refreshError);
        alert('Failed to generate access URL for this asset. Please try again.');
        return;
      }
    }

    // Open the asset URL
    if (urlToOpen) {
      window.open(urlToOpen, '_blank');
    }
  } catch (error) {
    console.error('Error accessing asset:', error);
    alert('Failed to access asset. Please try again.');
  }
};

// Helper functions
const isBillingCodeActive = (billingCode: BillingCode): boolean => {
  const now = new Date();
  const start = billingCode.active_start ? new Date(billingCode.active_start) : null;
  const end = billingCode.active_end ? new Date(billingCode.active_end) : null;
  
  const isAfterStart = start ? now >= start : true;
  const isBeforeEnd = end ? now <= end : true;
  
  return isAfterStart && isBeforeEnd;
};

const isProjectActive = (project: Project): boolean => {
  const now = new Date();
  const start = project.active_start ? new Date(project.active_start) : null;
  const end = project.active_end ? new Date(project.active_end) : null;
  
  const isAfterStart = start ? now >= start : true;
  const isBeforeEnd = end ? now <= end : true;
  
  return isAfterStart && isBeforeEnd;
};

const getRateName = (rateId: number | null | undefined): string => {
  if (!rateId) return 'N/A';
  const rate = rates.value.find(r => r.ID === rateId);
  return rate ? rate.name : 'Unknown';
};

const formatDate = (dateStr: string): string => {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
};

const getUniqueClients = (account: Account): any[] => {
  if (!account) return [];
  const allClients = [...(account.clients || []), ...(account.client_users || [])];
  // Deduplicate by email
  const seen = new Set();
  return allClients.filter(client => {
    const key = client.email || client.ID;
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
};

const formatBillingFrequency = (frequency: string): string => {
  const frequencies: Record<string, string> = {
    'BILLING_TYPE_WEEKLY': 'Weekly',
    'BILLING_TYPE_BIWEEKLY': 'Bi-Weekly',
    'BILLING_TYPE_MONTHLY': 'Monthly',
    'BILLING_TYPE_BIMONTHLY': 'Bi-Monthly',
    'BILLING_TYPE_PROJECT': 'Project-Based'
  };
  return frequencies[frequency] || frequency;
};

const formatRateType = (type: string): string => {
  const types: Record<string, string> = {
    'RATE_TYPE_EXTERNAL_CLIENT_BILLABLE': 'Client Billable',
    'RATE_TYPE_EXTERNAL_CLIENT_NON_BILLABLE': 'Client Non-Billable',
    'RATE_TYPE_INTERNAL': 'Internal',
    'RATE_TYPE_UNASSIGNED': 'Unassigned'
  };
  return types[type] || type;
};

const getRateDetails = (rateId: number, internalRateId?: number) => {
  const rate = rates.value.find(r => r.ID === rateId);
  if (!rate) return { value: 0, internalValue: 0, margin: 0 };
  
  const value = rate.amount || 0;
  
  // Get internal rate amount if internal rate ID is provided
  let internalValue = 0;
  if (internalRateId) {
    const internalRate = rates.value.find(r => r.ID === internalRateId);
    internalValue = internalRate?.amount || 0;
  }
  
  const margin = value > 0 ? ((value - internalValue) / value) * 100 : 0;
  
  return { value, internalValue, margin };
};

const formatBudgetPeriod = (frequency: string): string => {
  switch (frequency) {
    case 'BILLING_TYPE_WEEKLY':
      return '/wk';
    case 'BILLING_TYPE_BIWEEKLY':
      return '/2wk';
    case 'BILLING_TYPE_MONTHLY':
      return '/mo';
    case 'BILLING_TYPE_BIMONTHLY':
      return '/2mo';
    case 'BILLING_TYPE_PROJECT':
      return ' total';
    default:
      return '';
  }
};

const calculateDuration = (startDate: string, endDate: string): string => {
  const start = new Date(startDate);
  const end = new Date(endDate);
  const diffTime = Math.abs(end.getTime() - start.getTime());
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  
  if (diffDays < 30) {
    return `${diffDays} days`;
  } else if (diffDays < 365) {
    const months = Math.floor(diffDays / 30);
    return `${months} month${months !== 1 ? 's' : ''}`;
  } else {
    const years = Math.floor(diffDays / 365);
    const remainingMonths = Math.floor((diffDays % 365) / 30);
    if (remainingMonths > 0) {
      return `${years}y ${remainingMonths}m`;
    }
    return `${years} year${years !== 1 ? 's' : ''}`;
  }
};

const getAssetIcon = (assetType: string, contentType?: string | null): string => {
  // Check asset_type first (simple string values like 'pdf', 'excel', etc.)
  const typeMap: Record<string, string> = {
    'pdf': 'fas fa-file-pdf',
    'image': 'fas fa-image',
    'png': 'fas fa-image',
    'jpeg': 'fas fa-image',
    'jpg': 'fas fa-image',
    'excel': 'fas fa-file-excel',
    'xlsx': 'fas fa-file-excel',
    'csv': 'fas fa-file-csv',
    'google_doc': 'fab fa-google-drive',
    'google_sheet': 'fas fa-table',
    'google_slides': 'fas fa-file-powerpoint',
    'docx': 'fas fa-file-word',
    'doc': 'fas fa-file-word',
    'video': 'fas fa-video',
    'link': 'fas fa-link',
    'external_link': 'fas fa-link',
    'file': 'fas fa-file'
  };
  
  // Check content_type as fallback (MIME types)
  if (contentType && !typeMap[assetType.toLowerCase()]) {
    if (contentType.includes('pdf')) return 'fas fa-file-pdf';
    if (contentType.includes('image')) return 'fas fa-image';
    if (contentType.includes('spreadsheet') || contentType.includes('excel')) return 'fas fa-file-excel';
    if (contentType.includes('document') || contentType.includes('word')) return 'fas fa-file-word';
    if (contentType.includes('presentation') || contentType.includes('powerpoint')) return 'fas fa-file-powerpoint';
    if (contentType.includes('google-apps')) return 'fab fa-google-drive';
  }
  
  return typeMap[assetType.toLowerCase()] || 'fas fa-file';
};

const getAssetIconColor = (assetType: string, contentType?: string | null): string => {
  const lowerType = assetType.toLowerCase();
  
  // Images - Purple
  if (['image', 'png', 'jpeg', 'jpg', 'gif', 'webp'].includes(lowerType) || contentType?.includes('image')) {
    return 'text-purple-500 group-hover:text-purple-600';
  }
  
  // PDFs - Red
  if (lowerType === 'pdf' || contentType?.includes('pdf')) {
    return 'text-red-500 group-hover:text-red-600';
  }
  
  // Excel/Spreadsheets - Green
  if (['excel', 'xlsx', 'csv'].includes(lowerType) || contentType?.includes('spreadsheet') || contentType?.includes('excel')) {
    return 'text-green-600 group-hover:text-green-700';
  }
  
  // Word/Documents - Blue
  if (['docx', 'doc', 'google_doc'].includes(lowerType) || contentType?.includes('document') || contentType?.includes('word')) {
    return 'text-blue-600 group-hover:text-blue-700';
  }
  
  // Presentations - Orange
  if (['google_slides', 'ppt', 'pptx'].includes(lowerType) || contentType?.includes('presentation') || contentType?.includes('powerpoint')) {
    return 'text-orange-500 group-hover:text-orange-600';
  }
  
  // Videos - Pink
  if (lowerType === 'video' || contentType?.includes('video')) {
    return 'text-pink-500 group-hover:text-pink-600';
  }
  
  // Links - Teal
  if (['link', 'external_link'].includes(lowerType)) {
    return 'text-teal-500 group-hover:text-teal-600';
  }
  
  // Google Drive - Multi-color (use their brand colors)
  if (lowerType === 'google_doc' || lowerType === 'google_sheet' || contentType?.includes('google-apps')) {
    return 'text-blue-500 group-hover:text-blue-600';
  }
  
  // Default - Gray
  return 'text-gray-400 group-hover:text-sage';
};

const getStaffName = (staffId: number): string => {
  const staff = staffMembers.value.find(s => s.ID === staffId);
  if (staff) {
    return `${staff.first_name || ''} ${staff.last_name || ''}`.trim() || staff.email || 'Unknown';
  }
  return 'Unknown';
};

// Watch for route changes
watch(() => route.query, () => {
  if (route.name === 'organization') {
    parseURLParams();
  }
});

// Load data on mount
onMounted(() => {
  loadData();
});
</script>


