# Staffing Assignment Backfill Tool

This is a one-time script to backfill `staffing_assignment_id` values for entries that don't have them populated.

## Purpose

After adding staffing assignment tracking to entries, this script identifies the correct staffing assignment for existing entries based on:
- The entry's project ID
- Whether the entry's start time falls within the staffing assignment's date range (inclusive)

## Logic

For each entry without a `staffing_assignment_id`:
1. Find all `StaffingAssignment` records where:
   - `StaffingAssignment.project_id == Entry.project_id`
   - `Entry.start >= StaffingAssignment.start_date`
   - `Entry.start <= StaffingAssignment.end_date`

2. Outcomes:
   - **Exactly 1 match**: Assigns the staffing assignment ID to the entry
   - **0 matches**: Logs as "NO MATCH" (entry remains without assignment)
   - **Multiple matches**: Selects the staffing assignment with the earliest start date and assigns it (logged as "AMBIGUOUS (using earliest)")

## Usage

### Dry Run (Recommended First)

Run without making any changes to see what would happen:

```bash
# From the cronos directory
cd /Users/naterobinson/go/src/github.com/snowpackdata/cronos

# SQLite (default)
go run cmd/backfill-staffing/main.go

# Or specify the database explicitly
go run cmd/backfill-staffing/main.go -db cronos.db -type sqlite
```

### Execute Mode

Once you've reviewed the dry run output and are satisfied, run with `-execute`:

```bash
go run cmd/backfill-staffing/main.go -execute
```

### PostgreSQL

For a PostgreSQL database:

```bash
go run cmd/backfill-staffing/main.go \
  -type postgres \
  -user YOUR_USER \
  -password YOUR_PASSWORD \
  -dbname YOUR_DATABASE \
  -execute
```

### Cloud Database

For a cloud database connection:

```bash
go run cmd/backfill-staffing/main.go \
  -type cloud \
  -uri "YOUR_CONNECTION_STRING" \
  -execute
```

## Output

The script will log:
- **MATCH**: Successfully matched entries with exactly one staffing assignment
- **NO MATCH**: Entries that don't have any eligible staffing assignment
- **AMBIGUOUS (using earliest)**: Entries that matched multiple staffing assignments (automatically resolved by selecting the one with the earliest start date)
- **ERROR**: Any errors encountered during processing

At the end, it prints a summary with counts for each category.

## After Running

Once the backfill is complete and verified:
1. Review any NO MATCH cases if needed (entries that legitimately don't have assignments will remain without one)
2. Delete this script directory: `rm -rf cmd/backfill-staffing/`

