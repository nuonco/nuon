# Flow: Install management toggle (auto-approve)

Toggles auto-approval on and back off using the "Auto approval" switch on the install's Workflows page, verifying the confirm modal and both toasts. State-restoring (ends disabled), so it's safe to run against a shared install in parallel. Pure API/DB flow.

## Setup
- requires: at least one seed install (installIds[0]) — skip otherwise
- note: the seed installs use `approval_option: "prompt"` (auto-approve off) so the enable path is available
- start: /:orgId/installs/{installId}/workflows

## Steps

### Navigate to the workflows page
- action: goto | /:orgId/installs/{installId}/workflows
- action: wait | domcontentloaded
- expect: title | /^Workflows \|/

### Enable auto-approval
- action: click | switch "Auto approval"
- action: click | button "Enable auto approval"
- expect: visible | text "Auto approve enabled"

### Disable auto-approval (restore)
- action: click | switch "Auto approval"
- action: click | button "Disable auto approval"
- expect: visible | text "Auto approve disabled"
