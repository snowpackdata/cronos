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

## Quick Start

### Local Development (Fastest Way)

```bash
cd server
make run-local
```

Visit `http://localhost:8080` and login with:
- **Staff**: `dev@example.com` / `devpassword`
- **Client**: `client@example.com` / `password`

### See All Available Commands

```bash
cd server
make help
```

### Full Documentation

- **Local Development Guide**: [`server/LOCAL_DEVELOPMENT.md`](server/LOCAL_DEVELOPMENT.md) - Complete local setup
- **Deployment Guide**: [`DEPLOYMENT.md`](DEPLOYMENT.md) - Production deployment to App Engine
- **Setup Summary**: [`SETUP_SUMMARY.md`](SETUP_SUMMARY.md) - What's been configured

## Installation

### Prerequisites
- Go 1.21 or higher
- Node.js 20+ and npm (for admin and portal UIs)
- Google Cloud SDK (for deployment)
- Cloud SQL Proxy (optional, for development with Cloud SQL)

### Quick Setup

1. Clone this repository
2. Navigate to the server directory:
```bash
cd server
```

3. Install all dependencies:
```bash
make install
```

4. Start the server:
```bash
make run-local
```

This will:
- Build the Vue.js admin and portal interfaces
- Start the Go backend on port 8080
- Initialize SQLite database with seed data

5. Access the application:
- Landing Page: http://localhost:8080/
- Login: http://localhost:8080/login
- Admin Interface: http://localhost:8080/admin/
- Portal Interface: http://localhost:8080/portal/

### Development Modes

**Quick start (SQLite):**
```bash
make run-local
```

**Hot reload development (best for active coding):**
```bash
make run-dev-hot
```

**All available commands:**
```bash
make help
```

### Development Credentials

When running in local mode, the following test accounts are created:
- **Staff User**: `dev@example.com` / `devpassword` (Admin access)
- **Client User**: `client@example.com` / `password` (Portal access)

## Deployment

Deployment to Google App Engine is automated via GitHub Actions. See [`DEPLOYMENT.md`](DEPLOYMENT.md) for complete instructions.

### Manual Deployment

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
