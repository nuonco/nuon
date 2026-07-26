import { test, expect } from "../fixtures";

const pages = [
  { path: "", title: /^Dashboard \|/, name: "dashboard" },
  { path: "apps", title: /^Apps \|/ },
  { path: "installs", title: /^Installs \|/ },
  { path: "runner", title: /^Builds \|/ },
  { path: "team", title: /^Team \|/ },
  { path: "webhooks", title: /^Webhooks \|/ },
  { path: "api-tokens", title: /^API tokens \|/ },
  { path: "slack", title: /^Slack \|/ },
];

test.describe("Navigation", () => {
  for (const { path, title, name } of pages) {
    test(`${name ?? path} page renders`, async ({ page, orgId }) => {
      await page.goto(`/${orgId}/${path}`);
      await page.waitForLoadState("domcontentloaded");
      await expect(page).toHaveTitle(title, { timeout: 15000 });
    });
  }
});
