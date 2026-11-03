Create an API token for a service account in the current org.

Only service-account tokens are supported: `account_id` must be a service
account in the current org (see `POST /v1/service-accounts`). Managed
workload identities must use their owning issuer instead.

Both `role` and `scope` are required. Choose one of the nine customer
roles from `GET /v1/roles?context=api_token`, with its matching scope:
org roles require an org, app roles an app, and install roles an install.
Provide `scope.kind` and `scope.id`; `scope.org_id` may be omitted and
defaults to the selected org. The target must exist in that org.

For example, `{"account_id":"<account-id>","role":"app_reader",
"scope":{"kind":"app","id":"<app-id>"},"name":"acme-ci"}`.

Issuance requires `org.write`. Every permission in the requested role
must be covered at the target by the service account's live bindings,
the issuer's live bindings, and the issuer's credential limits. App and
Install Admin cannot delegate. A token records an immutable permission
ceiling for this one scope; it does not copy all of the account's bindings.
Requests remain constrained by current account grants, so later promotion
does not expand a token and removal of access takes effect immediately.

The token secret is returned only in this response; it cannot be
retrieved again. Defaults to a duration of one year (`8760h`) if
`duration` is not specified; durations must be positive and at most one
year. Non-interactive issuers cannot extend beyond their own expiry.
If `invalidate` is set, all existing credentials for the service account
and their descendants are revoked atomically with replacement. Failure
leaves the existing credentials untouched. Rotation cannot revoke its
own authorizing credential.

Deprecated creation endpoints return `410 Gone`; use this endpoint.
Existing tokens are not upgraded by issuance. Historical credentials
require independently evidenced limits during offline migration or
replacement before the v3 cutover; present-day roles are not that evidence.
