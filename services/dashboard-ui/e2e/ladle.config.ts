import { defineConfig, devices } from "@playwright/test";

const ladleUrl = process.env.LADLE_URL ?? "http://localhost:61000";

export default defineConfig({
  testDir: "./specs-ladle",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 1 : undefined,

  outputDir: "./.results-ladle",
  reporter: process.env.CI
    ? "github"
    : [["html", { outputFolder: "./.report-ladle" }]],

  use: {
    baseURL: ladleUrl,
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },

  webServer: {
    command: "bun run dev:ladle",
    url: ladleUrl,
    reuseExistingServer: true,
    timeout: 120_000,
  },

  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
