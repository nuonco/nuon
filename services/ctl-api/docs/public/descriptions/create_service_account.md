Create a service account for the current org. Service accounts can be used to
generate API tokens for automation and CI/CD workflows.

Both `name` and `role` are required. Choose a customer role returned by
`GET /v1/roles?context=service_account`.

Omitting `scope` assigns an org role at the current org. App and install roles
require a matching scope, for example `{"kind":"app","id":"<app_id>"}` or
`{"kind":"install","id":"<install_id>"}`. The scope's `org_id` defaults to the
current org and cannot name a different org. The resource must exist in that org.

Only Org Admin can create service accounts, and the assigned role's full permissions
must fit the caller's live authority and credential ceiling at that scope. App and
Install Admin cannot create accounts or delegate roles. Managed workload roles are
not assignable through this API.
