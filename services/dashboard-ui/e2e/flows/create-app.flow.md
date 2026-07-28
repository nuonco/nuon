# Flow: Create app

Creates a new app from the apps page and verifies the redirect to its branches page. Pure API/DB flow.

## Setup
- env: E2E_ORG_ID (required)
- note: global-setup enables the `app-branches-ui` org feature on the created test org
- start: /:orgId/apps

## Steps

### Navigate to apps page
- action: goto | /:orgId/apps
- action: wait | domcontentloaded
- expect: title | /^Apps \|/

### Open the create app modal
- action: click | button "Create app" first
- expect: visible | text "Create app"

### Fill the name and submit
- action: fill | input "my-app" | e2e-app-{timestamp}
- action: click | button "Create" last
- expect: visible | text "App created"

### Redirected to the app's branches page
- expect: url | /apps/.*/branches
