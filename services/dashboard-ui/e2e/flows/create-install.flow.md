# Flow: Create install

Opens the create install modal and verifies initial state.

## Setup
- env: E2E_ORG_ID (required)
- start: /:orgId/installs

## Steps

### Open create modal
- action: goto | /:orgId/installs
- action: wait | networkidle
- action: click | button "Create install"
- expect: visible | heading "Create install"

### App selection visible
- expect: visible | text "Select an app"
