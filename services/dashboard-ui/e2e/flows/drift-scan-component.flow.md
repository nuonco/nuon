# Flow: drift-scan-component

Trigger a drift scan for an install component. Asserts up to workflow creation only
(redirect to the workflows page).

## Setup
- fixtures: orgId, installIds
- install: installIds[0] (non-destructive)
- start: /:orgId/installs/:installId/components

## Steps

### open components page
- action: goto | /:orgId/installs/:installId/components
- expect: title | /^Components \|/

### open the component row action menu
- action: click | first `[id^="dropdown-button-component-quick-"]` trigger

### open the drift scan modal
- action: click | button "Drift scan component"
- expect: visible | dialog text "Drift scan" + "component"

### select a build
- action: check | first enabled `input[name="build-selection"]`
- skip-if: no active build becomes selectable within 30s

### drift scan
- action: click | dialog button "Drift scan build"
- expect: url | /workflows

## Notes
- Confirm button is "Drift scan build"; menu item is "Drift scan component".
- Do NOT assert the toast — it races with navigation.
