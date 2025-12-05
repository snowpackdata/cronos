<script setup lang="ts">
import { ref, reactive, defineProps, defineEmits, watch } from 'vue';
import type { Account } from '../../types/Account';
import { ACCOUNT_TYPES, BILLING_FREQUENCIES } from '../../types/Account';
import accountsApi from '../../api/accounts';

// Define props and emits
const props = defineProps<{
  account: Account;
  isNew: boolean;
}>();

const emit = defineEmits<{
  (e: 'save', account: Account): void;
  (e: 'delete', account: Account): void;
  (e: 'cancel'): void;
}>();

// Create a reactive copy of the account to avoid mutating props directly
const formData = reactive<Account>({ ...props.account });

// State for form handling
const isLoading = ref(false);
const error = ref<string | null>(null);
const inviteEmail = ref('');
const inviteLoading = ref(false);
const inviteError = ref<string | null>(null);
const inviteSuccess = ref(false);

// Watch for changes in the account prop
watch(() => props.account, (newAccount) => {
  Object.assign(formData, newAccount);
}, { deep: true });

// Save the account
const saveAccount = async () => {
  isLoading.value = true;
  error.value = null;
  
  try {
    if (props.isNew) {
      const createdAccount = await accountsApi.createAccount(formData);
      emit('save', createdAccount);
    } else {
      const updatedAccount = await accountsApi.updateAccount(formData);
      emit('save', updatedAccount);
    }
  } catch (err) {
    console.error('Error saving account:', err);
    error.value = 'Failed to save account. Please try again.';
  } finally {
    isLoading.value = false;
  }
};

// Delete the account
const deleteAccount = async () => {
  if (!confirm('Are you sure you want to delete this account?')) {
    return;
  }
  
  isLoading.value = true;
  error.value = null;
  
  try {
    await accountsApi.deleteAccount(formData.ID);
    emit('delete', formData);
  } catch (err) {
    console.error('Error deleting account:', err);
    error.value = 'Failed to delete account. Please try again.';
  } finally {
    isLoading.value = false;
  }
};

// Invite a user to the account
const inviteUser = async () => {
  if (!inviteEmail.value) {
    inviteError.value = 'Email is required';
    return;
  }
  
  inviteLoading.value = true;
  inviteError.value = null;
  inviteSuccess.value = false;
  
  try {
    const client = await accountsApi.inviteUser(formData.ID, inviteEmail.value);
    
    // Add the new client to the clients array if it doesn't exist already
    if (!formData.clients) {
      formData.clients = [];
    }
    
    const existingClientIndex = formData.clients.findIndex(c => c.email === client.email);
    if (existingClientIndex === -1) {
      formData.clients.push(client);
    }
    
    inviteEmail.value = '';
    inviteSuccess.value = true;
  } catch (err) {
    console.error('Error inviting user:', err);
    inviteError.value = 'Failed to invite user. Please try again.';
  } finally {
    inviteLoading.value = false;
  }
};

// Cancel editing
const cancel = () => {
  emit('cancel');
};
</script>

<template>
  <div class="account-form">
    <h3>{{ isNew ? 'Create New Account' : 'Edit Account' }}</h3>
    
    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-circle"></i> {{ error }}
    </div>
    
    <form @submit.prevent="saveAccount">
      <div class="form-group">
        <label for="name">Account Name*</label>
        <input 
          id="name" 
          v-model="formData.name" 
          type="text" 
          required 
          placeholder="Account Name"
        >
      </div>
      
      <div class="form-group">
        <label for="type">Account Type*</label>
        <select id="type" v-model="formData.type" required>
          <option v-for="type in ACCOUNT_TYPES" :key="type" :value="type">
            {{ type }}
          </option>
        </select>
      </div>
      
      <div class="form-group">
        <label for="legal_name">Legal Name</label>
        <input 
          id="legal_name" 
          v-model="formData.legal_name" 
          type="text" 
          placeholder="Legal Name"
        >
      </div>
      
      <div class="form-group">
        <label for="address">Address</label>
        <textarea 
          id="address" 
          v-model="formData.address" 
          placeholder="Address"
        ></textarea>
      </div>
      
      <div class="form-group">
        <label for="website">Website</label>
        <input 
          id="website" 
          v-model="formData.website" 
          type="url" 
          placeholder="https://example.com"
        >
      </div>
      
      <div class="form-group">
        <label for="email">Email</label>
        <input 
          id="email" 
          v-model="formData.email" 
          type="email" 
          placeholder="contact@example.com"
        >
      </div>
      
      <div class="form-group">
        <label for="billing_frequency">Billing Frequency*</label>
        <select id="billing_frequency" v-model="formData.billing_frequency" required>
          <option v-for="frequency in BILLING_FREQUENCIES" :key="frequency" :value="frequency">
            {{ frequency }}
          </option>
        </select>
      </div>
      
      <div class="form-row">
        <div class="form-group">
          <label for="budget_hours">Budget Hours</label>
          <input 
            id="budget_hours" 
            v-model="formData.budget_hours" 
            type="number" 
            min="0" 
            step="0.01"
            placeholder="0.00"
          >
        </div>
        
        <div class="form-group">
          <label for="budget_dollars">Budget Dollars</label>
          <input 
            id="budget_dollars" 
            v-model="formData.budget_dollars" 
            type="number" 
            min="0" 
            step="0.01"
            placeholder="0.00"
          >
        </div>
      </div>
      
      <div class="form-group checkbox-group">
        <input 
          id="projects_single_invoice" 
          v-model="formData.projects_single_invoice" 
          type="checkbox"
        >
        <label for="projects_single_invoice">Projects Single Invoice</label>
      </div>
      
      <div class="form-actions">
        <button 
          type="submit" 
          class="save-button" 
          :disabled="isLoading"
        >
          <i class="fas fa-save"></i> {{ isLoading ? 'Saving...' : 'Save' }}
        </button>
        
        <button 
          v-if="!isNew" 
          type="button" 
          class="delete-button" 
          @click="deleteAccount" 
          :disabled="isLoading"
        >
          <i class="fas fa-trash-alt"></i> Delete
        </button>
        
        <button 
          type="button" 
          class="cancel-button" 
          @click="cancel" 
          :disabled="isLoading"
        >
          <i class="fas fa-times"></i> Cancel
        </button>
      </div>
    </form>
    
    <!-- Invite User Section (only for existing accounts) -->
    <div v-if="!isNew" class="invite-section">
      <h4>Invite User</h4>
      
      <div v-if="inviteError" class="error-message">
        <i class="fas fa-exclamation-circle"></i> {{ inviteError }}
      </div>
      
      <div v-if="inviteSuccess" class="success-message">
        <i class="fas fa-check-circle"></i> User invited successfully!
      </div>
      
      <div class="invite-form">
        <input 
          v-model="inviteEmail" 
          type="email" 
          placeholder="Email address" 
          :disabled="inviteLoading"
        >
        <button 
          @click="inviteUser" 
          :disabled="inviteLoading || !inviteEmail"
        >
          <i class="fas fa-user-plus"></i> {{ inviteLoading ? 'Inviting...' : 'Invite' }}
        </button>
      </div>
      
      <!-- List of invited users -->
      <div v-if="formData.clients && formData.clients.length > 0" class="clients-list">
        <h5>Invited Users</h5>
        <ul>
          <li v-for="client in formData.clients" :key="client.email">
            <i class="fas fa-user"></i> {{ client.email }}
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<style scoped>
.account-form {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  padding: 25px;
}

h3 {
  margin-top: 0;
  margin-bottom: 20px;
  color: #333;
}

.error-message,
.success-message {
  padding: 10px;
  margin-bottom: 15px;
  border-radius: 4px;
}

.error-message {
  background-color: #f8d7da;
  color: #721c24;
}

.success-message {
  background-color: #d4edda;
  color: #155724;
}

.form-group {
  margin-bottom: 15px;
}

.form-row {
  display: flex;
  gap: 15px;
}

.form-row .form-group {
  flex: 1;
}

label {
  display: block;
  margin-bottom: 5px;
  font-weight: 500;
  color: #555;
}

input[type="text"],
input[type="email"],
input[type="url"],
input[type="number"],
select,
textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
}

textarea {
  min-height: 80px;
  resize: vertical;
}

.checkbox-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.checkbox-group label {
  margin-bottom: 0;
}

.checkbox-group input[type="checkbox"] {
  margin: 0;
}

.form-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
}

button {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 5px;
}

button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.save-button {
  background-color: #28a745;
  color: white;
}

.delete-button {
  background-color: #dc3545;
  color: white;
}

.cancel-button {
  background-color: #6c757d;
  color: white;
}

.invite-section {
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid #eee;
}

.invite-section h4 {
  margin-top: 0;
  margin-bottom: 15px;
}

.invite-form {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
}

.invite-form input {
  flex: 1;
}

.invite-form button {
  background-color: #007bff;
  color: white;
}

.clients-list {
  margin-top: 20px;
}

.clients-list h5 {
  margin-top: 0;
  margin-bottom: 10px;
  font-size: 1rem;
}

.clients-list ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.clients-list li {
  padding: 5px 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.clients-list i {
  color: #007bff;
}
</style> 