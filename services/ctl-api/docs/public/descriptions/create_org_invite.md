Invite a human user by email with one role at one org, app, or install scope.

`role_type` accepts any of the nine customer roles at its matching scope. Omitting it defaults to Org Read-only.
Omitting `scope` defaults to the selected org. An explicit scope requires `kind` and `id`; `org_id` defaults to
the selected org and cannot select another tenant. App and install roles require an explicit matching scope.

Creating or resending requires `org.write` and authority to delegate the complete role through both the caller's
current grants and credential limits. App/Install Admin cannot delegate. The response and invite inventory include
`scope`; pending member inventory includes `invite_scope`, not a live role binding. Private delegation proof is never
returned.

Acceptance requires a live, direct human login session with a matching current email. It rechecks the inviter's
recorded credential, current grants and limits, and the target's existence and ownership. Expiry, revocation, or
loss of authority blocks pending acceptance until an authorized resend. Acceptance never changes an existing
member's bindings. A new member receives only the recorded role and scope, not an additional org role.

Acceptance does not expand an already-issued session or OAuth token's limits. Sign in again after acceptance to
obtain a credential reflecting the new binding. Invites without explicit migrated scope cannot be accepted or resent.
