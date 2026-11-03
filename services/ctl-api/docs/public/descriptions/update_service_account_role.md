Update the role assigned to a service account in the current org.

The required `role` replaces bindings at one scope, preserving bindings at other
scopes. Omit `scope` to replace the org-scoped binding. App and install roles require
an explicit matching scope, such as `{"kind":"app","id":"<app_id>"}` or
`{"kind":"install","id":"<install_id>"}`. The optional `org_id` must match the
current org, and the resource must belong to that org.

Choose a role from `GET /v1/roles?context=service_account`. Only Org Admin can
assign roles, and the delegated permissions must fit both live grants and the
credential ceiling at the requested scope. Managed service accounts cannot be
modified here. An app-scoped assignment does not remove an existing org-scoped
grant or narrow tokens already issued; token ceilings are checked separately.

To remove one binding, send `remove: true` with its `role` and an explicit `scope`.
This removes only that role at that scope; other bindings are unchanged. Removal
requires `org.write`, but does not require permission to grant the removed role.
A missing binding returns 404. You cannot remove your own binding. Removing the final usable binding is rejected;
use the delete-service-account operation to retire the identity and its tokens.
Responses include the remaining current-org `role_bindings` with binding ID,
role ID, role type and scope. Existing credentials remain limited by current grants.
