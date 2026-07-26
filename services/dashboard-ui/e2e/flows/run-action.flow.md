# Flow: Run an action

From an install's actions list, opens the `healthcheck` action and triggers a manual run, verifying it kicks off an action workflow and redirects. Asserts up to workflow creation (like the create-install flow) — it does NOT wait for the run to complete, so it works in a sandbox-mode install without a live runner.

## Setup
- requires: at least one seed install (installIds[0]) — skip otherwise
- start: /:orgId/installs/{installId}/actions

## Steps

### Navigate to the actions list
- action: goto | /:orgId/installs/{installId}/actions
- action: wait | domcontentloaded
- expect: title | /^Actions \|/

### Open the healthcheck action detail
- action: click | link "healthcheck"
- expect: visible | button "Run action"

### Open the run modal
- action: click | button "Run action" first
- expect: visible | heading "Run action healthcheck"

### Trigger the run
- action: click | button "Run action" last
- expect: url | /workflows/
- expect: visible | text "Action workflow started"
