import { test, expect } from "../fixtures";
import { waitForComponentBuildActive } from "../helpers";

test.describe("Deploy component", () => {
  test.setTimeout(120000);

  test("deploy a build and redirect to its workflow", async ({
    page,
    orgId,
    installIds,
  }) => {
    const installId = installIds[0];
    test.skip(!installId, "No seed install available");

    const built = await waitForComponentBuildActive();
    test.skip(!built, "No active build available to deploy");

    await page.goto(`/${orgId}/installs/${installId}/components`);
    await page.waitForLoadState("domcontentloaded");
    await expect(page).toHaveTitle(/^Components \|/, { timeout: 15000 });

    const rowMenu = page
      .locator('[id^="dropdown-button-component-quick-"]')
      .first();
    await expect(rowMenu).toBeVisible({ timeout: 15000 });
    await rowMenu.click();

    await page.getByRole("button", { name: "Deploy component" }).click();

    const dialog = page.getByRole("dialog");
    await expect(
      dialog.getByRole("button", { name: "Deploy build" }),
    ).toBeVisible({ timeout: 10000 });

    const buildRadio = dialog
      .locator('input[name="build-selection"]:not([disabled])')
      .first();
    await expect(buildRadio).toBeVisible({ timeout: 20000 });
    await buildRadio.check({ force: true });

    const deployButton = dialog.getByRole("button", { name: "Deploy build" });
    await expect(deployButton).toBeEnabled({ timeout: 10000 });
    await deployButton.click();

    await expect(page).toHaveURL(/\/workflows(\/|$)/, { timeout: 30000 });
  });
});
