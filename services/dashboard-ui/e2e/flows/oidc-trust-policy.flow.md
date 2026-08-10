# Flow: OIDC Trust Policy

Creates an OIDC trust policy using the default GitHub Actions preset (issuer URL and audience are
prefilled), verifies it appears in the table, then (phase 2) opens the edit modal for that same policy,
toggles it disabled, and saves. Pure API flow — no infra or Temporal dependency.

The GitHub Actions preset is the minimal valid create path: the issuer URL and audience are prefilled by
the preset, so only the policy **name** and the **`sub` claim condition value** need to be filled by hand.
The repository dropdown (which would autofill name + `sub`) is skipped because it depends on a connected
VCS org, which the seed org does not have.

## Setup
- env: E2E_ORG_ID (required)
- config: `oidc_federation_enabled` must be true (runtime CLI config, injected into `window.__NUON_CONFIG__`) — the route `/:orgId/settings/oidc` renders NotFound otherwise
- scope: org-scoped (policies belong to the org, not an install)
- start: /:orgId/settings/oidc

## Steps

### Navigate to OIDC federation page
- action: goto | /:orgId/settings/oidc
- action: wait | domcontentloaded
- expect: title | /^OIDC federation \|/
- expect: visible | heading "OIDC federation"

### Open create modal
- action: click | button "Create trust policy" first
- expect: visible | heading "Create OIDC trust policy"

### Fill required fields (GitHub Actions preset — issuer URL and audience are prefilled)
- action: fill | input "Name" | e2e-oidc-{timestamp}
- action: fill | input "sub" | repo:acme/app:ref:refs/heads/main

### Submit and confirm creation
- action: click | button "Create trust policy" last
- expect: visible | text "Trust policy created"
- expect: visible | text "e2e-oidc-{timestamp}"

## Phase 2 — Edit (depends on the policy created above)

> Second phase. Runs only after the create phase succeeds — it edits the policy created in this flow
> (matched by the `e2e-oidc-{timestamp}` name). Do not run standalone; it has no policy to edit otherwise.

### Open the edit modal for the new policy
- action: click | button "Edit" first
- expect: visible | heading "Edit trust policy"

### Toggle the policy disabled and save
- action: click | button "Enabled"
- action: click | button "Save changes"
- expect: visible | text "Trust policy updated"
