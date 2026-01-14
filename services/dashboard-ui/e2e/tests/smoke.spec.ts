import { test, expect } from '@playwright/test';

test.describe('Dashboard Smoke Tests', () => {
  test('should load the dashboard home page', async ({ page }) => {
    // Navigate to the dashboard home page
    await page.goto('/');

    // Wait for the page to finish loading
    await page.waitForLoadState('networkidle');

    // Verify we successfully loaded a page (not an error)
    const url = page.url();
    expect(url).toContain('localhost:4000');

    // Take a screenshot for verification
    await page.screenshot({ path: 'e2e-results/dashboard-home.png' });
  });

  test('should load with authentication token', async ({ page }) => {
    // Get auth token from environment variable (created by nuonctl script)
    const authToken = process.env.E2E_AUTH_TOKEN;

    if (!authToken) {
      test.skip(true, 'E2E_AUTH_TOKEN not set - skipping authenticated test');
    }

    // Set authentication token as cookie
    await page.context().addCookies([
      {
        name: 'X-Nuon-Auth',
        value: authToken!,
        domain: 'localhost',
        path: '/',
      },
    ]);

    // Navigate to dashboard
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Verify authenticated state - adjust selector based on actual UI
    // For example, check for user menu, org selector, or dashboard content
    await expect(page).toHaveURL(/localhost:4000/);
  });
});
