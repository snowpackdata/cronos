<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getJournals, getAccountBalances } from '../../api/journals';
import type { Journal, BalanceSummary, AccountBalance } from '../../types/Journal';
import { formatCurrency, formatAccountName, getAccountCategory } from '../../types/Journal';

// State
const journals = ref<Journal[]>([]);
const balanceSummary = ref<BalanceSummary | null>(null);
const priorBalanceSummary = ref<BalanceSummary | null>(null);
const cumulativeBalanceSummary = ref<BalanceSummary | null>(null); // For Balance Sheet (inception → end date)
const beginningBalanceSummary = ref<BalanceSummary | null>(null); // For beginning cash (inception → day before start)
const isLoading = ref(true);
const error = ref<string | null>(null);

// Filters
const startDate = ref<string>(new Date(new Date().getFullYear(), 0, 1).toISOString().split('T')[0]); // Jan 1 of current year
const endDate = ref<string>(new Date().toISOString().split('T')[0]); // Today
const searchSubAccount = ref('');
const selectedAccount = ref<string>('');

// Expansion state
const expandedAccounts = ref<Set<string>>(new Set());
const expandedTrialBalanceAccounts = ref<Set<string>>(new Set());

// Manual journal entry modal
const showManualEntryModal = ref(false);
const manualEntryLines = ref<Array<{
  account: string;
  subaccount: string;
  debit: number;
  credit: number;
  memo: string;
}>>([
  { account: '', subaccount: '', debit: 0, credit: 0, memo: '' },
  { account: '', subaccount: '', debit: 0, credit: 0, memo: '' }
]);
const manualEntryDate = ref(new Date().toISOString().split('T')[0]);
const manualEntryError = ref<string | null>(null);
const isSubmittingManualEntry = ref(false);

// Chart of Accounts and Subaccounts for manual entry
const availableAccounts = ref<Array<{ code: string; name: string }>>([]);
const availableSubaccounts = ref<Array<{ code: string; name: string; account_code: string }>>([]);
const isLoadingAccountOptions = ref(false);

// Adjustment/Reversal modal
const adjustmentModalOpen = ref(false);
const adjustingJournal = ref<Journal | null>(null);
const adjustmentReason = ref('');
const createCorrectedEntry = ref(false);
const correctedForm = ref({
  account: '',
  sub_account: '',
  memo: '',
  debit: 0,
  credit: 0,
});

// Fetch data
async function fetchData() {
  isLoading.value = true;
  error.value = null;
  
  try {
    // Calculate prior period dates (same duration, ending day before current start)
    const start = new Date(startDate.value);
    const end = new Date(endDate.value);
    const durationMs = end.getTime() - start.getTime();
    const priorEnd = new Date(start.getTime() - 24 * 60 * 60 * 1000); // Day before start
    const priorStart = new Date(priorEnd.getTime() - durationMs);
    
    const priorStartStr = priorStart.toISOString().split('T')[0];
    const priorEndStr = priorEnd.toISOString().split('T')[0];
    
    // Calculate day before start for beginning balance
    const dayBeforeStart = new Date(start.getTime() - 24 * 60 * 60 * 1000);
    const dayBeforeStartStr = dayBeforeStart.toISOString().split('T')[0];
    
    // Fetch journals from the official GL only (offline journals are managed separately)
    // balancesData: Period-based for P&L (start → end)
    // cumulativeBalancesData: Cumulative for Balance Sheet (inception → end)
    // beginningBalancesData: Cumulative for beginning cash (inception → day before start)
    const [journalsData, balancesData, priorBalancesData, cumulativeBalancesData, beginningBalancesData] = await Promise.all([
      getJournals({ 
        start_date: startDate.value, 
        end_date: endDate.value,
        include_offline: false 
      }),
      getAccountBalances(startDate.value, endDate.value), // Period for P&L
      getAccountBalances(priorStartStr, priorEndStr),
      getAccountBalances(undefined, endDate.value), // Cumulative (inception → end) for Balance Sheet
      getAccountBalances(undefined, dayBeforeStartStr) // Cumulative (inception → day before start) for beginning cash
    ]);
    
    console.log('Journals received:', journalsData?.length, 'journals');
    console.log('Period balances (for P&L):', startDate.value, '→', endDate.value);
    console.log('Cumulative balances (for Balance Sheet): inception →', endDate.value);
    console.log('Beginning balances (for Cash Flow): inception →', dayBeforeStartStr);
    journals.value = journalsData || [];
    balanceSummary.value = balancesData;
    priorBalanceSummary.value = priorBalancesData;
    cumulativeBalanceSummary.value = cumulativeBalancesData;
    beginningBalanceSummary.value = beginningBalancesData;
    
    // Pre-load account options on first load
    if (availableAccounts.value.length === 0) {
      loadAccountOptions();
    }
  } catch (err) {
    console.error('Error fetching accounting data:', err);
    error.value = 'Failed to load accounting data. Please try again.';
    journals.value = [];
    balanceSummary.value = null;
    priorBalanceSummary.value = null;
    cumulativeBalanceSummary.value = null;
    beginningBalanceSummary.value = null;
  } finally {
    isLoading.value = false;
  }
}

// Get filtered journals
function getFilteredJournals(): Journal[] {
  return journals.value.filter(journal => {
    if (searchSubAccount.value && !journal.sub_account.toLowerCase().includes(searchSubAccount.value.toLowerCase())) {
      return false;
    }
    if (selectedAccount.value && journal.account !== selectedAccount.value) {
      return false;
    }
    return true;
  });
}

// Group journals by account (sorted alphabetically)
function getJournalsByAccount(): [string, Journal[]][] {
  const grouped = new Map<string, Journal[]>();
  const filtered = getFilteredJournals();
  
  filtered.forEach(journal => {
    if (!grouped.has(journal.account)) {
      grouped.set(journal.account, []);
    }
    grouped.get(journal.account)!.push(journal);
  });
  
  // Convert to array and sort alphabetically by account name
  return Array.from(grouped.entries()).sort((a, b) => a[0].localeCompare(b[0]));
}

// Group accounts by category
function getAccountsByCategory(): Map<string, AccountBalance[]> {
  const grouped = new Map<string, AccountBalance[]>();
  
  if (!balanceSummary.value || !balanceSummary.value.accounts) {
    return grouped;
  }
  
  // Filter accounts based on selectedAccount filter
  let accountsToShow = balanceSummary.value.accounts;
  if (selectedAccount.value) {
    accountsToShow = accountsToShow.filter(account => account.account === selectedAccount.value);
  }
  
  accountsToShow.forEach(account => {
    const category = getAccountCategory(account.account, account.account_type);
    if (!grouped.has(category)) {
      grouped.set(category, []);
    }
    grouped.get(category)!.push(account);
  });
  
  return grouped;
}

// Get unique accounts for filter dropdown
function getUniqueAccounts(): string[] {
  return Array.from(new Set(journals.value.map(j => j.account))).sort();
}

// Toggle account expansion (for journal entries)
function toggleAccount(account: string) {
  if (expandedAccounts.value.has(account)) {
    expandedAccounts.value.delete(account);
  } else {
    expandedAccounts.value.add(account);
  }
}

// Toggle account expansion in Trial Balance (to show subaccounts)
function toggleTrialBalanceAccount(account: string) {
  if (expandedTrialBalanceAccounts.value.has(account)) {
    expandedTrialBalanceAccounts.value.delete(account);
  } else {
    expandedTrialBalanceAccounts.value.add(account);
  }
}

// Check if account is expanded in Trial Balance
function isTrialBalanceAccountExpanded(account: string): boolean {
  return expandedTrialBalanceAccounts.value.has(account);
}

// Get subaccount balances for a specific account
function getSubaccountBalances(accountName: string): Array<{
  subaccount: string;
  total_debits: number;
  total_credits: number;
  net_balance: number;
}> {
  const subaccountMap = new Map<string, { debits: number; credits: number }>();
  
  // Filter journals for this specific account
  const accountJournals = journals.value.filter(j => j.account === accountName);
  
  // Aggregate by subaccount
  accountJournals.forEach(journal => {
    const subaccount = journal.sub_account || 'No Subaccount';
    if (!subaccountMap.has(subaccount)) {
      subaccountMap.set(subaccount, { debits: 0, credits: 0 });
    }
    const entry = subaccountMap.get(subaccount)!;
    entry.debits += journal.debit;
    entry.credits += journal.credit;
  });
  
  // Convert to array and calculate net balance
  let subaccounts = Array.from(subaccountMap.entries())
    .map(([subaccount, amounts]) => ({
      subaccount,
      total_debits: amounts.debits,
      total_credits: amounts.credits,
      net_balance: amounts.debits - amounts.credits
    }))
    .sort((a, b) => Math.abs(b.net_balance) - Math.abs(a.net_balance)); // Sort by absolute balance, largest first
  
  // Filter by searchSubAccount if provided
  if (searchSubAccount.value) {
    const search = searchSubAccount.value.toLowerCase();
    subaccounts = subaccounts.filter(sub => 
      sub.subaccount.toLowerCase().includes(search)
    );
  }
  
  return subaccounts;
}

// Get journals for a specific account and subaccount
// Aggregate journal entries by date + account + subaccount for cleaner display
// Preserves individual entries in the database (immutable audit trail)
// but groups them for UI readability
function aggregateJournals(journalList: Journal[]): Journal[] {
  const grouped = new Map<string, Journal>();
  
  journalList.forEach(journal => {
    // Group by: date + account + subaccount + memo pattern
    const date = new Date(journal.CreatedAt).toISOString().split('T')[0]; // YYYY-MM-DD
    const account = journal.account || '';
    const subAccount = journal.sub_account || 'No Subaccount';
    
    // Determine if this is a payroll accrual (should be aggregated)
    const isPayrollAccrual = journal.memo?.includes('Payroll accrual for approved entries');
    const memoPattern = isPayrollAccrual ? 'payroll-accrual' : journal.memo;
    
    const key = `${date}-${account}-${subAccount}-${memoPattern}`;
    
    if (grouped.has(key) && isPayrollAccrual) {
      // Aggregate: sum debits and credits
      const existing = grouped.get(key)!;
      existing.debit += journal.debit;
      existing.credit += journal.credit;
      
      // Mark as aggregated
      (existing as any).isAggregated = true;
      
      // Update memo to show it's aggregated
      const billMatch = journal.memo?.match(/bill #(\d+)/);
      const billId = billMatch ? billMatch[1] : '';
      const existingCount = existing.memo?.match(/(\d+) approvals/) || [null, '1'];
      const newCount = parseInt(existingCount[1]) + 1;
      existing.memo = `Payroll accrual for approved entries (bill #${billId}) - ${newCount} approvals`;
    } else {
      // New entry: add to map
      grouped.set(key, {
        ...journal,
        debit: journal.debit,
        credit: journal.credit,
        isAggregated: false
      } as any);
    }
  });
  
  return Array.from(grouped.values());
}

function getJournalsForSubaccount(accountName: string, subaccount: string): Journal[] {
  const filtered = journals.value.filter(j => 
    j.account === accountName && 
    (j.sub_account || 'No Subaccount') === subaccount
  );
  
  // Aggregate before displaying
  return aggregateJournals(filtered);
}

// Format date
function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('en-US', {
    month: '2-digit',
    day: '2-digit',
    year: 'numeric'
  });
}

// Clear filters
function clearFilters() {
  searchSubAccount.value = '';
  selectedAccount.value = '';
}

// Fetch clients and staff for subaccount suggestions
async function loadAccountOptions() {
  if (isLoadingAccountOptions.value) return;
  if (availableAccounts.value.length > 0) return; // Already loaded
  
  isLoadingAccountOptions.value = true;
  console.log('Loading Chart of Accounts and Subaccounts...');
  
  try {
    const token = localStorage.getItem('snowpack_token');
    
    // Fetch Chart of Accounts
    const accountsResponse = await fetch('/api/cronos/chart-of-accounts?active_only=true', {
      headers: { 'x-access-token': token || '' }
    });
    if (accountsResponse.ok) {
      const accounts = await accountsResponse.json();
      availableAccounts.value = accounts.map((a: any) => ({
        code: a.account_code,
        name: a.account_name
      }));
      console.log(`Loaded ${availableAccounts.value.length} accounts`);
    } else {
      console.error('Failed to load chart of accounts:', accountsResponse.status);
    }
    
    // Fetch Subaccounts
    const subaccountsResponse = await fetch('/api/cronos/subaccounts?active_only=true', {
      headers: { 'x-access-token': token || '' }
    });
    if (subaccountsResponse.ok) {
      const subaccounts = await subaccountsResponse.json();
      availableSubaccounts.value = subaccounts.map((s: any) => ({
        code: s.code,
        name: s.name,
        account_code: s.account_code
      }));
      console.log(`Loaded ${availableSubaccounts.value.length} subaccounts`);
    } else {
      console.error('Failed to load subaccounts:', subaccountsResponse.status);
    }
  } catch (err) {
    console.error('Error loading account options:', err);
  } finally {
    isLoadingAccountOptions.value = false;
  }
}

// Get subaccount suggestions based on selected account
function getSubaccountSuggestions(account: string): Array<{ value: string; label: string }> {
  if (!account) return [];
  
  // Filter subaccounts by the selected account
  const filtered = availableSubaccounts.value.filter(sub => sub.account_code === account);
  
  return filtered.map(sub => ({
    value: sub.code,
    label: `${sub.code} (${sub.name})`
  }));
}

// Manual journal entry functions
async function openManualEntryModal() {
  showManualEntryModal.value = true;
  manualEntryLines.value = [
    { account: '', subaccount: '', debit: 0, credit: 0, memo: '' },
    { account: '', subaccount: '', debit: 0, credit: 0, memo: '' }
  ];
  manualEntryDate.value = new Date().toISOString().split('T')[0];
  manualEntryError.value = null;
  
  // Pre-load subaccount options
  await loadAccountOptions();
}

function addManualEntryLine() {
  manualEntryLines.value.push({ account: '', subaccount: '', debit: 0, credit: 0, memo: '' });
}

function removeManualEntryLine(index: number) {
  if (manualEntryLines.value.length > 2) {
    manualEntryLines.value.splice(index, 1);
  }
}

function calculateTotalDebits(): number {
  return manualEntryLines.value.reduce((sum, line) => sum + (line.debit || 0), 0);
}

function calculateTotalCredits(): number {
  return manualEntryLines.value.reduce((sum, line) => sum + (line.credit || 0), 0);
}

function isBalanced(): boolean {
  const totalDebits = calculateTotalDebits();
  const totalCredits = calculateTotalCredits();
  return Math.abs(totalDebits - totalCredits) < 0.01; // Allow for rounding
}

async function submitManualEntry() {
  manualEntryError.value = null;
  
  // Validation
  if (!isBalanced()) {
    manualEntryError.value = 'Entry must balance: Total debits must equal total credits';
    return;
  }
  
  if (calculateTotalDebits() === 0) {
    manualEntryError.value = 'Entry cannot have zero amounts';
    return;
  }
  
  // Check all lines have required fields
  for (const line of manualEntryLines.value) {
    if (!line.account) {
      manualEntryError.value = 'All lines must have an account selected';
      return;
    }
    if (line.debit === 0 && line.credit === 0) {
      manualEntryError.value = 'Each line must have either a debit or credit amount';
      return;
    }
    if (line.debit > 0 && line.credit > 0) {
      manualEntryError.value = 'Each line cannot have both debit and credit amounts';
      return;
    }
  }
  
  isSubmittingManualEntry.value = true;
  
  try {
    const token = localStorage.getItem('snowpack_token');
    const response = await fetch('/api/cronos/journals/manual', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-access-token': token || ''
      },
      body: JSON.stringify({
        date: manualEntryDate.value,
        lines: manualEntryLines.value.map(line => ({
          account: line.account,
          sub_account: line.subaccount,
          debit: Math.round(line.debit * 100), // Convert to cents
          credit: Math.round(line.credit * 100), // Convert to cents
          memo: line.memo
        }))
      })
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(errorData.error || 'Failed to create manual journal entry');
    }
    
    // Success - refresh data and close modal
    showManualEntryModal.value = false;
    await fetchData();
  } catch (err) {
    console.error('Error creating manual journal entry:', err);
    manualEntryError.value = err instanceof Error ? err.message : 'Failed to create manual journal entry';
  } finally {
    isSubmittingManualEntry.value = false;
  }
}

// Adjustment/Reversal modal functions
function openAdjustmentModal(journal: Journal) {
  adjustingJournal.value = journal;
  adjustmentReason.value = '';
  createCorrectedEntry.value = false;
  correctedForm.value = {
    account: journal.account,
    sub_account: journal.sub_account || '',
    memo: journal.memo || '',
    debit: journal.debit / 100, // Convert cents to dollars
    credit: journal.credit / 100,
  };
  adjustmentModalOpen.value = true;
  
  // Pre-load account options if not already loaded
  loadAccountOptions();
}

function closeAdjustmentModal() {
  adjustmentModalOpen.value = false;
  adjustingJournal.value = null;
}

async function submitAdjustment() {
  if (!adjustingJournal.value || !adjustmentReason.value.trim()) return;
  
  try {
    const token = localStorage.getItem('snowpack_token');
    const payload: any = {
      reason: adjustmentReason.value.trim(),
      create_corrected: createCorrectedEntry.value,
    };
    
    if (createCorrectedEntry.value) {
      payload.corrected = {
        account: correctedForm.value.account,
        sub_account: correctedForm.value.sub_account,
        memo: correctedForm.value.memo,
        debit: correctedForm.value.debit,
        credit: correctedForm.value.credit,
      };
    }
    
    const response = await fetch(`/api/cronos/journals/${adjustingJournal.value.ID}/reverse`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-access-token': token || ''
      },
      body: JSON.stringify(payload)
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(errorData.error || 'Failed to reverse journal entry');
    }
    
    closeAdjustmentModal();
    await fetchData();
  } catch (err) {
    console.error('Error reversing journal entry:', err);
    alert('Failed to reverse journal entry: ' + (err instanceof Error ? err.message : 'Unknown error'));
  }
}

// Quick date preset functions
function setThisMonth() {
  const now = new Date();
  startDate.value = new Date(now.getFullYear(), now.getMonth(), 1).toISOString().split('T')[0];
  endDate.value = new Date().toISOString().split('T')[0];
  fetchData();
}

function setLastMonth() {
  const now = new Date();
  const lastMonth = new Date(now.getFullYear(), now.getMonth() - 1, 1);
  startDate.value = lastMonth.toISOString().split('T')[0];
  const lastDayOfLastMonth = new Date(now.getFullYear(), now.getMonth(), 0);
  endDate.value = lastDayOfLastMonth.toISOString().split('T')[0];
  fetchData();
}

function setThisQuarter() {
  const now = new Date();
  const quarter = Math.floor(now.getMonth() / 3);
  startDate.value = new Date(now.getFullYear(), quarter * 3, 1).toISOString().split('T')[0];
  endDate.value = new Date().toISOString().split('T')[0];
  fetchData();
}

function setYearToDate() {
  const now = new Date();
  startDate.value = new Date(now.getFullYear(), 0, 1).toISOString().split('T')[0];
  endDate.value = new Date().toISOString().split('T')[0];
  fetchData();
}

function setLastYear() {
  const now = new Date();
  startDate.value = new Date(now.getFullYear() - 1, 0, 1).toISOString().split('T')[0];
  endDate.value = new Date(now.getFullYear() - 1, 11, 31).toISOString().split('T')[0];
  fetchData();
}

// Get summary metrics
function getSummaryMetrics() {
  if (!balanceSummary.value || !balanceSummary.value.accounts) {
    return { cash: 0, ar: 0, ap: 0, netProfit: 0 };
  }

  let cash = 0;
  let ar = 0;
  let ap = 0;

  balanceSummary.value.accounts.forEach(account => {
    if (account.account === 'CASH') {
      cash = account.net_balance;
    } else if (account.account === 'ACCOUNTS_RECEIVABLE') {
      ar = account.net_balance;
    } else if (account.account === 'ACCOUNTS_PAYABLE') {
      ap = Math.abs(account.net_balance); // AP is negative, show as positive
    }
  });

  const netProfit = calculateProfitLoss().netIncome;

  return { cash, ar, ap, netProfit };
}

// Calculate Balance Sheet (using cumulative balances from inception → end date)
function calculateBalanceSheet() {
  // Use cumulative balances for Balance Sheet (inception → end date)
  if (!cumulativeBalanceSummary.value || !cumulativeBalanceSummary.value.accounts) {
    return { 
      assets: [], totalAssets: 0,
      liabilities: [], totalLiabilities: 0,
      equity: [], totalEquity: 0,
      partnerCapital: 0,
      ownerDistributions: 0,
      cumulativeNetIncome: 0,
      currentPeriodNetIncome: 0,
      retainedEarnings: 0,
      balances: false
    };
  }

  const assets: Array<{ account: string; balance: number }> = [];
  const liabilities: Array<{ account: string; balance: number }> = [];
  const equity: Array<{ account: string; balance: number }> = [];
  
  let totalAssets = 0;
  let totalLiabilities = 0;
  let totalEquity = 0;
  let ownerDistributions = 0;
  let partnerCapital = 0; // Track partner capital contributions separately
  
  // Use CUMULATIVE balances (from inception to end date)
  cumulativeBalanceSummary.value.accounts.forEach(account => {
    const accountName = account.account;
    const category = getAccountCategory(accountName, account.account_type);
    const balance = account.net_balance;

    if (category === 'Assets') {
      assets.push({ account: accountName, balance });
      totalAssets += balance;
    } else if (category === 'Liabilities') {
      // Keep the natural accounting sign (negative for credit balances)
      liabilities.push({ account: accountName, balance });
      totalLiabilities += balance;
      console.log(`Liability (cumulative): ${accountName}, Balance: ${balance}, Running Total: ${totalLiabilities}`);
    } else if (category === 'Equity') {
      if (accountName === 'OWNER_DISTRIBUTIONS') {
        // Owner distributions REDUCE equity (debit balance, cumulative)
        ownerDistributions = balance;
      } else if (accountName === 'EQUITY_OWNERSHIP') {
        // Partner capital contributions (initial investments)
        partnerCapital = balance;
      } else {
        equity.push({ account: accountName, balance });
        totalEquity += balance;
      }
    }
  });
  
  console.log(`Total Liabilities (cumulative): ${totalLiabilities}, Abs: ${Math.abs(totalLiabilities)}`);

  // Calculate cumulative net income from cumulative revenue and expenses
  // (all revenue/expenses from inception to end date)
  let cumulativeRevenue = 0;
  let cumulativeExpenses = 0;
  
  cumulativeBalanceSummary.value.accounts.forEach(account => {
    const accountName = account.account;
    
    // Revenue accounts (credits, negative balances)
    if (accountName === 'REVENUE' || 
        accountName === 'ADJUSTMENT_REVENUE' ||
        accountName === 'OTHER_INCOME') {
      cumulativeRevenue += account.net_balance;
    }
    
    // Expense accounts (debits, positive balances)
    if (accountName === 'PAYROLL_EXPENSE' || 
        accountName === 'ADJUSTMENT_EXPENSE' ||
        accountName.startsWith('OPERATING_EXPENSES_') ||
        accountName === 'EQUIPMENT_EXPENSE' ||
        accountName === 'OTHER_EXPENSES') {
      cumulativeExpenses += account.net_balance;
    }
  });
  
  // Cumulative net income (all-time)
  const cumulativeNetIncome = Math.abs(cumulativeRevenue) - Math.abs(cumulativeExpenses);
  
  // Get current period net income for display
  const pl = calculateProfitLoss();
  const currentPeriodNetIncome = pl.netIncome;
  
  // Partnership/LLC Equity Structure (using natural accounting signs):
  // Assets (DR, positive) + Liabilities (CR, negative) + Equity (CR, negative) = 0
  
  // Calculate total equity using natural signs:
  // - Partner capital: credit balance (negative)
  // - Cumulative net income: increases equity, so it's a credit (subtract since it's positive)
  // - Owner distributions: debit balance (positive), reduces equity (add to make equity less negative)
  const calculatedEquity = partnerCapital + totalEquity - cumulativeNetIncome + ownerDistributions;
  
  // Balance sheet equation check: Assets + Liabilities + Equity ≈ 0
  const balanceCheck = totalAssets + totalLiabilities + calculatedEquity;
  const balances = Math.abs(balanceCheck) < 0.01;
  
  console.log(`Cumulative Net Income (all-time): Revenue(${Math.abs(cumulativeRevenue)}) - Expenses(${Math.abs(cumulativeExpenses)}) = ${cumulativeNetIncome}`);
  console.log(`Balance Check: Assets(${totalAssets}) + Liabilities(${totalLiabilities}) + Equity(${calculatedEquity}) = ${balanceCheck}`);

  return {
    assets,
    totalAssets: Math.abs(totalAssets),
    liabilities,
    totalLiabilities: Math.abs(totalLiabilities),
    equity,
    partnerCapital: Math.abs(partnerCapital),
    ownerDistributions: Math.abs(ownerDistributions),
    cumulativeNetIncome,  // All-time net income for balance sheet
    currentPeriodNetIncome, // Current period for comparison
    totalEquity: Math.abs(calculatedEquity),
    retainedEarnings: cumulativeNetIncome, // Use cumulative for balance sheet
    balances
  };
}

// Calculate P&L
function calculateProfitLoss() {
  if (!balanceSummary.value || !balanceSummary.value.accounts) {
    return { revenue: 0, expenses: 0, netIncome: 0 };
  }

  let revenue = 0;
  let expenses = 0;

  balanceSummary.value.accounts.forEach(account => {
    const accountName = account.account;
    
    // Revenue accounts (credits increase revenue)
    if (accountName === 'REVENUE' || 
        accountName === 'ADJUSTMENT_REVENUE' ||
        accountName === 'OTHER_INCOME') {
      revenue += account.net_balance;
    }
    
    // Expense accounts (debits increase expenses)
    // Include all expense types from both Journal DB and Beancount
    // NOTE: OWNER_DISTRIBUTIONS is NOT an expense - it's an equity distribution
    if (accountName === 'PAYROLL_EXPENSE' || 
        accountName === 'ADJUSTMENT_EXPENSE' ||
        accountName.startsWith('OPERATING_EXPENSES_') ||
        accountName === 'EQUIPMENT_EXPENSE' ||
        accountName === 'OTHER_EXPENSES') {
      expenses += account.net_balance;
    }
  });

  // Convert to absolute values for display
  const revenueAmount = Math.abs(revenue);
  const expensesAmount = Math.abs(expenses);
  const netIncome = revenueAmount - expensesAmount;

  return { revenue: revenueAmount, expenses: expensesAmount, netIncome };
}

// Calculate prior period P&L
function calculatePriorProfitLoss() {
  if (!priorBalanceSummary.value || !priorBalanceSummary.value.accounts) {
    return { revenue: 0, expenses: 0, netIncome: 0 };
  }

  let revenue = 0;
  let expenses = 0;

  priorBalanceSummary.value.accounts.forEach(account => {
    const accountName = account.account;
    
    // Revenue accounts
    if (accountName === 'REVENUE' || 
        accountName === 'ADJUSTMENT_REVENUE' ||
        accountName === 'OTHER_INCOME') {
      revenue += account.net_balance;
    }
    
    // Expense accounts - include all types
    // NOTE: OWNER_DISTRIBUTIONS is NOT an expense - it's an equity distribution
    if (accountName === 'PAYROLL_EXPENSE' || 
        accountName === 'ADJUSTMENT_EXPENSE' ||
        accountName.startsWith('OPERATING_EXPENSES_') ||
        accountName === 'EQUIPMENT_EXPENSE' ||
        accountName === 'OTHER_EXPENSES') {
      expenses += account.net_balance;
    }
  });

  const revenueAmount = Math.abs(revenue);
  const expensesAmount = Math.abs(expenses);
  const netIncome = revenueAmount - expensesAmount;

  return { revenue: revenueAmount, expenses: expensesAmount, netIncome };
}

// Calculate percentage change
function calculatePercentChange(current: number, prior: number): { value: number, direction: 'up' | 'down' | 'flat' } {
  if (prior === 0) return { value: 0, direction: 'flat' };
  const change = ((current - prior) / prior) * 100;
  return {
    value: Math.abs(change),
    direction: change > 0 ? 'up' : change < 0 ? 'down' : 'flat'
  };
}

// Get source for a journal entry (helper for combined view)

// Calculate Cash Flow Statement
function calculateCashFlow() {
  if (!cumulativeBalanceSummary.value || !beginningBalanceSummary.value) {
    return { 
      operatingCash: 0, 
      netCashChange: 0, 
      beginningCash: 0, 
      endingCash: 0,
      collections: 0,
      payments: 0
    };
  }

  let collections = 0; // Cash collected from customers
  let payments = 0;    // Cash paid to employees/vendors
  let beginningCash = 0;
  let endingCash = 0;

  // Get beginning cash (cumulative as of day before start date)
  const beginningCashAccount = beginningBalanceSummary.value.accounts?.find(a => a.account === 'CASH');
  if (beginningCashAccount) {
    beginningCash = Math.abs(beginningCashAccount.net_balance);
  }

  // Get ending cash (cumulative as of end date)
  const endingCashAccount = cumulativeBalanceSummary.value.accounts?.find(a => a.account === 'CASH');
  if (endingCashAccount) {
    endingCash = Math.abs(endingCashAccount.net_balance);
  }

  // Calculate cash collections and payments from period journals
  const cashJournals = journals.value.filter(j => j.account === 'CASH');
  cashJournals.forEach(j => {
    if (j.debit > 0) collections += j.debit / 100; // Convert cents to dollars
    if (j.credit > 0) payments += j.credit / 100; // Convert cents to dollars
  });

  const netCashChange = endingCash - beginningCash;
  const operatingCash = netCashChange; // Simplified - all cash flow is from operations

  console.log(`Cash Flow: Beginning($${beginningCash}) + Change($${netCashChange}) = Ending($${endingCash})`);
  console.log(`  Collections: $${collections}, Payments: $${payments}`);

  return { 
    operatingCash, 
    netCashChange, 
    beginningCash, 
    endingCash,
    collections,
    payments
  };
}

// Calculate prior period Cash Flow
function calculatePriorCashFlow() {
  if (!priorBalanceSummary.value || !priorBalanceSummary.value.accounts) {
    return { 
      operatingCash: 0, 
      netCashChange: 0, 
      beginningCash: 0, 
      endingCash: 0,
      collections: 0,
      payments: 0
    };
  }

  let collections = 0;
  let payments = 0;
  let endingCash = 0;

  priorBalanceSummary.value.accounts.forEach(account => {
    if (account.account === 'CASH') {
      endingCash = account.net_balance;
    }
  });

  // For prior period, we'd need prior period journals, but we don't have those
  // So we'll estimate based on revenue/expenses
  priorBalanceSummary.value.accounts.forEach(account => {
    if (account.account === 'REVENUE' || account.account === 'ADJUSTMENT_REVENUE') {
      collections += Math.abs(account.net_balance);
    }
    if (account.account === 'PAYROLL_EXPENSE' || account.account === 'ADJUSTMENT_EXPENSE') {
      payments += Math.abs(account.net_balance);
    }
  });

  const netCashChange = collections - payments;
  const beginningCash = endingCash - netCashChange;
  const operatingCash = netCashChange;

  return { 
    operatingCash, 
    netCashChange, 
    beginningCash, 
    endingCash,
    collections,
    payments
  };
}

// Export to CSV
function exportToCSV() {
  const headers = ['Date', 'Account', 'SubAccount', 'Memo', 'Debit', 'Credit', 'Reference'];
  const rows = journals.value.map(j => [
    formatDate(j.CreatedAt),
    formatAccountName(j.account),
    j.sub_account || '',
    j.memo || '',
    j.debit > 0 ? (j.debit / 100).toFixed(2) : '',
    j.credit > 0 ? (j.credit / 100).toFixed(2) : '',
    j.invoice_id ? `INV-${j.invoice_id}` : j.bill_id ? `BILL-${j.bill_id}` : ''
  ]);

  const csvContent = [
    headers.join(','),
    ...rows.map(row => row.map(cell => `"${cell}"`).join(','))
  ].join('\n');

  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);
  link.setAttribute('href', url);
  link.setAttribute('download', `general-ledger-${startDate.value}-to-${endDate.value}.csv`);
  link.style.visibility = 'hidden';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

// Export Trial Balance to CSV
function exportTrialBalanceToCSV() {
  if (!balanceSummary.value || !balanceSummary.value.accounts) return;

  const headers = ['Account', 'Category', 'Total Debits', 'Total Credits', 'Net Balance'];
  const rows = balanceSummary.value.accounts.map(account => [
    formatAccountName(account.account),
    getAccountCategory(account.account, account.account_type),
    (account.total_debits / 100).toFixed(2),
    (account.total_credits / 100).toFixed(2),
    (account.net_balance / 100).toFixed(2)
  ]);

  const csvContent = [
    headers.join(','),
    ...rows.map(row => row.map(cell => `"${cell}"`).join(','))
  ].join('\n');

  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);
  link.setAttribute('href', url);
  link.setAttribute('download', `trial-balance-${startDate.value}-to-${endDate.value}.csv`);
  link.style.visibility = 'hidden';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

onMounted(() => {
  fetchData();
});
</script>

<template>
  <div class="p-4 bg-white min-h-screen">
    <!-- Header -->
    <div class="mb-3 pb-2 border-b-2 border-gray-900">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-xl font-bold text-gray-900 uppercase tracking-wide">General Ledger</h1>
          <p class="text-xs text-gray-600 mt-0.5">{{ formatDate(startDate) }} - {{ formatDate(endDate) }}</p>
        </div>
        <div class="flex gap-3 items-center">
          <!-- Beancount Toggle -->
          <button
            @click="openManualEntryModal"
            class="px-3 py-1.5 text-xs font-medium text-white bg-green-600 hover:bg-green-700 rounded flex items-center gap-1"
          >
            <i class="fas fa-plus"></i> Book Entry
          </button>
          <button
            @click="exportTrialBalanceToCSV"
            class="px-3 py-1.5 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded hover:bg-gray-50 flex items-center gap-1"
          >
            <i class="fas fa-download"></i> Export Trial Balance
          </button>
          <button
            @click="exportToCSV"
            class="px-3 py-1.5 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded hover:bg-gray-50 flex items-center gap-1"
          >
            <i class="fas fa-download"></i> Export Journals
          </button>
        </div>
      </div>
    </div>

    <!-- Summary Cards -->
    <div v-if="balanceSummary" class="grid grid-cols-4 gap-3 mb-3">
      <div class="bg-white border border-gray-300 p-3 rounded">
        <div class="text-xs font-medium text-gray-600 uppercase">Cash Balance</div>
        <div class="text-2xl font-bold font-mono tabular-nums mt-1" :class="getSummaryMetrics().cash >= 0 ? 'text-green-700' : 'text-red-700'">
          {{ formatCurrency(getSummaryMetrics().cash) }}
        </div>
      </div>
      <div class="bg-white border border-gray-300 p-3 rounded">
        <div class="text-xs font-medium text-gray-600 uppercase">AR Outstanding</div>
        <div class="text-2xl font-bold font-mono tabular-nums text-blue-700 mt-1">
          {{ formatCurrency(getSummaryMetrics().ar) }}
        </div>
      </div>
      <div class="bg-white border border-gray-300 p-3 rounded">
        <div class="text-xs font-medium text-gray-600 uppercase">AP Outstanding</div>
        <div class="text-2xl font-bold font-mono tabular-nums text-orange-700 mt-1">
          {{ formatCurrency(getSummaryMetrics().ap) }}
        </div>
      </div>
      <div class="bg-white border border-gray-300 p-3 rounded">
        <div class="text-xs font-medium text-gray-600 uppercase">Net Profit</div>
        <div class="text-2xl font-bold font-mono tabular-nums mt-1" :class="getSummaryMetrics().netProfit >= 0 ? 'text-green-700' : 'text-red-700'">
          {{ formatCurrency(getSummaryMetrics().netProfit) }}
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-gray-50 border border-gray-300 p-2 mb-3">
      <!-- Quick Date Presets -->
      <div class="flex gap-2 mb-2 pb-2 border-b border-gray-300">
        <button
          @click="setThisMonth"
          class="px-2 py-1 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded hover:bg-gray-100"
        >
          This Month
        </button>
        <button
          @click="setLastMonth"
          class="px-2 py-1 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded hover:bg-gray-100"
        >
          Last Month
        </button>
        <button
          @click="setThisQuarter"
          class="px-2 py-1 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded hover:bg-gray-100"
        >
          This Quarter
        </button>
        <button
          @click="setYearToDate"
          class="px-2 py-1 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded hover:bg-gray-100"
        >
          Year to Date
        </button>
        <button
          @click="setLastYear"
          class="px-2 py-1 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded hover:bg-gray-100"
        >
          Last Year
        </button>
      </div>
      
      <div class="flex items-center gap-3 text-xs">
        <div>
          <label class="block font-medium text-gray-700 mb-0.5">Start Date</label>
          <input
            type="date"
            v-model="startDate"
            class="block w-36 rounded border-gray-300 text-xs py-1 px-2"
          />
        </div>
        
        <div>
          <label class="block font-medium text-gray-700 mb-0.5">End Date</label>
          <input
            type="date"
            v-model="endDate"
            class="block w-36 rounded border-gray-300 text-xs py-1 px-2"
          />
        </div>

        <div class="flex-1">
          <label class="block font-medium text-gray-700 mb-0.5">Account</label>
          <select 
            v-model="selectedAccount"
            class="block w-full rounded border-gray-300 text-xs py-1"
          >
            <option value="">All Accounts</option>
            <option v-for="account in getUniqueAccounts()" :key="account" :value="account">
              {{ formatAccountName(account) }}
            </option>
          </select>
        </div>

        <div class="flex-1">
          <label class="block font-medium text-gray-700 mb-0.5">Subaccount</label>
          <input
            v-model="searchSubAccount"
            type="text"
            placeholder="Search..."
            class="block w-full rounded border-gray-300 text-xs py-1"
          />
        </div>

        <div class="flex items-end gap-2">
          <button
            @click="clearFilters"
            class="px-2 py-1 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded hover:bg-gray-50"
          >
            Clear
          </button>
          <button
            @click="fetchData"
            :disabled="isLoading"
            class="px-2 py-1 text-xs font-medium text-white bg-gray-800 rounded hover:bg-gray-900 disabled:bg-gray-400"
          >
            Refresh
          </button>
        </div>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="isLoading" class="text-center py-8">
      <i class="fas fa-spinner fa-spin text-2xl text-gray-600"></i>
      <p class="mt-2 text-sm text-gray-600">Loading...</p>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="border border-red-300 bg-red-50 p-2 text-red-800 text-sm">
      {{ error }}
    </div>

    <!-- Empty state -->
    <div v-else-if="journals.length === 0" class="border border-gray-300 p-8 text-center">
      <p class="text-sm text-gray-600">No journal entries for this period.</p>
    </div>

    <!-- Content -->
    <div v-else>
      <!-- Balance Summary -->
      <div v-if="balanceSummary" class="mb-4 border border-gray-900">
        <div class="bg-gray-900 text-white px-3 py-1">
          <h2 class="text-sm font-bold uppercase tracking-wide">Trial Balance</h2>
        </div>
        
        <!-- Summary totals -->
        <div class="grid grid-cols-4 gap-0 border-b border-gray-300">
          <div class="border-r border-gray-300 px-3 py-2">
            <div class="text-xs text-gray-600 uppercase">Total Debits</div>
            <div class="text-lg font-bold font-mono tabular-nums text-gray-900">{{ formatCurrency(balanceSummary.total_debits) }}</div>
          </div>
          <div class="border-r border-gray-300 px-3 py-2">
            <div class="text-xs text-gray-600 uppercase">Total Credits</div>
            <div class="text-lg font-bold font-mono tabular-nums text-gray-900">{{ formatCurrency(balanceSummary.total_credits) }}</div>
          </div>
          <div class="border-r border-gray-300 px-3 py-2">
            <div class="text-xs text-gray-600 uppercase">Net</div>
            <div class="text-lg font-bold font-mono tabular-nums" :class="balanceSummary.is_balanced ? 'text-gray-900' : 'text-red-600'">
              {{ formatCurrency(balanceSummary.net_balance) }}
            </div>
          </div>
          <div class="px-3 py-2 flex items-center justify-center">
            <span v-if="balanceSummary.is_balanced" class="text-xs font-medium text-green-700">✓ BALANCED</span>
            <span v-else class="text-xs font-medium text-red-700">⚠ UNBALANCED</span>
          </div>
        </div>

        <!-- Account categories -->
        <div v-if="balanceSummary.accounts && balanceSummary.accounts.length > 0">
          <div v-for="[category, accounts] in getAccountsByCategory()" :key="category" class="border-b border-gray-200">
            <div class="bg-gray-100 px-3 py-1 border-b border-gray-300">
              <h3 class="text-xs font-bold text-gray-700 uppercase">{{ category }}</h3>
            </div>
            <table class="w-full text-xs">
              <thead>
                <tr class="border-b border-gray-300">
                  <th class="px-3 py-1 text-left font-semibold text-gray-700 uppercase">Account</th>
                  <th class="px-3 py-1 text-right font-semibold text-gray-700 uppercase w-28">Debit</th>
                  <th class="px-3 py-1 text-right font-semibold text-gray-700 uppercase w-28">Credit</th>
                  <th class="px-3 py-1 text-right font-semibold text-gray-700 uppercase w-28">Balance</th>
                </tr>
              </thead>
              <tbody>
                <template v-for="account in accounts" :key="account.account">
                  <!-- Account row (clickable to show subaccounts) -->
                  <tr 
                    @click="toggleTrialBalanceAccount(account.account)" 
                    class="border-b border-gray-100 hover:bg-gray-50 cursor-pointer"
                  >
                    <td class="px-3 py-1 font-medium text-gray-900">
                      <i 
                        :class="isTrialBalanceAccountExpanded(account.account) ? 'fas fa-chevron-down' : 'fas fa-chevron-right'" 
                        class="text-gray-400 mr-2 text-xs"
                      ></i>
                      {{ formatAccountName(account.account) }}
                    </td>
                    <td class="px-3 py-1 text-right font-mono tabular-nums text-gray-700 w-28">{{ formatCurrency(account.total_debits) }}</td>
                    <td class="px-3 py-1 text-right font-mono tabular-nums text-gray-700 w-28">{{ formatCurrency(account.total_credits) }}</td>
                    <td class="px-3 py-1 text-right font-medium font-mono tabular-nums text-gray-900 w-28">{{ formatCurrency(account.net_balance) }}</td>
                  </tr>

                  <!-- Subaccount rows (shown when account is expanded) -->
                  <template v-if="isTrialBalanceAccountExpanded(account.account)">
                    <template v-for="sub in getSubaccountBalances(account.account)" :key="`${account.account}-${sub.subaccount}`">
                      <!-- Subaccount summary row -->
                      <tr 
                        @click.stop="toggleAccount(`${account.account}:${sub.subaccount}`)"
                        class="bg-gray-50 border-b border-gray-100 hover:bg-gray-100 cursor-pointer"
                      >
                        <td class="px-3 py-1 pl-8 text-gray-700">
                          <i 
                            :class="expandedAccounts.has(`${account.account}:${sub.subaccount}`) ? 'fas fa-chevron-down' : 'fas fa-chevron-right'" 
                            class="text-gray-400 mr-2 text-xs"
                          ></i>
                          {{ sub.subaccount }}
                        </td>
                        <td class="px-3 py-1 text-right font-mono tabular-nums text-gray-600 w-28">{{ formatCurrency(sub.total_debits) }}</td>
                        <td class="px-3 py-1 text-right font-mono tabular-nums text-gray-600 w-28">{{ formatCurrency(sub.total_credits) }}</td>
                        <td class="px-3 py-1 text-right font-mono tabular-nums text-gray-700 w-28">{{ formatCurrency(sub.net_balance) }}</td>
                      </tr>

                      <!-- Individual journal entries for this subaccount -->
                      <tr 
                        v-for="journal in getJournalsForSubaccount(account.account, sub.subaccount)" 
                        v-if="expandedAccounts.has(`${account.account}:${sub.subaccount}`)"
                        :key="journal.ID"
                        class="bg-white border-b border-gray-50"
                      >
                        <td class="px-3 py-1 pl-12 text-gray-600 text-xs">
                          <div>{{ formatDate(journal.CreatedAt) }}</div>
                          <div class="text-gray-500">{{ journal.memo || '-' }}</div>
                          <div v-if="journal.notes" class="text-gray-400 italic text-[10px] mt-0.5">
                            Note: {{ journal.notes }}
                          </div>
                        </td>
                        <td class="px-3 py-1 text-right font-mono tabular-nums text-gray-600 text-xs w-28">
                          {{ journal.debit > 0 ? formatCurrency(journal.debit) : '-' }}
                        </td>
                        <td class="px-3 py-1 text-right font-mono tabular-nums text-gray-600 text-xs w-28">
                          {{ journal.credit > 0 ? formatCurrency(journal.credit) : '-' }}
                        </td>
                        <td class="px-3 py-1 text-right w-28"></td>
                      </tr>
                    </template>
                  </template>
                </template>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Financial Statements Grid -->
      <div v-if="balanceSummary" class="grid grid-cols-3 gap-4 mb-4">
        <!-- Profit & Loss Statement -->
        <div class="border border-gray-900">
          <div class="bg-gray-900 text-white px-3 py-1">
            <h2 class="text-sm font-bold uppercase tracking-wide">Profit & Loss</h2>
          </div>
          
          <table class="w-full text-xs">
            <tbody>
              <tr class="border-b border-gray-200">
                <td class="px-3 py-1.5 font-bold text-gray-900 uppercase">Revenue</td>
                <td class="px-3 py-1.5 text-right font-bold font-mono tabular-nums text-gray-900">
                  {{ formatCurrency(calculateProfitLoss().revenue) }}
                  <span 
                    v-if="priorBalanceSummary" 
                    class="ml-2 text-xs font-normal"
                    :class="{
                      'text-green-600': calculatePercentChange(calculateProfitLoss().revenue, calculatePriorProfitLoss().revenue).direction === 'up',
                      'text-red-600': calculatePercentChange(calculateProfitLoss().revenue, calculatePriorProfitLoss().revenue).direction === 'down',
                      'text-gray-600': calculatePercentChange(calculateProfitLoss().revenue, calculatePriorProfitLoss().revenue).direction === 'flat'
                    }"
                  >
                    ({{ calculatePercentChange(calculateProfitLoss().revenue, calculatePriorProfitLoss().revenue).direction === 'up' ? '+' : calculatePercentChange(calculateProfitLoss().revenue, calculatePriorProfitLoss().revenue).direction === 'down' ? '-' : '' }}{{ calculatePercentChange(calculateProfitLoss().revenue, calculatePriorProfitLoss().revenue).value.toFixed(1) }}%)
                  </span>
                </td>
              </tr>
              <tr class="border-b border-gray-200">
                <td class="px-3 py-1.5 font-bold text-gray-900 uppercase">Expenses</td>
                <td class="px-3 py-1.5 text-right font-bold font-mono tabular-nums text-gray-900">
                  {{ formatCurrency(calculateProfitLoss().expenses) }}
                  <span 
                    v-if="priorBalanceSummary" 
                    class="ml-2 text-xs font-normal"
                    :class="{
                      'text-red-600': calculatePercentChange(calculateProfitLoss().expenses, calculatePriorProfitLoss().expenses).direction === 'up',
                      'text-green-600': calculatePercentChange(calculateProfitLoss().expenses, calculatePriorProfitLoss().expenses).direction === 'down',
                      'text-gray-600': calculatePercentChange(calculateProfitLoss().expenses, calculatePriorProfitLoss().expenses).direction === 'flat'
                    }"
                  >
                    ({{ calculatePercentChange(calculateProfitLoss().expenses, calculatePriorProfitLoss().expenses).direction === 'up' ? '+' : calculatePercentChange(calculateProfitLoss().expenses, calculatePriorProfitLoss().expenses).direction === 'down' ? '-' : '' }}{{ calculatePercentChange(calculateProfitLoss().expenses, calculatePriorProfitLoss().expenses).value.toFixed(1) }}%)
                  </span>
                </td>
              </tr>
              <tr class="bg-gray-50">
                <td class="px-3 py-1.5 font-bold text-gray-900 uppercase">Net Income</td>
                <td 
                  class="px-3 py-1.5 text-right text-lg font-bold font-mono tabular-nums"
                  :class="calculateProfitLoss().netIncome >= 0 ? 'text-green-700' : 'text-red-700'"
                >
                  {{ formatCurrency(calculateProfitLoss().netIncome) }}
                  <span 
                    v-if="priorBalanceSummary" 
                    class="ml-2 text-xs font-normal"
                    :class="{
                      'text-green-600': calculatePercentChange(calculateProfitLoss().netIncome, calculatePriorProfitLoss().netIncome).direction === 'up',
                      'text-red-600': calculatePercentChange(calculateProfitLoss().netIncome, calculatePriorProfitLoss().netIncome).direction === 'down',
                      'text-gray-600': calculatePercentChange(calculateProfitLoss().netIncome, calculatePriorProfitLoss().netIncome).direction === 'flat'
                    }"
                  >
                    ({{ calculatePercentChange(calculateProfitLoss().netIncome, calculatePriorProfitLoss().netIncome).direction === 'up' ? '+' : calculatePercentChange(calculateProfitLoss().netIncome, calculatePriorProfitLoss().netIncome).direction === 'down' ? '-' : '' }}{{ calculatePercentChange(calculateProfitLoss().netIncome, calculatePriorProfitLoss().netIncome).value.toFixed(1) }}%)
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Cash Flow Statement -->
        <div class="border border-gray-900">
          <div class="bg-gray-900 text-white px-3 py-1">
            <h2 class="text-sm font-bold uppercase tracking-wide">Cash Flow</h2>
          </div>
          
          <table class="w-full text-xs">
            <tbody>
              <tr class="border-b border-gray-100">
                <td class="px-3 py-1 text-gray-700">Cash Collections</td>
                <td class="px-3 py-1 text-right font-mono tabular-nums text-gray-900">
                  {{ formatCurrency(calculateCashFlow().collections) }}
                  <span 
                    v-if="priorBalanceSummary" 
                    class="ml-2 text-xs font-normal"
                    :class="{
                      'text-green-600': calculatePercentChange(calculateCashFlow().collections, calculatePriorCashFlow().collections).direction === 'up',
                      'text-red-600': calculatePercentChange(calculateCashFlow().collections, calculatePriorCashFlow().collections).direction === 'down',
                      'text-gray-600': calculatePercentChange(calculateCashFlow().collections, calculatePriorCashFlow().collections).direction === 'flat'
                    }"
                  >
                    ({{ calculatePercentChange(calculateCashFlow().collections, calculatePriorCashFlow().collections).direction === 'up' ? '+' : calculatePercentChange(calculateCashFlow().collections, calculatePriorCashFlow().collections).direction === 'down' ? '-' : '' }}{{ calculatePercentChange(calculateCashFlow().collections, calculatePriorCashFlow().collections).value.toFixed(1) }}%)
                  </span>
                </td>
              </tr>
              <tr class="border-b border-gray-200">
                <td class="px-3 py-1 text-gray-700">Cash Payments</td>
                <td class="px-3 py-1 text-right font-mono tabular-nums text-gray-900">
                  ({{ formatCurrency(calculateCashFlow().payments) }})
                  <span 
                    v-if="priorBalanceSummary" 
                    class="ml-2 text-xs font-normal"
                    :class="{
                      'text-red-600': calculatePercentChange(calculateCashFlow().payments, calculatePriorCashFlow().payments).direction === 'up',
                      'text-green-600': calculatePercentChange(calculateCashFlow().payments, calculatePriorCashFlow().payments).direction === 'down',
                      'text-gray-600': calculatePercentChange(calculateCashFlow().payments, calculatePriorCashFlow().payments).direction === 'flat'
                    }"
                  >
                    ({{ calculatePercentChange(calculateCashFlow().payments, calculatePriorCashFlow().payments).direction === 'up' ? '+' : calculatePercentChange(calculateCashFlow().payments, calculatePriorCashFlow().payments).direction === 'down' ? '-' : '' }}{{ calculatePercentChange(calculateCashFlow().payments, calculatePriorCashFlow().payments).value.toFixed(1) }}%)
                  </span>
                </td>
              </tr>
              <tr class="border-b border-gray-300 bg-gray-50">
                <td class="px-3 py-1.5 font-bold text-gray-900 uppercase">Operating Cash Flow</td>
                <td 
                  class="px-3 py-1.5 text-right font-bold font-mono tabular-nums"
                  :class="calculateCashFlow().operatingCash >= 0 ? 'text-gray-900' : 'text-red-700'"
                >
                  {{ formatCurrency(calculateCashFlow().operatingCash) }}
                  <span 
                    v-if="priorBalanceSummary" 
                    class="ml-2 text-xs font-normal"
                    :class="{
                      'text-green-600': calculatePercentChange(calculateCashFlow().operatingCash, calculatePriorCashFlow().operatingCash).direction === 'up',
                      'text-red-600': calculatePercentChange(calculateCashFlow().operatingCash, calculatePriorCashFlow().operatingCash).direction === 'down',
                      'text-gray-600': calculatePercentChange(calculateCashFlow().operatingCash, calculatePriorCashFlow().operatingCash).direction === 'flat'
                    }"
                  >
                    ({{ calculatePercentChange(calculateCashFlow().operatingCash, calculatePriorCashFlow().operatingCash).direction === 'up' ? '+' : calculatePercentChange(calculateCashFlow().operatingCash, calculatePriorCashFlow().operatingCash).direction === 'down' ? '-' : '' }}{{ calculatePercentChange(calculateCashFlow().operatingCash, calculatePriorCashFlow().operatingCash).value.toFixed(1) }}%)
                  </span>
                </td>
              </tr>
              <tr class="border-b border-gray-100">
                <td class="px-3 py-1 text-gray-700">Beginning Cash</td>
                <td class="px-3 py-1 text-right font-mono tabular-nums text-gray-900">{{ formatCurrency(calculateCashFlow().beginningCash) }}</td>
              </tr>
              <tr class="bg-gray-50">
                <td class="px-3 py-1.5 font-bold text-gray-900 uppercase">Ending Cash</td>
                <td 
                  class="px-3 py-1.5 text-right text-lg font-bold font-mono tabular-nums"
                  :class="calculateCashFlow().endingCash >= 0 ? 'text-green-700' : 'text-red-700'"
                >
                  {{ formatCurrency(calculateCashFlow().endingCash) }}
                  <span 
                    v-if="priorBalanceSummary" 
                    class="ml-2 text-xs font-normal"
                    :class="{
                      'text-green-600': calculatePercentChange(calculateCashFlow().endingCash, calculatePriorCashFlow().endingCash).direction === 'up',
                      'text-red-600': calculatePercentChange(calculateCashFlow().endingCash, calculatePriorCashFlow().endingCash).direction === 'down',
                      'text-gray-600': calculatePercentChange(calculateCashFlow().endingCash, calculatePriorCashFlow().endingCash).direction === 'flat'
                    }"
                  >
                    ({{ calculatePercentChange(calculateCashFlow().endingCash, calculatePriorCashFlow().endingCash).direction === 'up' ? '+' : calculatePercentChange(calculateCashFlow().endingCash, calculatePriorCashFlow().endingCash).direction === 'down' ? '-' : '' }}{{ calculatePercentChange(calculateCashFlow().endingCash, calculatePriorCashFlow().endingCash).value.toFixed(1) }}%)
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Balance Sheet -->
        <div class="border border-gray-900">
          <div class="bg-gray-900 text-white px-3 py-1">
            <h2 class="text-sm font-bold uppercase tracking-wide">Balance Sheet</h2>
          </div>
          
          <table class="w-full text-xs">
            <tbody>
              <!-- Assets Section -->
              <tr class="bg-gray-100 border-b border-gray-300">
                <td colspan="2" class="px-3 py-1.5 font-bold text-gray-900 uppercase">Assets</td>
              </tr>
              <tr 
                v-for="asset in calculateBalanceSheet().assets" 
                :key="asset.account"
                class="border-b border-gray-50"
              >
                <td class="px-3 py-1 pl-6 text-gray-700">{{ formatAccountName(asset.account) }}</td>
                <td class="px-3 py-1 text-right font-mono tabular-nums text-gray-900">
                  {{ formatCurrency(Math.abs(asset.balance)) }}
                </td>
              </tr>
              <tr class="border-b border-gray-300">
                <td class="px-3 py-1.5 pl-6 font-bold text-gray-900">Total Assets</td>
                <td class="px-3 py-1.5 text-right font-bold font-mono tabular-nums text-gray-900">
                  {{ formatCurrency(Math.abs(calculateBalanceSheet().totalAssets)) }}
                </td>
              </tr>

              <!-- Liabilities Section -->
              <tr class="bg-gray-100 border-b border-gray-300">
                <td colspan="2" class="px-3 py-1.5 font-bold text-gray-900 uppercase">Liabilities</td>
              </tr>
              <tr 
                v-for="liability in calculateBalanceSheet().liabilities" 
                :key="liability.account"
                class="border-b border-gray-50"
              >
                <td class="px-3 py-1 pl-6 text-gray-700">{{ formatAccountName(liability.account) }}</td>
                <td class="px-3 py-1 text-right font-mono tabular-nums text-gray-900">
                  {{ formatCurrency(Math.abs(liability.balance)) }}
                </td>
              </tr>
              <tr class="border-b border-gray-300">
                <td class="px-3 py-1.5 pl-6 font-bold text-gray-900">Total Liabilities</td>
                <td class="px-3 py-1.5 text-right font-bold font-mono tabular-nums text-gray-900">
                  {{ formatCurrency(Math.abs(calculateBalanceSheet().totalLiabilities)) }}
                </td>
              </tr>

              <!-- Equity Section -->
              <tr class="bg-gray-100 border-b border-gray-300">
                <td colspan="2" class="px-3 py-1.5 font-bold text-gray-900 uppercase">Equity</td>
              </tr>
              
              <!-- Partner Capital Contributions -->
              <tr v-if="calculateBalanceSheet().partnerCapital > 0" class="border-b border-gray-50">
                <td class="px-3 py-1 pl-6 text-gray-700">Partner Capital Contributions</td>
                <td class="px-3 py-1 text-right font-mono tabular-nums text-gray-900">
                  {{ formatCurrency(calculateBalanceSheet().partnerCapital) }}
                </td>
              </tr>
              
              <!-- Other Equity Accounts (if any) -->
              <tr 
                v-for="eq in calculateBalanceSheet().equity" 
                :key="eq.account" 
                class="border-b border-gray-50"
              >
                <td class="px-3 py-1 pl-6 text-gray-700">{{ formatAccountName(eq.account) }}</td>
                <td class="px-3 py-1 text-right font-mono tabular-nums text-gray-900">
                  {{ formatCurrency(Math.abs(eq.balance)) }}
                </td>
              </tr>
              
              <!-- Net Income -->
              <tr class="border-b border-gray-50">
                <td class="px-3 py-1 pl-6 text-gray-700">Net Income</td>
                <td 
                  class="px-3 py-1 text-right font-mono tabular-nums"
                  :class="calculateBalanceSheet().cumulativeNetIncome >= 0 ? 'text-green-700' : 'text-red-700'"
                >
                  {{ formatCurrency(Math.abs(calculateBalanceSheet().cumulativeNetIncome)) }}
                </td>
              </tr>
              
              <!-- Owner Distributions -->
              <tr class="border-b border-gray-50" v-if="calculateBalanceSheet().ownerDistributions && calculateBalanceSheet().ownerDistributions > 0">
                <td class="px-3 py-1 pl-6 text-gray-700">Owner Distributions</td>
                <td class="px-3 py-1 text-right font-mono tabular-nums text-red-700">
                  ({{ formatCurrency(calculateBalanceSheet().ownerDistributions || 0) }})
                </td>
              </tr>
              
              <tr class="border-b border-gray-300 bg-gray-50">
                <td class="px-3 py-1.5 pl-6 font-bold text-gray-900">Retained Equity</td>
                <td class="px-3 py-1.5 text-right font-bold font-mono tabular-nums text-gray-900">
                  {{ formatCurrency(Math.abs(calculateBalanceSheet().totalEquity)) }}
                  <span v-if="!calculateBalanceSheet().balances" class="ml-2 text-xs text-red-600">⚠️</span>
                  <span v-else class="ml-2 text-xs text-green-600">✓</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Journal Entries -->
      <div class="border border-gray-900">
        <div class="bg-gray-900 text-white px-3 py-1 flex justify-between items-center">
          <h2 class="text-sm font-bold uppercase tracking-wide">Journal Entries</h2>
          <span class="text-xs">{{ journals.length }} entries</span>
        </div>

        <div>
          <div v-for="[account, accountJournals] in getJournalsByAccount()" :key="account" class="border-b border-gray-300">
            <!-- Account header (clickable) -->
            <div
              @click="toggleAccount(account)"
              class="px-3 py-1.5 cursor-pointer hover:bg-gray-50 flex items-center justify-between bg-gray-100"
            >
              <div class="flex items-center gap-2 flex-1">
                <i :class="[expandedAccounts.has(account) ? 'fa-caret-down' : 'fa-caret-right', 'fas text-gray-600 text-xs']"></i>
                <div>
                  <span class="font-semibold text-sm text-gray-900">{{ formatAccountName(account) }}</span>
                  <span class="text-xs text-gray-600 ml-2">({{ accountJournals.length }})</span>
                </div>
              </div>
              <div class="flex items-center text-xs font-medium font-mono tabular-nums">
                <div class="w-36 text-right pr-2">
                  <span class="text-gray-700">DR: {{ formatCurrency(accountJournals.reduce((sum, j) => sum + j.debit, 0)) }}</span>
                </div>
                <div class="w-36 text-right pr-2">
                  <span class="text-gray-700">CR: {{ formatCurrency(accountJournals.reduce((sum, j) => sum + j.credit, 0)) }}</span>
                </div>
              </div>
            </div>

            <!-- Expanded journal entries -->
            <div v-if="expandedAccounts.has(account)">
              <table class="min-w-full text-xs table-fixed">
                <colgroup>
                  <col class="w-24"> <!-- Date -->
                  <col class="w-44"> <!-- Subaccount -->
                  <col> <!-- Memo (flexible) -->
                  <col class="w-36"> <!-- Debit -->
                  <col class="w-36"> <!-- Credit -->
                  <col class="w-24"> <!-- Ref -->
                </colgroup>
                <thead class="bg-gray-50 border-b border-gray-300">
                  <tr>
                    <th class="px-2 py-1 text-left font-semibold text-gray-700 uppercase">Date</th>
                    <th class="px-2 py-1 text-left font-semibold text-gray-700 uppercase">Subaccount</th>
                    <th class="px-2 py-1 text-left font-semibold text-gray-700 uppercase">Memo</th>
                    <th class="px-2 py-1 text-right font-semibold text-gray-700 uppercase">Debit</th>
                    <th class="px-2 py-1 text-right font-semibold text-gray-700 uppercase">Credit</th>
                    <th class="px-2 py-1 text-center font-semibold text-gray-700 uppercase">Ref</th>
                    <th class="px-2 py-1 text-center font-semibold text-gray-700 uppercase">Actions</th>
                  </tr>
                </thead>
                <tbody class="bg-white">
                  <tr v-for="journal in accountJournals" :key="journal.ID" class="border-b border-gray-100 hover:bg-gray-50">
                    <td class="px-2 py-1 text-gray-900 whitespace-nowrap">{{ formatDate(journal.CreatedAt) }}</td>
                    <td class="px-2 py-1 text-gray-700 truncate" :title="journal.sub_account">{{ journal.sub_account || '-' }}</td>
                    <td class="px-2 py-1 text-gray-600">
                      <div>{{ journal.memo }}</div>
                      <div v-if="journal.notes" class="text-gray-400 italic text-[10px] mt-0.5">
                        Note: {{ journal.notes }}
                      </div>
                    </td>
                    <td class="px-2 py-1 text-right font-mono text-gray-900 tabular-nums">
                      {{ journal.debit > 0 ? formatCurrency(journal.debit) : '' }}
                    </td>
                    <td class="px-2 py-1 text-right font-mono text-gray-900 tabular-nums">
                      {{ journal.credit > 0 ? formatCurrency(journal.credit) : '' }}
                    </td>
                    <td class="px-2 py-1 text-center">
                      <a
                        v-if="journal.invoice_id"
                        :href="`/admin/accounts-receivable`"
                        class="text-blue-700 hover:underline text-xs"
                      >
                        INV-{{ journal.invoice_id }}
                      </a>
                      <a
                        v-else-if="journal.bill_id"
                        :href="`/admin/accounts-payable`"
                        class="text-blue-700 hover:underline text-xs"
                      >
                        BILL-{{ journal.bill_id }}
                      </a>
                      <span v-else class="text-gray-400">-</span>
                    </td>
                    <td class="px-2 py-1 text-center">
                      <button
                        v-if="!(journal as any).isAggregated"
                        @click="openAdjustmentModal(journal)"
                        class="text-orange-600 hover:text-orange-900"
                        title="Adjust/Reverse Entry"
                      >
                        <i class="fas fa-exclamation-triangle text-xs"></i>
                      </button>
                      <span v-else class="text-gray-400 text-xs" title="Cannot adjust aggregated entries">-</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- Empty state -->
        <div v-if="getJournalsByAccount().length === 0" class="px-3 py-8 text-center text-sm text-gray-500">
          No entries match the selected filters.
        </div>
      </div>
    </div>
  </div>

  <!-- Manual Journal Entry Modal -->
  <div
    v-if="showManualEntryModal"
    class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
    @click.self="showManualEntryModal = false"
  >
    <div class="bg-white rounded-lg shadow-xl max-w-4xl w-full mx-4 max-h-[90vh] overflow-y-auto">
      <!-- Header -->
      <div class="px-6 py-4 border-b border-gray-200 flex justify-between items-center">
        <h2 class="text-lg font-bold text-gray-900">Book Manual Journal Entry</h2>
        <button
          @click="showManualEntryModal = false"
          class="text-gray-400 hover:text-gray-600"
        >
          <i class="fas fa-times text-xl"></i>
        </button>
      </div>

      <!-- Content -->
      <div class="px-4 py-3">
        <!-- Date -->
        <div class="mb-3">
          <label class="block text-xs font-medium text-gray-700 mb-1">Entry Date</label>
          <input
            type="date"
            v-model="manualEntryDate"
            class="block w-40 rounded border-gray-300 text-xs py-1 px-2"
          />
        </div>

        <!-- Error Message -->
        <div v-if="manualEntryError" class="mb-3 bg-red-50 border border-red-200 text-red-800 px-3 py-2 rounded text-xs">
          <i class="fas fa-exclamation-circle mr-1"></i>{{ manualEntryError }}
        </div>

        <!-- Entry Lines -->
        <div class="mb-3">
          <table class="min-w-full text-xs">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-2 py-1.5 text-left text-xs font-medium text-gray-700 uppercase tracking-wide">Account</th>
                <th class="px-2 py-1.5 text-left text-xs font-medium text-gray-700 uppercase tracking-wide">Subaccount</th>
                <th class="px-2 py-1.5 text-right text-xs font-medium text-gray-700 uppercase tracking-wide">Debit ($)</th>
                <th class="px-2 py-1.5 text-right text-xs font-medium text-gray-700 uppercase tracking-wide">Credit ($)</th>
                <th class="px-2 py-1.5 text-left text-xs font-medium text-gray-700 uppercase tracking-wide">Memo</th>
                <th class="px-2 py-1.5 text-center text-xs font-medium text-gray-700 uppercase tracking-wide w-8"></th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 bg-white">
              <tr v-for="(line, index) in manualEntryLines" :key="index">
                <td class="px-2 py-1.5 w-48">
                  <select
                    v-model="line.account"
                    class="block w-full rounded border-gray-300 text-[11px] py-0.5 px-1"
                    style="font-size: 11px;"
                    required
                  >
                    <option value="">Select Account...</option>
                    <option v-for="account in availableAccounts" :key="account.code" :value="account.code">
                      {{ account.code }} - {{ account.name }}
                    </option>
                  </select>
                </td>
                <td class="px-2 py-1.5 w-44">
                  <!-- Show select dropdown if suggestions exist, otherwise text input -->
                  <select
                    v-if="getSubaccountSuggestions(line.account).length > 0"
                    v-model="line.subaccount"
                    class="block w-full rounded border-gray-300 text-[11px] py-0.5 px-1"
                    style="font-size: 11px;"
                    :disabled="isLoadingAccountOptions"
                  >
                    <option value="">{{ isLoadingAccountOptions ? 'Loading...' : 'Select...' }}</option>
                    <option 
                      v-for="suggestion in getSubaccountSuggestions(line.account)" 
                      :key="suggestion.value" 
                      :value="suggestion.value"
                    >
                      {{ suggestion.label }}
                    </option>
                  </select>
                  <input
                    v-else
                    type="text"
                    v-model="line.subaccount"
                    :placeholder="isLoadingAccountOptions ? 'Loading...' : 'Optional'"
                    class="block w-full rounded border-gray-300 text-[11px] py-0.5 px-1.5"
                    style="font-size: 11px;"
                    :disabled="isLoadingAccountOptions"
                  />
                </td>
                <td class="px-2 py-1.5 w-24">
                  <input
                    type="number"
                    v-model.number="line.debit"
                    step="0.01"
                    min="0"
                    class="block w-full rounded border-gray-300 text-[11px] py-0.5 px-1.5 text-right font-mono"
                    style="font-size: 11px;"
                  />
                </td>
                <td class="px-2 py-1.5 w-24">
                  <input
                    type="number"
                    v-model.number="line.credit"
                    step="0.01"
                    min="0"
                    class="block w-full rounded border-gray-300 text-[11px] py-0.5 px-1.5 text-right font-mono"
                    style="font-size: 11px;"
                  />
                </td>
                <td class="px-2 py-1.5">
                  <input
                    type="text"
                    v-model="line.memo"
                    placeholder="Description..."
                    class="block w-full rounded border-gray-300 text-[11px] py-0.5 px-1.5"
                    style="font-size: 11px;"
                  />
                </td>
                <td class="px-2 py-1.5 text-center w-6">
                  <button
                    v-if="manualEntryLines.length > 2"
                    @click="removeManualEntryLine(index)"
                    class="text-red-600 hover:text-red-800 text-xs"
                  >
                    <i class="fas fa-times"></i>
                  </button>
                </td>
              </tr>
            </tbody>
            <tfoot class="bg-gray-50 font-bold border-t-2 border-gray-300">
              <tr>
                <td colspan="2" class="px-2 py-1.5 text-right text-xs">Totals:</td>
                <td class="px-2 py-1.5 text-right font-mono text-xs" :class="isBalanced() ? 'text-green-700' : 'text-red-700'">
                  ${{ calculateTotalDebits().toFixed(2) }}
                </td>
                <td class="px-2 py-1.5 text-right font-mono text-xs" :class="isBalanced() ? 'text-green-700' : 'text-red-700'">
                  ${{ calculateTotalCredits().toFixed(2) }}
                </td>
                <td colspan="2" class="px-2 py-1.5 text-xs" :class="isBalanced() ? 'text-green-700' : 'text-red-700'">
                  <i v-if="isBalanced() && calculateTotalDebits() > 0" class="fas fa-check-circle mr-1"></i>
                  <i v-if="!isBalanced() && calculateTotalDebits() > 0" class="fas fa-exclamation-circle mr-1"></i>
                  {{ isBalanced() ? 'Balanced' : 'Unbalanced' }}
                </td>
              </tr>
            </tfoot>
          </table>
        </div>

        <!-- Add Line Button -->
        <button
          @click="addManualEntryLine"
          class="px-2 py-1 text-xs font-medium text-blue-600 border border-blue-600 rounded hover:bg-blue-50"
        >
          <i class="fas fa-plus mr-1"></i> Add Line
        </button>
      </div>

      <!-- Footer -->
      <div class="px-4 py-2.5 border-t border-gray-200 flex justify-end gap-2">
        <button
          @click="showManualEntryModal = false"
          class="px-3 py-1.5 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded hover:bg-gray-50"
          :disabled="isSubmittingManualEntry"
        >
          Cancel
        </button>
        <button
          @click="submitManualEntry"
          class="px-3 py-1.5 text-xs font-medium text-white bg-green-600 hover:bg-green-700 rounded disabled:bg-gray-400"
          :disabled="isSubmittingManualEntry || !isBalanced() || calculateTotalDebits() === 0"
        >
          <i v-if="isSubmittingManualEntry" class="fas fa-spinner fa-spin mr-1"></i>
          {{ isSubmittingManualEntry ? 'Booking...' : 'Book Entry' }}
        </button>
      </div>
    </div>
  </div>

  <!-- Adjustment/Reversal Modal -->
  <div
    v-if="adjustmentModalOpen"
    @click="closeAdjustmentModal"
    class="fixed inset-0 z-50 flex items-center justify-center bg-gray-500/75"
  >
    <div
      @click.stop
      class="bg-white rounded-lg shadow-xl w-full max-w-2xl mx-4 max-h-[90vh] overflow-y-auto"
    >
      <div class="bg-orange-600 px-4 py-3 rounded-t-lg">
        <h3 class="text-sm font-semibold text-white">Adjust/Reverse Journal Entry</h3>
      </div>
      <div class="p-4 space-y-4">
        <!-- Warning Banner -->
        <div class="flex items-start gap-3 bg-yellow-50 border border-yellow-200 rounded p-3">
          <i class="fas fa-exclamation-triangle text-yellow-600 text-xl"></i>
          <div class="text-xs text-gray-700">
            <p class="font-medium mb-1">This will create reversing journal entries</p>
            <p>The original entry will remain in the ledger for audit purposes. A new reversing entry will be created to cancel it out, and optionally a corrected entry can be created with the right values.</p>
          </div>
        </div>

        <!-- Original Entry (Read-only) -->
        <div v-if="adjustingJournal" class="border border-gray-300 rounded p-3 bg-gray-50">
          <h4 class="text-xs font-semibold text-gray-900 mb-2">Original Entry</h4>
          <div class="grid grid-cols-2 gap-2 text-xs">
            <div><span class="font-medium">Date:</span> {{ formatDate(adjustingJournal.CreatedAt) }}</div>
            <div><span class="font-medium">Account:</span> {{ formatAccountName(adjustingJournal.account) }}</div>
            <div><span class="font-medium">Subaccount:</span> {{ adjustingJournal.sub_account || '-' }}</div>
            <div><span class="font-medium">Memo:</span> {{ adjustingJournal.memo }}</div>
            <div><span class="font-medium">Debit:</span> {{ formatCurrency(adjustingJournal.debit) }}</div>
            <div><span class="font-medium">Credit:</span> {{ formatCurrency(adjustingJournal.credit) }}</div>
          </div>
        </div>

        <!-- Reason for Reversal -->
        <div>
          <label class="block text-xs font-medium text-gray-700 mb-1">Reason for Reversal *</label>
          <textarea
            v-model="adjustmentReason"
            rows="3"
            class="w-full px-2 py-1.5 text-xs border rounded focus:ring-1 focus:ring-orange-500"
            placeholder="Explain why this entry needs to be reversed..."
          ></textarea>
        </div>

        <!-- Create Corrected Entry -->
        <div class="text-xs">
          <label class="flex items-center gap-3 font-medium text-gray-700 cursor-pointer text-xs">
            <input
              type="checkbox"
              v-model="createCorrectedEntry"
              class="rounded border-gray-300 text-sage focus:ring-sage h-4 w-4"
            />
            <span class="text-xs">Create corrected entry</span>
          </label>
        </div>

        <!-- Corrected Entry Form -->
        <div v-if="createCorrectedEntry" class="border border-blue-200 rounded p-2 bg-blue-50 space-y-1.5 text-xs">
          <h4 class="font-semibold text-gray-900 text-xs mb-1">Corrected Entry Details</h4>
          
          <div class="grid grid-cols-2 gap-1.5">
            <div>
              <label class="block font-medium text-gray-700 mb-0.5 text-xs">Account</label>
              <select v-model="correctedForm.account" class="w-full px-1.5 py-0.5 border rounded text-xs" style="font-size: 11px;">
                <option value="">Select account...</option>
                <option v-for="account in availableAccounts" :key="account.code" :value="account.code">
                  {{ account.code }} - {{ account.name }}
                </option>
              </select>
            </div>
            <div>
              <label class="block font-medium text-gray-700 mb-0.5 text-xs">Subaccount</label>
              <select
                v-if="getSubaccountSuggestions(correctedForm.account).length > 0"
                v-model="correctedForm.sub_account"
                class="w-full px-1.5 py-0.5 border rounded text-xs"
                style="font-size: 11px;"
              >
                <option value="">None</option>
                <option v-for="sub in getSubaccountSuggestions(correctedForm.account)" :key="sub.value" :value="sub.value">
                  {{ sub.label }}
                </option>
              </select>
              <input
                v-else
                type="text"
                v-model="correctedForm.sub_account"
                class="w-full px-1.5 py-0.5 border rounded text-xs"
                style="font-size: 11px;"
                placeholder="Optional"
              />
            </div>
          </div>

          <div>
            <label class="block font-medium text-gray-700 mb-0.5 text-xs">Memo</label>
            <input
              type="text"
              v-model="correctedForm.memo"
              class="w-full px-1.5 py-0.5 border rounded text-xs"
              style="font-size: 11px;"
              placeholder="Description of corrected entry..."
            />
          </div>

          <div class="grid grid-cols-2 gap-1.5">
            <div>
              <label class="block font-medium text-gray-700 mb-0.5 text-xs">Debit ($)</label>
              <input
                type="number"
                v-model.number="correctedForm.debit"
                step="0.01"
                class="w-full px-1.5 py-0.5 border rounded text-right font-mono text-xs"
                style="font-size: 11px;"
              />
            </div>
            <div>
              <label class="block font-medium text-gray-700 mb-0.5 text-xs">Credit ($)</label>
              <input
                type="number"
                v-model.number="correctedForm.credit"
                step="0.01"
                class="w-full px-1.5 py-0.5 border rounded text-right font-mono text-xs"
                style="font-size: 11px;"
              />
            </div>
          </div>
        </div>

        <!-- Actions -->
        <div class="flex justify-end gap-2 pt-2">
          <button
            @click="closeAdjustmentModal"
            class="px-3 py-1.5 text-xs text-gray-700 border border-gray-300 rounded hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            @click="submitAdjustment"
            :disabled="!adjustmentReason.trim()"
            class="px-3 py-1.5 text-xs bg-orange-600 text-white rounded hover:bg-orange-700 disabled:bg-gray-400"
          >
            <i class="fas fa-exclamation-triangle mr-1"></i>
            {{ createCorrectedEntry ? 'Reverse & Correct' : 'Reverse Only' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
