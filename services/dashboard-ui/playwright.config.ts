import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright E2E Test Configuration for Nuon Dashboard
 *
 * Prerequisites:
 * - Next.js dev server running on localhost:4000
 * - ctl-api service running on localhost:8081
 * - E2E_AUTH_TOKEN set in environment (created by nuonctl script)
 */
export default defineConfig({
  // Directory containing E2E test files
  testDir: './e2e/tests',

  // Run tests sequentially to avoid auth conflicts
  fullyParallel: false,

  // Fail the build on CI if you accidentally left test.only
  forbidOnly: !!process.env.CI,

  // Retry on CI only
  retries: process.env.CI ? 2 : 0,

  // Single worker to avoid conflicts
  workers: 1,

  // Reporter to use
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['list'],
  ],

  // Shared settings for all projects
  use: {
    baseURL: 'http://localhost:4000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 15000,
    navigationTimeout: 30000,
  },

  // Test timeout
  timeout: 60000,

  // Expect timeout
  expect: {
    timeout: 10000,
  },

  // Configure projects for major browsers
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 1920, height: 1080 },
      },
    },
  ],

  // Assume services already running
  webServer: undefined,
});
