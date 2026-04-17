import { test, expect } from "../fixtures";

test("dashboard loads and shows logo", async ({ page, orgId }) => {
  await page.goto(`/${orgId}`);
  await page.waitForLoadState("networkidle");

  const logo = page.locator(".logo-link");
  await expect(logo).toBeVisible();
  await expect(logo.locator("span.sr-only")).toHaveText("Nuon");
});
