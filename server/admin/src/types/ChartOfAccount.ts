export interface ChartOfAccount {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt?: string;
  account_code: string;
  account_name: string;
  account_type: 'ASSET' | 'LIABILITY' | 'EQUITY' | 'REVENUE' | 'EXPENSE';
  parent_id?: number;
  is_active: boolean;
  description: string;
  is_system_defined: boolean;
}

export interface ChartOfAccountCreate {
  account_code: string;
  account_name: string;
  account_type: 'ASSET' | 'LIABILITY' | 'EQUITY' | 'REVENUE' | 'EXPENSE';
  description?: string;
  parent_id?: number;
  is_active?: boolean;
}

export interface ChartOfAccountUpdate {
  account_name?: string;
  account_type?: 'ASSET' | 'LIABILITY' | 'EQUITY' | 'REVENUE' | 'EXPENSE';
  description?: string;
  parent_id?: number;
  is_active?: boolean;
}

export const ACCOUNT_TYPES = [
  { value: 'ASSET', label: 'Asset' },
  { value: 'LIABILITY', label: 'Liability' },
  { value: 'EQUITY', label: 'Equity' },
  { value: 'REVENUE', label: 'Revenue' },
  { value: 'EXPENSE', label: 'Expense' },
] as const;

export function getAccountTypeLabel(type: string): string {
  const found = ACCOUNT_TYPES.find(t => t.value === type);
  return found ? found.label : type;
}

export function getAccountTypeColor(type: string): string {
  switch (type) {
    case 'ASSET':
      return 'bg-green-100 text-green-800';
    case 'LIABILITY':
      return 'bg-red-100 text-red-800';
    case 'EQUITY':
      return 'bg-purple-100 text-purple-800';
    case 'REVENUE':
      return 'bg-blue-100 text-blue-800';
    case 'EXPENSE':
      return 'bg-orange-100 text-orange-800';
    default:
      return 'bg-gray-100 text-gray-800';
  }
}

