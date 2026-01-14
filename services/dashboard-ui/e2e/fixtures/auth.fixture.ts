import { test as base, expect } from '@playwright/test';

/**
 * Authentication fixture for Playwright E2E tests
 *
 * Provides an authenticated page context with the X-Nuon-Auth cookie already set.
 * This eliminates the need to manually set authentication in every test.
 *
 * Usage:
 * ```typescript
 * import { test, expect } from '../fixtures/auth.fixture';
 *
 * test('my test', async ({ authenticatedPage }) => {
 *   await authenticatedPage.goto('/installs');
 *   // Page is already authenticated!
 * });
 * ```
 *
 * Requires E2E_AUTH_TOKEN environment variable to be set.
 */
export const test = base.extend({
  authenticatedPage: async ({ page }, use) => {
    // Get auth token from environment
    const authToken = process.env.E2E_AUTH_TOKEN;

    if (!authToken) {
      throw new Error(
        'E2E_AUTH_TOKEN environment variable not set. ' +
        'Set it with: export E2E_AUTH_TOKEN="your-token-here"'
      );
    }

    // Add authentication cookie
    await page.context().addCookies([
      {
        name: 'X-Nuon-Auth',
        value: authToken,
        domain: 'localhost',
        path: '/',
      },
    ]);

    // Provide the authenticated page to the test
    await use(page);
  },
});

export { expect };
