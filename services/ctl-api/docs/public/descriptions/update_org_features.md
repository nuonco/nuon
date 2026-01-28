Update feature flags for your current organization.

This endpoint allows organization users to manage feature flags, but requires the `user-managed-features` flag to be enabled for the organization. The `user-managed-features` flag itself cannot be modified through this endpoint and can only be enabled/disabled by administrators.

**Requirements:**
- The `user-managed-features` flag must be enabled for your organization
- You cannot toggle the `user-managed-features` flag through this endpoint (admin-only)

**Example Request:**
```json
{
  "features": {
    "api-pagination": true,
    "install-delete": false
  }
}
```

The request will update only the specified feature flags. Features not included in the request will retain their current values.
