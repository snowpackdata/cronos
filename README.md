# Cronos
Cronos is our internal timekeeping system, the core logic of which we are open sourcing.

The goal of cronos is to provide simple visibility into internal accounts, users, billing codes, and worked time. The end goal will be to provide a smooth AP/AR system that can be used both internally and externally at Snowpack.

The internal APIs and website logic is modeled in our internal website so only the core logic of the application is available here.

## Features

### Impersonation
Cronos supports an "Impersonate" feature for timesheet entries, allowing one staff member to create timesheet entries that appear to be done by another staff member to clients. This is useful in scenarios where work needs to be attributed to a specific team member for client-facing purposes, while ensuring the actual creator gets paid internally.

How it works:
- When creating a timesheet entry, a user can optionally choose to "impersonate" another staff member
- On client-facing invoices, the impersonated staff member's name will appear instead of the actual creator
- For internal billing and payment purposes, the entry is still attributed to the original creator
- Both the original creator and impersonated user can view the entry in their respective dashboards

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
