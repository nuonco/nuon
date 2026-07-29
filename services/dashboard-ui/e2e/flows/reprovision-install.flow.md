# Flow: reprovision-install (destructive)

Reprovision an install from the install Settings panel. Asserts up to workflow creation only.

## Setup
- fixtures: orgId
- isolation: throwaway install (helpers.createThrowawayInstall). Sandbox installs stay
  `queued` (no runner), but the reprovision endpoint returns a workflow regardless.
- start: /:orgId/installs/:throwawayInstallId?panel=settings

## Steps

### open the settings panel
- action: goto | /:orgId/installs/:throwawayInstallId?panel=settings
- note: `?panel=settings` auto-opens the Settings panel (role="complementary").

### open the reprovision modal
- action: click | panel button "Reprovision install" (scope to role="complementary")
- expect: visible | dialog text "Reprovision install?"

### reprovision
- action: click | dialog button "Reprovision install"
- expect: url | /workflows

## Notes
- Panel is role="complementary"; the confirm modal is role="dialog".
- Do NOT assert the toast — it races with navigation.
