import { test, expect } from "../fixtures";

const pages = [
  { path: "apps", title: /^Apps \|/ },
  { path: "installs", title: /^Installs \|/ },
  { path: "runner", title: /^Builds \|/ },
  { path: "team", title: /^Team \|/ },
];

test.describe("Navigation", () => {
  for (const { path, title } of pages) {
    test(`${path} page renders`, async ({ page, orgId }) => {
      await page.goto(`/${orgId}/${path}`);
      await page.waitForLoadState("domcontentloaded");
      await expect(page).toHaveTitle(title, { timeout: 15000 });
    });
  }
});
