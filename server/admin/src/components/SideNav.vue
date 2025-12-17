  <router-link 
    to="/accounts" 
    class="flex items-center px-2 py-2 text-base font-medium rounded-md"
    :class="{ 'text-white bg-sage': isActive('/accounts'), 'text-gray-300 hover:bg-black/30 hover:text-white': !isActive('/accounts') }"
  >
    <i class="fas fa-building mr-3 text-xl"></i>
    Accounts
  </router-link>
  
  <router-link 
    to="/settings" 
    class="flex items-center px-2 py-2 text-base font-medium rounded-md"
    :class="{ 'text-white bg-sage': isActive('/settings'), 'text-gray-300 hover:bg-black/30 hover:text-white': !isActive('/settings') }"
  >
    <i class="fas fa-cog mr-3 text-xl"></i>
    Settings
  </router-link>
  
</nav> 