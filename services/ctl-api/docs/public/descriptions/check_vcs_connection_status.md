Check the real-time status of a VCS (GitHub App) connection.

This endpoint queries GitHub's API directly to fetch the current installation status, including:
- Active/Suspended state
- Account information
- Permissions
- Suspension details (if applicable)

**Important**: This endpoint always fetches fresh data from GitHub (no caching) to ensure accurate status information.

## Response Status Values

- `active`: The GitHub App installation is active and functioning
- `suspended`: The installation has been suspended (see `suspended_at` and `suspended_by` for details)
- `unknown`: Unable to determine status (GitHub API error - see `error` field)

## Use Cases

- Troubleshooting connection issues
- Monitoring installation health
- Detecting suspended or revoked installations
- Validating permissions before operations
