# Flow: Install quick management dropdown

Opens the quick management (three-dots) dropdown on an install row in the installs table and verifies the management actions behave: Edit inputs opens its modal without crashing, and Current inputs / View state navigate to the correct pages. Regression coverage for the missing `InstallAppConfigProvider` crash and the "Current inputs" busted-link bug.

## Setup
- requires: at least one seed install (installIds[0]) — skip otherwise
- start: /:orgId/installs

## Steps

### Edit inputs opens without crashing
- action: goto | /:orgId/installs
- action: wait | domcontentloaded
- expect: visible | heading "Installs"
- action: click | .locator "#dropdown-button-{installId}"
- action: click | button "Edit inputs"
- expect: visible | text "Edit install inputs"
- expect: not-visible | text "Something went wrong."

### Current inputs navigates to the inputs page
- action: goto | /:orgId/installs
- action: wait | domcontentloaded
- action: click | .locator "#dropdown-button-{installId}"
- action: click | link "Current inputs"
- expect: url | /installs/{installId}/inputs
- expect: visible | text "The current input values for this install."

### View state navigates to the state page
- action: goto | /:orgId/installs
- action: wait | domcontentloaded
- action: click | .locator "#dropdown-button-{installId}"
- action: click | link "View state"
- expect: url | /installs/{installId}/state
