# Flow: Create and delete an API token

Creates an org API token, verifies the one-time token reveal modal, confirms the token appears in the table, then deletes it. Pure API flow — no infra dependency, and self-cleaning.

## Setup
- env: E2E_ORG_ID (required)
- start: /:orgId/api-tokens

## Steps

### Navigate to API tokens page
- action: goto | /:orgId/api-tokens
- action: wait | domcontentloaded
- expect: title | /^API tokens \|/

### Open create modal
- action: click | button "Create token" first
- expect: visible | heading "Create API token"

### Fill token name (default role + expiry)
- action: fill | input "e.g. ci-deploy" | e2e-token-{timestamp}

### Submit and confirm the reveal modal
- action: click | button "Create token" last
- expect: visible | heading "API token created"
- expect: visible | button "Done"

### Dismiss the reveal modal
- action: click | button "Done"

### Token appears in the table
- expect: visible | text "e2e-token-{timestamp}"

### Open the row menu for the new token and delete
- action: click | .locator | row menu trigger for the e2e-token-{timestamp} row
- action: click | button "Delete token"
- expect: visible | heading "Delete API token?"
- action: click | button "Delete token"
- expect: visible | text "Token deleted"
- expect: not-visible | text "e2e-token-{timestamp}"
