# Flow: deploy-component

Core product flow. Deploy a build of an install component. Asserts up to workflow
creation only (redirect to the workflows page) — the deploy workflow itself needs
real cloud and is not awaited.

## Setup
- fixtures: orgId, installIds
- install: installIds[0] (non-destructive, additive)
- start: /:orgId/installs/:installId/components

## Steps

### open components page
- action: goto | /:orgId/installs/:installId/components
- expect: title | /^Components \|/

### open the component row action menu
- action: click | first `[id^="dropdown-button-component-quick-"]` trigger

### open the deploy modal
- action: click | button "Deploy component"
- expect: visible | dialog text "Deploy" + "component"

### select a build
- note: BuildSelect auto-selects the most recent active build; only active builds are selectable
- action: check | first enabled `input[name="build-selection"]`
- skip-if: no active build becomes selectable within 30s

### deploy
- action: click | dialog button "Deploy build"
- expect: url | /workflows

## Notes
- Confirm button ("Deploy build") differs from the menu item ("Deploy component"), but scope the
  confirm click to the dialog anyway.
- Do NOT assert the "Deploying component" toast — it races with navigation to the workflows page.
