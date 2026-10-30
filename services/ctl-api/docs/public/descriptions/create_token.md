Create an API token for a service account in the current org.

Only service-account tokens are supported: `account_id` must be a service
account in the current org (see `POST /v1/service-accounts`). The token
uses the service account's role.

The token secret is returned only in this response; it cannot be
retrieved again. Defaults to a duration of one year (`8760h`) if
`duration` is not specified. If `invalidate` is set, all existing tokens
for the service account are revoked before the new token is created.
