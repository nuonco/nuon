# Flow: deprovision-install (destructive)

Deprovision an install from the install Settings panel. Type-to-confirm the install name.
Asserts up to workflow creation only.

## Setup
- fixtures: orgId
- isolation: throwaway install (helpers.createThrowawayInstall).
- start: /:orgId/installs/:throwawayInstallId?panel=settings

## Steps

### open the settings panel
- action: goto | /:orgId/installs/:throwawayInstallId?panel=settings
- note: `?panel=settings` auto-opens the Settings panel (role="complementary").

### open the deprovision modal
- action: click | panel button "Deprovision install" (scope to role="complementary")
- expect: visible | dialog text "Deprovision install?"

### type-to-confirm
- action: fill | `#confirm-install-name` | <throwaway install name>

### deprovision
- action: click | dialog button "Deprovision install"
- expect: url | /workflows

## Notes
- Panel is role="complementary"; the confirm modal is role="dialog".
- The confirm input requires the exact install name (known from creation).
