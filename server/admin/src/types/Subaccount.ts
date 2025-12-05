export interface Subaccount {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt?: string;
  code: string;
  name: string;
  account_code: string;
  type: 'VENDOR' | 'CLIENT' | 'EMPLOYEE' | 'CUSTOM' | 'PROJECT';
  is_active: boolean;
}

export interface SubaccountCreate {
  code: string;
  name: string;
  account_code: string;
  type: 'VENDOR' | 'CLIENT' | 'EMPLOYEE' | 'CUSTOM' | 'PROJECT';
  is_active?: boolean;
}

export interface SubaccountUpdate {
  name?: string;
  account_code?: string;
  type?: 'VENDOR' | 'CLIENT' | 'EMPLOYEE' | 'CUSTOM' | 'PROJECT';
  is_active?: boolean;
}

export const SUBACCOUNT_TYPES = [
  { value: 'VENDOR', label: 'Vendor' },
  { value: 'CLIENT', label: 'Client' },
  { value: 'EMPLOYEE', label: 'Employee' },
  { value: 'CUSTOM', label: 'Custom' },
] as const;

export function getSubaccountTypeLabel(type: string): string {
  const found = SUBACCOUNT_TYPES.find(t => t.value === type);
  return found ? found.label : type;
}

export function getSubaccountTypeColor(type: string): string {
  switch (type) {
    case 'VENDOR':
      return 'bg-purple-100 text-purple-800';
    case 'CLIENT':
      return 'bg-blue-100 text-blue-800';
    case 'EMPLOYEE':
      return 'bg-green-100 text-green-800';
    case 'CUSTOM':
      return 'bg-gray-100 text-gray-800';
    default:
      return 'bg-gray-100 text-gray-800';
  }
}

