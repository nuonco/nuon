Retired: returns `410 Gone`. Use `POST /v1/tokens` with explicit
`account_id`, `role` and `scope` instead. No token is created by this
endpoint. Existing tokens can still be listed and revoked.
