List the API tokens for service accounts in the current org. Only
service-account tokens are returned, and token secrets are never
returned.

Filter to a single service account with `account_id`, and search with
`q`, which matches token names case-insensitively or a token ID exactly.

Each result includes the owning account's `account_name` and
`account_email`, plus `account_role`, the account's *current* role in this
org (not the role copied onto the token at creation time).
