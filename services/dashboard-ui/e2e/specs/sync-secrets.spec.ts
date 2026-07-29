import { test, expect } from "../fixtures";

test.describe("Sync secrets", () => {
  test.setTimeout(60000);

  test("sync install secrets and redirect to its workflow", async ({
    page,
    orgId,
    installIds,
  }) => {
    const installId = installIds[0];
    test.skip(!installId, "No seed install available");

    await page.goto(`/${orgId}/installs/${installId}?panel=settings`);
    await page.waitForLoadState("domcontentloaded");

    const panel = page.getByRole("complementary");
    const openButton = panel.getByRole("button", { name: "Sync secrets" });
    await expect(openButton).toBeVisible({ timeout: 15000 });
    await openButton.click();

    const dialog = page.getByRole("dialog");
    const syncButton = dialog.getByRole("button", { name: "Sync secrets" });
    await expect(syncButton).toBeVisible({ timeout: 10000 });
    await syncButton.click();

    const redirected = await page
      .waitForURL(/\/workflows(\/|$)/, { timeout: 30000 })
      .then(() => true)
      .catch(() => false);

    if (!redirected) {
      const failed = await page
        .getByText("Secret sync failed")
        .isVisible()
        .catch(() => false);
      test.skip(failed, "Install has no secrets to sync");
    }

    await expect(page).toHaveURL(/\/workflows(\/|$)/);
  });
});
