import { test, expect } from "../fixtures";
import { waitForComponentBuildActive } from "../helpers";

test.describe("Drift scan component", () => {
  test.setTimeout(120000);

  test("drift scan a component and redirect to its workflow", async ({
    page,
    orgId,
    installIds,
  }) => {
    const installId = installIds[0];
    test.skip(!installId, "No seed install available");

    const built = await waitForComponentBuildActive();
    test.skip(!built, "No active build available to drift scan");

    await page.goto(`/${orgId}/installs/${installId}/components`);
    await page.waitForLoadState("domcontentloaded");
    await expect(page).toHaveTitle(/^Components \|/, { timeout: 15000 });

    const rowMenu = page
      .locator('[id^="dropdown-button-component-quick-"]')
      .first();
    await expect(rowMenu).toBeVisible({ timeout: 15000 });
    await rowMenu.click();

    await page.getByRole("button", { name: "Drift scan component" }).click();

    const dialog = page.getByRole("dialog");
    await expect(
      dialog.getByRole("button", { name: "Drift scan build" }),
    ).toBeVisible({ timeout: 10000 });

    const buildRadio = dialog
      .locator('input[name="build-selection"]:not([disabled])')
      .first();
    await expect(buildRadio).toBeVisible({ timeout: 20000 });
    await buildRadio.check({ force: true });

    const scanButton = dialog.getByRole("button", { name: "Drift scan build" });
    await expect(scanButton).toBeEnabled({ timeout: 10000 });
    await scanButton.click();

    await expect(page).toHaveURL(/\/workflows(\/|$)/, { timeout: 30000 });
  });
});
