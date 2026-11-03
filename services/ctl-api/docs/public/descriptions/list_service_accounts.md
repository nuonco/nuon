List the service accounts that belong to the current organization, along with
their roles. Supports offset-based pagination.

Each account reports whether it is a system account, its purposes, and, for
accounts managed by a resource, the owning resource.

`role_bindings` lists the account's assignments in the current org, with each
binding's ID, role ID, role type and scope. Use these bindings rather than the
unscoped `roles` catalog to display access. Bindings from other orgs are not returned.

- `management` filters to `user` accounts, `system` accounts, or `all`. When set,
  it replaces `include_runners` and `include_stacks`.
- `purpose` filters to accounts with that exact purpose.
- `q` matches a case-insensitive substring of the account name, email, or ID, or
  of the owning resource's name or ID.
