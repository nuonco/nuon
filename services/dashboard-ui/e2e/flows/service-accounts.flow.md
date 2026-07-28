# Flow: Service accounts lifecycle

Creates a service account, mints a token for it, renames it, changes its role, then deletes it. Pure API/DB flow — no infra dependency, self-cleaning.

## Setup
- env: E2E_ORG_ID (required)
- start: /:orgId/service-accounts

## Steps

### Navigate to service accounts page
- action: goto | /:orgId/service-accounts
- action: wait | domcontentloaded
- expect: title | /^Service accounts \|/

### Create a service account
- action: click | button "Create service account" first
- expect: visible | text "Create service account"
- action: fill | input "e.g. ci-deploy" | e2e-sa-{timestamp}
- action: click | button "Create service account" last
- expect: visible | text "Service account created"
- expect: visible | text "e2e-sa-{timestamp}"

### Mint a token for it
- action: click | .locator | row menu trigger for the e2e-sa-{timestamp} row
- action: click | button "Create token"
- expect: visible | text "Token created"
- action: click | button "Done"

### Rename it
- action: click | .locator | row menu trigger for the e2e-sa-{timestamp} row
- action: click | button "Rename"
- expect: visible | text "Rename service account"
- action: fill | input "e.g. ci-deploy" | e2e-sa-renamed-{timestamp}
- action: click | button "Save"
- expect: visible | text "Service account renamed"
- expect: visible | text "e2e-sa-renamed-{timestamp}"

### Change its role
- action: click | .locator | row menu trigger for the e2e-sa-renamed-{timestamp} row
- action: click | button "Change role"
- expect: visible | text "Change role"
- action: click | button "Save"
- expect: visible | text "Role updated"

### Delete it (type-to-confirm the identity)
- action: click | .locator | row menu trigger for the e2e-sa-renamed-{timestamp} row
- action: click | button "Delete service account"
- expect: visible | text "Delete service account?"
- action: fill | input "service account identity" | <the account's identity, read from its row>
- action: click | button "Delete service account"
- expect: visible | text "Service account deleted"
- expect: not-visible | text "e2e-sa-renamed-{timestamp}"
