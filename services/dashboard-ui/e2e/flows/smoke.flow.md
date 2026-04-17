# Flow: Smoke

Validates the test infrastructure works: auth cookie injection, page rendering, basic UI.

## Setup
- env: E2E_ORG_ID (required)
- start: /:orgId

## Steps

### Dashboard loads
- action: goto | /:orgId
- action: wait | networkidle
- expect: visible | .logo-link
- expect: visible | text "Nuon"
