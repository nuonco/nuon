# Flow: Navigation

Verifies the major org-level pages are reachable and render correctly. Each page sets a distinct browser tab title via `PageTitle`, used as the "rendered correctly" assertion.

## Setup
- start: /:orgId

## Steps

### Apps page
- action: goto | /:orgId/apps
- action: wait | domcontentloaded
- expect: title | /^Apps \|/

### Installs page
- action: goto | /:orgId/installs
- action: wait | domcontentloaded
- expect: title | /^Installs \|/

### Builds (runner) page
- action: goto | /:orgId/runner
- action: wait | domcontentloaded
- expect: title | /^Builds \|/

### Team page
- action: goto | /:orgId/team
- action: wait | domcontentloaded
- expect: title | /^Team \|/
