// Employment Status Constants (matching Go cronos/models.go)
export const EmploymentStatus = {
  ACTIVE: 'EMPLOYMENT_STATUS_ACTIVE',
  INACTIVE: 'EMPLOYMENT_STATUS_INACTIVE',
  TERMINATED: 'EMPLOYMENT_STATUS_TERMINATED'
} as const;

export type EmploymentStatusType = typeof EmploymentStatus[keyof typeof EmploymentStatus];

// Compensation Type Constants (matching Go cronos/models.go)
export const CompensationType = {
  FULLY_VARIABLE: 'COMPENSATION_TYPE_FULLY_VARIABLE',
  SALARIED: 'COMPENSATION_TYPE_SALARIED',
  BASE_PLUS_VARIABLE: 'COMPENSATION_TYPE_BASE_PLUS_VARIABLE'
} as const;

export type CompensationTypeType = typeof CompensationType[keyof typeof CompensationType];

// Human readable labels for display
export const EmploymentStatusLabels = {
  [EmploymentStatus.ACTIVE]: 'Active',
  [EmploymentStatus.INACTIVE]: 'Inactive',
  [EmploymentStatus.TERMINATED]: 'Terminated'
} as const;

export const CompensationTypeLabels = {
  [CompensationType.FULLY_VARIABLE]: 'Fully Variable',
  [CompensationType.SALARIED]: 'Salaried',
  [CompensationType.BASE_PLUS_VARIABLE]: 'Base + Variable'
} as const;

// Human readable descriptions
export const CompensationTypeDescriptions = {
  [CompensationType.FULLY_VARIABLE]: '100% based on billable hours',
  [CompensationType.SALARIED]: 'Fixed annual salary',
  [CompensationType.BASE_PLUS_VARIABLE]: 'Base salary + client billables'
} as const;
