# Flow: Navigation

Verifies the major org-level pages in the main nav are reachable and render correctly. Each page sets a distinct browser tab title via `PageTitle`, used as the "rendered correctly" assertion. Pages correspond to the links in `client/components/navigation/main-nav-links.ts`.

## Setup
- start: /:orgId

## Steps

### Dashboard page
- action: goto | /:orgId
- action: wait | domcontentloaded
- expect: title | /^Dashboard \|/

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

### Webhooks page
- action: goto | /:orgId/webhooks
- action: wait | domcontentloaded
- expect: title | /^Webhooks \|/

### API tokens page
- action: goto | /:orgId/api-tokens
- action: wait | domcontentloaded
- expect: title | /^API tokens \|/

### Slack page
- action: goto | /:orgId/slack
- action: wait | domcontentloaded
- expect: title | /^Slack \|/
