import { test, expect } from "../fixtures";

test("apps page shows the seeded app", async ({ page, orgId, appConfig }) => {
  test.skip(!appConfig, "no app config seeded");

  await page.goto(`/${orgId}/apps`);

  const row = page.locator("table tbody tr").filter({ hasText: appConfig! });
  await expect(row).toBeVisible();
});
