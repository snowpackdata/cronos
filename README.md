# Cronos
Cronos is a comprehensive timekeeping and accounting system providing visibility into accounts, users, billing codes, worked time, and full AP/AR functionality.

This application includes both the core business logic and a complete web interface with admin and client portal applications.

## Features

### Impersonation
Cronos supports an "Impersonate" feature for timesheet entries, allowing one staff member to create timesheet entries that appear to be done by another staff member to clients. This is useful in scenarios where work needs to be attributed to a specific team member for client-facing purposes, while ensuring the actual creator gets paid internally.

How it works:
- When creating a timesheet entry, a user can optionally choose to "impersonate" another staff member
- On client-facing invoices, the impersonated staff member's name will appear instead of the actual creator
- For internal billing and payment purposes, the entry is still attributed to the original creator
- Both the original creator and impersonated user can view the entry in their respective dashboards

## Installation

### Prerequisites
- Go 1.21 or higher
- Node.js and npm (for admin and portal UIs)
- Google Cloud SDK (for deployment)
- Cloud SQL Proxy (for development with Cloud SQL)

### Local Development Setup

1. Clone this repository
2. Navigate to the server directory:
```bash
cd server
```

3. Set up environment variables. For local development with SQLite:
```bash
export ENVIRONMENT=local
```

For development with Cloud SQL:
```bash
export ENVIRONMENT=development
export CLOUD_SQL_USERNAME=<username>
export CLOUD_SQL_PASSWORD=<password>
export CLOUD_SQL_CONNECTION_NAME=<project:region:instance>
export CLOUD_SQL_DATABASE_NAME=<database>
export SENDGRID_API_KEY=<key>
export GCS_BUCKET=<bucket>
```

4. Start the development server with hot reloading:
```bash
make run-dev-hot
```

This will:
- Start the Vue.js dev servers for admin and portal interfaces on ports 5173 and 5174
- Start the Go backend with hot reloading on port 8080
- Initialize SQLite database with seed data

5. Access the application:
- Admin Interface: http://localhost:5173/admin/
- Portal Interface: http://localhost:5174/portal/
- Backend API: http://localhost:8080

### Running with Compiled UIs

To run with production-like compiled assets:
```bash
make run-dev
```

### Development Credentials

When running in local mode, the following test accounts are created:
- Dev User: dev@example.com / devpassword
- Client User: client@example.com / password

## Deployment

### Deploy to Google App Engine

Ensure you have the Google Cloud SDK installed and configured:

```bash
gcloud auth login
gcloud config set project snowpack-368423
```

Deploy the application:
```bash
cd server
make deploy
```

This will:
1. Build the admin and portal Vue.js applications
2. Deploy to Google App Engine

### Environment Variables for Production

Set these in the GAE environment or via Secret Manager:
- `ENVIRONMENT=production`
- `CLOUD_SQL_USERNAME`
- `CLOUD_SQL_PASSWORD`
- `CLOUD_SQL_CONNECTION_NAME`
- `CLOUD_SQL_DATABASE_NAME`
- `SENDGRID_API_KEY`
- `GCS_BUCKET`
- `JWT_SECRET`
- `GIT_HASH` (optional, for version tracking)

## Development

### Test Data
The application includes a seed database function that populates the database with comprehensive test data for development and testing purposes:

- **Users and Employees**: Multiple user accounts with various roles (Admin, Staff) and corresponding employee profiles
- **Accounts**: Both internal (Snowpack) and client accounts with complete billing settings  
- **Projects**: Various project types (New, Existing) with appropriate metadata, including:
  - Project type designation (New/Existing client) for commission calculation
  - Budget hours and dollars
  - AE (Account Executive) and SDR (Sales Development Representative) assignments
  - Billing frequency settings
- **Rates**: External and internal rate structures
- **Billing Codes**: Complete billing codes with both external and internal rate associations
- **Entries**: Sample timesheet entries including examples of:
  - Regular entries
  - Impersonated entries
  - Internal entries
  - Internal impersonated entries

The seed data is designed to showcase all the features of the application and support comprehensive testing of the invoicing, billing, and commission calculation functionalities.
