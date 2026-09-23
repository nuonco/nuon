# Flow: create-trigger

## Setup
- env: E2E_ORG_ID (required)
- start: /:orgId/settings/triggers

## Steps

### Load the triggers page
- action: goto | /:orgId/settings/triggers
- expect: visible | heading "Triggers"

### Open the create-trigger modal
- action: click | button "Create trigger"
- expect: visible | heading "Create trigger"

### Fill the trigger name
- action: fill | input "Name" | e2e-trigger
- expect: visible | heading "Create trigger"

### Submit the trigger
- action: click | button "Create trigger"
- expect: url | /settings/triggers/

### Confirm the trigger detail page
- expect: visible | heading "e2e-trigger"
