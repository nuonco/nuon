List the API tokens for service accounts in the current org. Only
service-account tokens are returned, and token secrets are never
returned.

Filter to a single service account with `account_id`, and search with
`q`, which matches token names case-insensitively or a token ID exactly.

Each result includes the owning account's `account_name` and
`account_email`, plus `account_role_bindings`, its current bindings in
this org. `account_role` summarizes only an org-scoped binding and is
empty for accounts with only app/install bindings.

`role` is the label recorded at issuance, not the account's current
authority. `scope` is returned when the persisted credential ceiling
identifies a single scope; historical credentials without explicit
limits have no inferred scope. Neither these labels nor current bindings
replace the immutable stored ceiling used for authorization.
