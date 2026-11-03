Returns only installs in the selected org permitted by `install.read` in live v3 grants and both credential ceilings.
Filtering precedes search, pagination and lookahead in a read-only snapshot; org-wide read access is not required.
Deleted parents and sibling or cross-tenant installs outside the effective scopes are excluded. Embedded parent-app
data contains only ID and name; it does not grant app-wide read access. Health counts and label facets use the same
install visibility rule.
