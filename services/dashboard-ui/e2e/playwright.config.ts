import { defineConfig, devices } from "@playwright/test";
import { env } from "./env";

export default defineConfig({
  testDir: "./specs",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 1 : undefined,

  outputDir: "./.results",
  reporter: process.env.CI ? "github" : [["html", { outputFolder: "./.report" }]],

  globalSetup: "./global-setup.ts",
  globalTeardown: "./global-teardown.ts",

  use: {
    baseURL: env.baseUrl,
    storageState: "e2e/.auth/user.json",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },

  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
