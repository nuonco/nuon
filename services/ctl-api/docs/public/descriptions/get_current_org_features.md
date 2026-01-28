Get the current organization's feature flag values.

Returns a map of feature flag names to their enabled/disabled status for the authenticated organization.

This endpoint shows which features are currently enabled or disabled for your organization, unlike `/v1/orgs/features` which returns all available features with their descriptions.

Example response:
```json
{
  "api-pagination": true,
  "org-dashboard": false,
  "org-runner": true,
  "stratus-layout": true,
  "user-managed-features": false
}
```
