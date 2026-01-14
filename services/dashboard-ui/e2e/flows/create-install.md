# Create Install Flow

**Priority**: High
**Status**: ✅ Automated
**Test File**: [`tests/create-install.spec.ts`](../tests/create-install.spec.ts)

## Prerequisites

Before starting this flow, ensure:
- [ ] User is authenticated
- [ ] User has an org created
- [ ] User has at least one app with valid app config
- [ ] User starts at root path `/` (dashboard will redirect to `/:org_id/apps`)

## User Flow

### Step 1: View Apps Page

*(Optional screenshot: `screenshots/apps-page.png`)*

- **User Action**: Navigate to root path `/` - dashboard automatically redirects to `/:org_id/apps`
- **Expected Result**: User is redirected to apps page with apps table visible (at least one app listed)
- **Key Selectors**:
  - Apps table with columns: App name, Config version, Sandbox, Platform
  - Each app row has a "View" link on the right side
- **URL Pattern**: `/:org_id/apps` (org_id is dynamic, determined by dashboard redirect)

### Step 2: Click App Name to Navigate to App Details

*(Optional screenshot: `screenshots/apps-page.png`)*

- **User Action**: Click the first app name link in the apps table (e.g., "httpbin")
- **Expected Result**: Navigates to app detail page `/:org_id/apps/:app_id` for that app
- **Key Selectors**:
  - App name link (first link in first table row): `table tbody tr:first a:first`
  - Alternative: Can also click "View >" link on the right side of the row
- **Note**: Uses generic selector targeting first app in table - works for any app name

### Step 3: Open Create Install Modal

*(Optional screenshot: `screenshots/app-details-page.png`)*

- **User Action**: Click the purple "Create install" button in the top right corner of the page
- **Expected Result**: Modal opens with heading "Create install"
- **Key Selectors**:
  - Button: `page.getByRole('button', { name: 'Create install' })`

### Step 4: Fill Install Form

*(Optional screenshot: `screenshots/create-install-form.png`)*

- **User Action**: Enter install name and select AWS region
  - Enter install name in "Install name *" text field (placeholder: "Enter install name")
  - Select AWS region from "Select AWS region *" dropdown (placeholder: "Choose AWS region")
- **Expected Result**: Form accepts input values
- **Key Selectors**:
  - Install name field: Input with label "Install name *"
  - Region dropdown: Dropdown with label "Select AWS region *"
- **Form Structure**:
  - Section header: "Set AWS settings (required)"
  - Both fields are required (marked with *)

### Step 5: Submit Form

*(Optional screenshot: `screenshots/create-install-form.png`)*

- **User Action**: Click the purple "Create install" button at the bottom of the modal
- **Expected Result**: Form submits successfully, modal closes
- **Alternative Actions**:
  - Click "Cancel" button to cancel operation without creating install
  - Click X close button in top right to close modal without creating install
- **Key Selectors**:
  - Submit button: `page.getByRole('button', { name: 'Create install' })` (in modal)
  - Cancel button: `page.getByRole('button', { name: 'Cancel' })`

### Step 6: Verify Redirect to Workflow

- **System Action**: Automatically redirects to provision workflow page
- **Expected Result**: User lands on workflow page (URL contains workflow route)
- **Verification**: `await expect(page).toHaveURL(/workflow/)`
- **Important**: Test stops here - workflow execution is a separate test

---

## Test Data

Example data needed to run this flow:

```typescript
const testInstall = {
  name: `test-install-${Date.now()}`, // Unique name using timestamp
  region: 'us-west-2', // For AWS apps
  // OR location: 'eastus' for Azure apps (depending on platform)
}
```

## Edge Cases to Test

Document scenarios that should be tested beyond the happy path:

- [ ] Empty install name shows validation error
- [ ] Empty region shows validation error
- [ ] Form submission with invalid data shows error message
- [ ] Escape key closes modal without creating install
- [ ] Clicking outside modal closes it without creating install
- [ ] Cancel button closes modal without creating install

## Notes

Any additional context, gotchas, or important information:

- **Screenshots**: Stored in `/services/dashboard-ui/e2e/flows/screenshots/`
  - Screenshots are optional - only add if they clarify complex UI
  - Current screenshots: apps-page.png, app-details-page.png, create-install-form.png
  - Screenshots show "httpbin" app as example, but flow uses generic selectors
- **Prerequisites**: Test assumes valid app config exists (must be set up before test runs)
- **Region field**: Dropdown with AWS region options (e.g., us-east-1, us-west-2)
- **Test boundary**: Test stops at workflow page redirect - workflow execution is a separate test
- **Generic approach**: Flow works for any app in the table (uses first "View" link, not specific app names)

---

## For Developers

**Selector Recommendations**:
- Buttons: `page.getByRole('button', { name: 'Button Text' })`
- Form fields: `page.locator('input[type="text"]')` or target by label
- Headings: `page.getByRole('heading', { name: 'Heading Text' })`
- Links: `page.getByRole('link', { name: 'Link Text' })`
- Dropdowns: `page.locator('select')` or target by label

**Test Structure Example**:

```typescript
import { test, expect } from '../fixtures/auth.fixture';

test.describe('Create Install Flow', () => {
  test('should complete happy path', async ({ authenticatedPage }) => {
    // Step 1: Navigate to root - dashboard redirects to /:org_id/apps
    await authenticatedPage.goto('/'); // Dashboard handles org redirect automatically

    // Step 2: Click first app name link in table
    const appTable = authenticatedPage.locator('table').first();
    const firstAppNameLink = appTable.locator('tbody tr').first().locator('a').first();
    await firstAppNameLink.click();

    // Step 3: Open create install modal
    await authenticatedPage.getByRole('button', { name: 'Create install' }).click();

    // Step 4: Fill form
    await authenticatedPage.locator('input[placeholder="Enter install name"]').fill(`test-install-${Date.now()}`);
    await authenticatedPage.locator('select').selectOption('us-west-2'); // Or appropriate selector

    // Step 5: Submit form
    await authenticatedPage.getByRole('button', { name: 'Create install' }).last().click(); // Use .last() to target modal button

    // Step 6: Verify redirect to workflow
    await expect(authenticatedPage).toHaveURL(/workflow/);
  });

  test('should show validation error for empty install name', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/'); // Navigate to root
    const appTable = authenticatedPage.locator('table').first();
    await appTable.locator('tbody tr').first().locator('a').first().click();
    await authenticatedPage.getByRole('button', { name: 'Create install' }).click();

    // Try to submit without filling required fields
    await authenticatedPage.getByRole('button', { name: 'Create install' }).last().click();

    // Expect validation error (adjust selector based on actual error display)
    await expect(authenticatedPage.getByText(/required/i)).toBeVisible();
  });
});
```
