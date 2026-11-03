Returns organization IDs and names for navigation, filtered by the authenticated principal's live v3 grants and both
request and persisted credential limits. App- or install-scoped access can expose its parent org's ID and name without
granting org-wide read access. Membership alone is not sufficient.

Authorization and the paginated query share one read-only snapshot. Search and pagination apply after authorization;
inaccessible orgs do not affect page boundaries or next-page indicators. Results are ordered case-insensitively by
name, then ID. Search matches a literal case-insensitive substring of name or ID.

This endpoint no longer returns full org objects, including settings, features, notification configuration or other
tenant metadata. Use the org detail endpoint with `org.read` for that data. Clients must adopt the ID/name response
before the v3 cutover. A valid credential with no effective access receives an empty list; revoked, expired, malformed
or unsupported credentials are rejected. No org-selection header is required.
