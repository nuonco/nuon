List the service accounts that belong to the current organization, along with
their roles. Supports offset-based pagination.

Each account reports whether it is a system account, its purposes, and, for
accounts managed by a resource, the owning resource.

- `management` filters to `user` accounts, `system` accounts, or `all`. When set,
  it replaces `include_runners` and `include_stacks`.
- `purpose` filters to accounts with that exact purpose.
- `q` matches a case-insensitive substring of the account name, email, or ID, or
  of the owning resource's name or ID.
