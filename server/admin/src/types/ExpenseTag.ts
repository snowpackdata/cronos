export interface ExpenseTag {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string | null;
  name: string;
  description: string;
  active: boolean;
  budget: number | null; // Budget in cents
  total_spent?: number; // Populated by backend
  remaining_budget?: number | null; // Populated by backend
  budget_percentage?: number | null; // Populated by backend
}

