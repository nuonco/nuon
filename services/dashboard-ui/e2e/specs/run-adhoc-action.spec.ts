import { test, expect } from "../fixtures";

test.describe("Run adhoc action", () => {
  test.setTimeout(60000);

  test("open the adhoc action form, fill a command, and run", async ({
    page,
    orgId,
    installIds,
  }) => {
    const installId = installIds[0];
    test.skip(!installId, "No seed install available");

    await page.goto(`/${orgId}/installs/${installId}/actions`);
    await page.waitForLoadState("domcontentloaded");
    await expect(page).toHaveTitle(/^Actions \|/, { timeout: 15000 });

    await page.getByRole("button", { name: "Run adhoc action" }).click();

    // A persisted draft surfaces a "Resume draft" modal first — start fresh if so.
    const startFresh = page.getByRole("button", { name: "Start fresh" });
    if (await startFresh.isVisible({ timeout: 2000 }).catch(() => false)) {
      await startFresh.click();
    }

    const dialog = page.getByRole("dialog");
    await expect(dialog.getByLabel("Command")).toBeVisible({ timeout: 10000 });
    await dialog.getByLabel("Command").fill("echo hello from e2e");

    await dialog.getByRole("button", { name: "Run action" }).click();

    await expect(page).toHaveURL(/\/workflows\//, { timeout: 30000 });
  });
});
