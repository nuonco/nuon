import { test, expect } from "../fixtures";

const subPages = [
  { path: "", label: "overview", title: /^Overview \|/ },
  {
    path: "/inputs",
    label: "current inputs",
    title: /^Current inputs \|/,
    visibleText: "The current input values for this install.",
  },
  { path: "/state", label: "view state", title: /^State \|/ },
  { path: "/components", label: "components", title: /^Components \|/ },
  { path: "/actions", label: "actions", title: /^Actions \|/ },
];

test.describe("Install navigation", () => {
  test.setTimeout(60000);

  for (const { path, label, title, visibleText } of subPages) {
    test(`${label} page renders`, async ({ page, orgId, installIds }) => {
      const installId = installIds[0];
      test.skip(!installId, "No seed install available");

      await page.goto(`/${orgId}/installs/${installId}${path}`);
      await page.waitForLoadState("domcontentloaded");

      await expect(page).toHaveTitle(title, { timeout: 15000 });
      await expect(page.getByText("Something went wrong.")).not.toBeVisible();

      if (visibleText) {
        await expect(page.getByText(visibleText)).toBeVisible({
          timeout: 15000,
        });
      }
    });
  }
});
