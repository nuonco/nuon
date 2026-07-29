# Flow: sandbox-management (destructive)

Sandbox controls on the install sandbox page: drift scan, reprovision, and deprovision.
Each asserts up to workflow creation only (redirect to the workflows page). Run in
order of blast radius: drift → reprovision → deprovision (terminal).

## Setup
- fixtures: orgId
- isolation: throwaway install (helpers.createThrowawayInstall + waitForInstallStatus).
- start: /:orgId/installs/:throwawayInstallId/sandbox

## Steps (repeat per action)

### open sandbox page
- action: goto | /:orgId/installs/:throwawayInstallId/sandbox
- expect: title | /^Sandbox \|/

### drift scan
- action: click | button "Sandbox controls" → button "Drift scan sandbox"
- expect: visible | dialog heading "Drift scan sandbox"
- action: click | dialog button "Drift scan sandbox"
- expect: url | /workflows

### reprovision
- action: click | button "Sandbox controls" → button "Reprovision sandbox"
- expect: visible | dialog heading "Reprovision sandbox?"
- action: click | dialog button "Reprovision sandbox"
- expect: url | /workflows

### deprovision
- action: click | button "Sandbox controls" → button "Deprovision sandbox"
- expect: visible | dialog heading "Deprovision sandbox?"
- action: fill | confirm input placeholder "deprovision" | deprovision
- action: click | dialog button "Deprovision sandbox"
- expect: url | /workflows

## Notes
- Deprovision requires typing the literal string "deprovision" into the confirm input.
- Menu items and confirm buttons share text — scope confirms to the dialog.
