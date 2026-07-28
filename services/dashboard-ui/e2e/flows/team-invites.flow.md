# Flow: Team invites lifecycle

Invites a user to the org, verifies the pending invite appears, resends it, then revokes it. Pure API/DB flow — the invite is created without real email delivery in a local sandbox, and the flow is self-cleaning.

## Setup
- env: E2E_ORG_ID (required)
- start: /:orgId/team

## Steps

### Navigate to team page
- action: goto | /:orgId/team
- action: wait | domcontentloaded
- expect: title | /^Team \|/

### Invite a user
- action: click | button "Invite user" first
- expect: visible | text "Invite team member"
- action: fill | input "user@email.com" | e2e-invite-{timestamp}@example.com
- action: click | button "Invite user" last
- expect: visible | text "Invitation sent"

### Pending invite appears
- expect: visible | text "e2e-invite-{timestamp}@example.com"

### Re-inviting the same email surfaces the existing invite
- action: click | button "Invite user" first
- expect: visible | text "Invite team member"
- action: fill | input "user@email.com" | e2e-invite-{timestamp}@example.com
- expect: visible | text /already has a pending invite/i
- expect: visible | button "Resend invite"
- action: click | button "Resend invite"
- expect: visible | text "Invite resent"

### Resend the invite
- action: click | button "Resend" first
- expect: visible | text "Resend invite"
- action: click | button "Resend invite"
- expect: visible | text "Invite resent"

### Revoke the invite
- action: click | button "Revoke" first
- expect: visible | text "Revoke invite?"
- action: click | button "Revoke invite"
- expect: visible | text "Invite revoked"
- expect: not-visible | text "e2e-invite-{timestamp}@example.com"
