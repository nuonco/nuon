# Flow: teardown-component (destructive)

Teardown a single install component from the component-detail "Component controls"
dropdown. Type-to-confirm the component name. Asserts up to workflow creation only.

## Setup
- fixtures: orgId
- isolation: creates a throwaway install via the API (helpers.createThrowawayInstall),
  waits for it to provision, acts on its first component. Global teardown drops the org.
- start: /:orgId/installs/:throwawayInstallId/components

## Steps

### open a component detail page
- action: goto | components list → click first component name link
- expect: url | /components/

### open Component controls
- action: click | button "Component controls"

### open the teardown modal
- action: click | button "Teardown component"
- expect: visible | dialog text /^Teardown .+\?$/

### type-to-confirm
- action: read component name from the modal heading, fill | `#confirm-component-name`

### teardown
- action: click | dialog button "Teardown component"
- expect: url | /workflows

## Notes
- Teardown is disabled when the component is already torn down (status "inactive").
- Menu item and confirm both read "Teardown component" — scope the confirm to the dialog.
