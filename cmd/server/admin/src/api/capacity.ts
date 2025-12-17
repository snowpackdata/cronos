import { fetchAll } from './apiUtils';

export interface WeeklyUtilization {
  week_start: string;
  actual_hours: number;
  commitment: number;
  utilization: number;
}

export interface CommitmentSegment {
  start_date: string;
  end_date: string;
  commitment: number;
}

export interface Entry {
  ID: number;
  start: string;
  notes: string;
  duration_minutes: number;
}

export interface EntryDetail {
  id: number;
  start: string;
  notes: string;
  duration_minutes: number;
}

export interface CapacityAssignment {
  ID: number;
  employee_id: number;
  employee: {
    ID: number;
    first_name: string;
    last_name: string;
    headshot_asset?: {
      url?: string;
    };
  };
  project_id: number;
  project: {
    ID: number;
    name: string;
    account?: {
      ID: number;
      name: string;
    };
  };
  commitment: number; // Legacy/fallback
  start_date: string;
  end_date: string;
  commitment_schedule?: string; // JSON string of segments
  segments?: CommitmentSegment[]; // Parsed segments (if available)
  entries?: Entry[]; // Time entries for this assignment
  weekly_utilization: Record<string, WeeklyUtilization>;
}

// Fetches all staffing assignments for capacity management
export const fetchCapacityData = async (): Promise<CapacityAssignment[]> => {
  try {
    return await fetchAll<CapacityAssignment>('capacity');
  } catch (error) {
    console.error('Error fetching capacity data:', error);
    throw error;
  }
};

// Fetches detailed time entries for a specific assignment and week
export const fetchCapacityEntries = async (assignmentId: number, weekStart: string): Promise<EntryDetail[]> => {
  try {
    const token = localStorage.getItem('snowpack_token');
    const response = await fetch(`/api/capacity/detail?assignment_id=${assignmentId}&week_start=${weekStart}`, {
      headers: {
        'x-access-token': token || '',
      },
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return await response.json();
  } catch (error) {
    console.error('Error fetching capacity entries:', error);
    throw error;
  }
};
