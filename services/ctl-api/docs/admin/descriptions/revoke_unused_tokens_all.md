Enqueues the revoke-unused-tokens signal on every runner in a default org that has a live process. When `org_id` is
set, only that org's runners are included, and orgs that are not default orgs are skipped.

See the per-runner endpoint for what each signal revokes and keeps.
