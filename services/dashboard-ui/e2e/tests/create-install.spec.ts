import { test, expect } from '../fixtures/auth.fixture';

test.describe('Create Install Flow', () => {
  test('should complete happy path - create install for first app', async ({ authenticatedPage }) => {
    // Step 1: Navigate to root - dashboard will redirect to /:org-id/apps automatically
    await authenticatedPage.goto('/');
    // Wait for URL to change instead of networkidle (app has polling)
    await authenticatedPage.waitForURL(/\/org[a-zA-Z0-9]+\/apps/, { timeout: 10000 });

    // Verify we're redirected to an org's apps page
    await expect(authenticatedPage).toHaveURL(/\/org[a-zA-Z0-9]+\/apps/);

    // Step 2: Click first app in table to navigate to app details
    // Get the href from the first app link and navigate directly
    const appTable = authenticatedPage.locator('table').first();
    const firstAppNameLink = appTable.locator('tbody tr').first().locator('a').first();
    await expect(firstAppNameLink).toBeVisible();

    // Get the href and navigate to it (more reliable than clicking for Next.js apps)
    const appHref = await firstAppNameLink.getAttribute('href');
    if (!appHref) {
      throw new Error('App link does not have href attribute');
    }

    await authenticatedPage.goto(appHref);
    // Wait for URL to change instead of networkidle (app has polling)
    await authenticatedPage.waitForURL(/\/org[a-zA-Z0-9]+\/apps\/app[a-zA-Z0-9]+/, { timeout: 10000 });

    // Verify we're on an app detail page
    await expect(authenticatedPage).toHaveURL(/\/org[a-zA-Z0-9]+\/apps\/app[a-zA-Z0-9]+/);

    // Step 3: Open create install modal
    const createInstallButton = authenticatedPage.getByRole('button', { name: 'Create install' });
    await expect(createInstallButton).toBeVisible();
    await createInstallButton.click();

    // Verify modal opened by checking for the form field (wait up to 15s for loading to complete)
    await expect(authenticatedPage.locator('input[placeholder="Enter install name"]')).toBeVisible({ timeout: 15000 });

    // Step 4: Fill install form
    const installName = `test-install-${Date.now()}`;

    // Fill install name field
    const installNameInput = authenticatedPage.locator('input[placeholder="Enter install name"]');
    await expect(installNameInput).toBeVisible();
    await installNameInput.fill(installName);

    // Select AWS region from dropdown using keyboard navigation
    // Find the dropdown by its text content
    const regionDropdown = authenticatedPage.getByText('Choose AWS region');
    await expect(regionDropdown).toBeVisible();

    // Focus on the dropdown and use keyboard to select first option
    await regionDropdown.click(); // Click to focus
    await authenticatedPage.keyboard.press('Space'); // Open dropdown
    await authenticatedPage.keyboard.press('Tab');   // Navigate to first option
    await authenticatedPage.keyboard.press('Tab');   // Navigate to second item (first actual option)
    await authenticatedPage.keyboard.press('Enter'); // Select the option

    // Brief wait for selection to register
    await authenticatedPage.waitForTimeout(500);

    // Step 5: Submit form
    // Use .last() to target the modal's submit button (not the page button)
    const submitButton = authenticatedPage.getByRole('button', { name: 'Create install' }).last();
    await submitButton.click();

    // Wait for navigation after form submission
    await authenticatedPage.waitForTimeout(1000);

    // Step 6: Verify redirect to workflow page
    // Test stops here - workflow execution is separate test
    await expect(authenticatedPage).toHaveURL(/workflow/, { timeout: 10000 });

    // Take screenshot for verification
    await authenticatedPage.screenshot({
      path: 'e2e-results/create-install-success.png',
      fullPage: true
    });
  });

  test('should show validation error for empty install name', async ({ authenticatedPage }) => {
    // Navigate to root - dashboard will redirect to /:org-id/apps
    await authenticatedPage.goto('/');
    await authenticatedPage.waitForURL(/\/org[a-zA-Z0-9]+\/apps/, { timeout: 10000 });

    // Get first app href and navigate to it
    const appTable = authenticatedPage.locator('table').first();
    const firstAppNameLink = appTable.locator('tbody tr').first().locator('a').first();
    const appHref = await firstAppNameLink.getAttribute('href');
    if (appHref) {
      await authenticatedPage.goto(appHref);
      await authenticatedPage.waitForURL(/\/org[a-zA-Z0-9]+\/apps\/app[a-zA-Z0-9]+/, { timeout: 10000 });
    }

    // Open create install modal
    await authenticatedPage.getByRole('button', { name: 'Create install' }).click();
    // Verify modal opened by checking for form field (wait up to 15s for loading to complete)
    await expect(authenticatedPage.locator('input[placeholder="Enter install name"]')).toBeVisible({ timeout: 15000 });

    // Try to submit without filling required fields
    const submitButton = authenticatedPage.getByRole('button', { name: 'Create install' }).last();
    await submitButton.click();

    // Expect validation error to appear (adjust selector based on actual error display)
    // Common patterns: error message, required field indicator, disabled submit
    const errorMessage = authenticatedPage.getByText(/required/i);
    await expect(errorMessage).toBeVisible({ timeout: 5000 });
  });

  test('should close modal when clicking Cancel button', async ({ authenticatedPage }) => {
    // Navigate to root - dashboard will redirect to /:org-id/apps
    await authenticatedPage.goto('/');
    await authenticatedPage.waitForURL(/\/org[a-zA-Z0-9]+\/apps/, { timeout: 10000 });

    // Get first app href and navigate to it
    const appTable = authenticatedPage.locator('table').first();
    const firstAppNameLink = appTable.locator('tbody tr').first().locator('a').first();
    const appHref = await firstAppNameLink.getAttribute('href');
    if (appHref) {
      await authenticatedPage.goto(appHref);
      await authenticatedPage.waitForURL(/\/org[a-zA-Z0-9]+\/apps\/app[a-zA-Z0-9]+/, { timeout: 10000 });
    }

    // Open create install modal
    await authenticatedPage.getByRole('button', { name: 'Create install' }).click();
    // Verify modal opened by checking for form field (wait up to 15s for loading to complete)
    await expect(authenticatedPage.locator('input[placeholder="Enter install name"]')).toBeVisible({ timeout: 15000 });

    // Click Cancel button
    const cancelButton = authenticatedPage.getByRole('button', { name: 'Cancel' });
    await expect(cancelButton).toBeVisible();
    await cancelButton.click();

    // Verify modal is closed (form field should not be visible)
    await expect(authenticatedPage.locator('input[placeholder="Enter install name"]')).not.toBeVisible();
  });
});
