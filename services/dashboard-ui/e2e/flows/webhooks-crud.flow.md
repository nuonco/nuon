# Flow: Webhooks CRUD

Creates a webhook with the default scope (everything in the org) and all events, verifies it appears in the table, edits it to set a signing secret, then deletes it. Pure API flow — no infra or Temporal dependency, and self-cleaning (the webhook it creates is deleted in the last step).

## Setup
- env: E2E_ORG_ID (required)
- start: /:orgId/webhooks

## Steps

### Navigate to webhooks page
- action: goto | /:orgId/webhooks
- action: wait | domcontentloaded
- expect: title | /^Webhooks \|/

### Open create modal
- action: click | button "Create webhook" first
- expect: visible | heading "Create webhook"

### Fill the webhook URL (default scope + all events)
- action: fill | input "https://example.com/webhooks/nuon" | https://example.com/e2e-{timestamp}

### Submit and confirm creation
- action: click | button "Create webhook" last
- expect: visible | text "Webhook created"
- expect: visible | text "https://example.com/e2e-{timestamp}"

### Open the edit modal for the new webhook
- action: click | button "Edit" first
- expect: visible | heading "Edit webhook"

### Set a new signing secret and save
- action: click | radio "Set a new secret"
- action: fill | input "webhook-secret" | e2e-secret-value
- action: click | button "Save changes"
- expect: visible | text "Webhook updated"

### Delete the webhook
- action: click | button "Delete" first
- expect: visible | heading "Delete webhook?"
- action: click | button "Delete webhook"
- expect: visible | text "Webhook deleted"
- expect: not-visible | text "https://example.com/e2e-{timestamp}"
