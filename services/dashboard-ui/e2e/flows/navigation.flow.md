# Flow: Navigation

Verifies major pages are reachable and render correctly.

## Setup
- env: E2E_ORG_ID (required)
- start: /:orgId

## Steps

### Apps page
- action: goto | /:orgId/apps
- action: wait | networkidle
- expect: visible | heading "Apps"

### Installs page
- action: goto | /:orgId/installs
- action: wait | networkidle
- expect: visible | heading "Installs"

### Runners page
- action: goto | /:orgId/runners
- action: wait | networkidle
- expect: visible | heading "Runners"

### Team page
- action: goto | /:orgId/team
- action: wait | networkidle
- expect: visible | heading "Team"
