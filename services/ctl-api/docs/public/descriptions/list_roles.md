List your org's roles. Each role carries its display metadata (`title`,
`description`) and the assignment surfaces it may be offered on via the
`applies_to` field (`team`, `service_account`, `api_token`,
`oidc_trust_policy`). A role with no `applies_to` entries exists and may be
displayed, but cannot be newly assigned. Pass `?context=<surface>` to filter
to the roles assignable on a single surface.
