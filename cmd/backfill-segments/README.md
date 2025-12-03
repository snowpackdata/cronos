# Backfill Commitment Schedules

This script converts existing simple staffing assignments to the new segment-based format.

## What it does

1. Adds the `commitment_schedule` field to the `staffing_assignments` table (if not exists)
2. Converts each existing assignment with a single commitment period into a JSON segment format
3. Preserves all existing data - fully backward compatible

## Running the script

### For local SQLite database:
```bash
cd /Users/naterobinson/go/src/github.com/snowpackdata/cronos/cmd/backfill-segments
DB_PATH=../../cronos.db go run main.go
```

### For Production Database (via Cloud SQL Proxy) - RECOMMENDED:

1. Start the cloud-sql-proxy in one terminal:
```bash
cloud-sql-proxy project:region:instance --port 5432
```

2. In another terminal, run DRY RUN first to preview changes:
```bash
cd /Users/naterobinson/go/src/github.com/snowpackdata/cronos/cmd/backfill-segments
DRY_RUN=true \
CLOUD_SQL_USERNAME=your_user \
CLOUD_SQL_PASSWORD=your_password \
CLOUD_SQL_CONNECTION_NAME=project:region:instance \
CLOUD_SQL_DATABASE_NAME=your_db \
go run main.go
```

3. If everything looks good, run WITHOUT dry run to apply:
```bash
CLOUD_SQL_USERNAME=your_user \
CLOUD_SQL_PASSWORD=your_password \
CLOUD_SQL_CONNECTION_NAME=project:region:instance \
CLOUD_SQL_DATABASE_NAME=your_db \
go run main.go
```

### Options:

- `DRY_RUN=true` - Preview changes without modifying database
- `DB_PORT=5432` - Override default PostgreSQL port (if needed)
- `DB_PATH=/path/to/db.sqlite` - Use local SQLite instead of Cloud SQL

## What gets changed

Before:
```
StaffingAssignment {
  ID: 1
  Commitment: 40
  StartDate: "2024-01-01"
  EndDate: "2024-03-31"
  CommitmentSchedule: null
}
```

After:
```
StaffingAssignment {
  ID: 1
  Commitment: 40  // Unchanged - kept as fallback
  StartDate: "2024-01-01"  // Unchanged
  EndDate: "2024-03-31"  // Unchanged
  CommitmentSchedule: '{"segments":[{"start_date":"2024-01-01","end_date":"2024-03-31","commitment":40}]}'
}
```

## Safety

- Non-destructive: Only adds data, never removes or modifies existing fields
- Idempotent: Can be run multiple times safely (only processes null schedules)
- Backward compatible: Old code continues to work with existing fields

