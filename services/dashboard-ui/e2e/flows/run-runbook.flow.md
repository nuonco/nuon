# Flow: Run a runbook

From an install's runbooks list, runs the `verify_status` runbook (which takes no inputs, so the run modal skips the input/steps wizard) via the row menu, verifying it kicks off a workflow and redirects. Asserts up to workflow creation — it does NOT wait for the run to complete, so it works in a sandbox-mode install without a live runner.

## Setup
- requires: at least one seed install (installIds[0]) — skip otherwise
- start: /:orgId/installs/{installId}/runbooks

## Steps

### Navigate to the runbooks list
- action: goto | /:orgId/installs/{installId}/runbooks
- action: wait | domcontentloaded
- expect: title | /^Runbooks \|/

### Open the verify_status row menu and start a run
- action: click | .locator | row menu trigger for the verify_status row
- action: click | button "Run runbook"

### Confirm in the run modal
- expect: visible | text "Run verify_status"
- action: click | button "Run runbook" last

### Run kicks off and redirects
- expect: url | /workflows/
- expect: visible | text "Runbook run started"
