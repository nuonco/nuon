# Flow: Create install

Opens the create install modal, selects an app, fills out the form, submits, and verifies redirect to the provision workflow page.

## Setup
- env: E2E_ORG_ID (required)
- start: /:orgId/installs

## Steps

### Navigate to installs page
- action: goto | /:orgId/installs
- action: wait | networkidle
- expect: visible | heading "Installs"

### Open create modal
- action: click | button "Create install"
- expect: visible | heading "Create install"
- expect: visible | text "Select an app to create an install"

### App list loads with search
- expect: visible | input "Search apps..."

### Select first app
- action: click | radio "app-selection" first
- expect: not-visible | text "Select an app to create an install"

### Form loads with required fields
- expect: visible | text "Install name"
- expect: visible | input "Enter install name"

### Fill install name
- action: fill | input "Enter install name" | e2e-test-install

### Select AWS region (if AWS app)
- action: click | text "Choose AWS region"
- action: click | text "us-west-2"

### Submit the form
- action: click | button "Create install"
- action: wait | networkidle

### Redirected to provision workflow
- expect: url | /workflows/
- expect: visible | text "Install created successfully"
