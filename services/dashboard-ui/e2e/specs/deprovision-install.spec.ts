import { test, expect } from "../fixtures";
import { createThrowawayInstall } from "../helpers";

test.describe("Deprovision install", () => {
  test.setTimeout(120000);

  test("deprovision an install and redirect to its workflow", async ({
    page,
    orgId,
  }) => {
    const install = await createThrowawayInstall("e2e-deprovision");
    test.skip(!install, "Could not create throwaway install");

    await page.goto(`/${orgId}/installs/${install!.id}?panel=settings`);
    await page.waitForLoadState("domcontentloaded");

    const panel = page.getByRole("complementary");
    const openButton = panel.getByRole("button", {
      name: "Deprovision install",
    });
    await expect(openButton).toBeVisible({ timeout: 15000 });
    await openButton.click();

    const dialog = page.getByRole("dialog");
    await expect(
      dialog.getByText("Deprovision install?", { exact: true }),
    ).toBeVisible({ timeout: 10000 });

    await dialog.locator("#confirm-install-name").fill(install!.name);

    const deprovisionButton = dialog.getByRole("button", {
      name: "Deprovision install",
    });
    await expect(deprovisionButton).toBeEnabled({ timeout: 10000 });
    await deprovisionButton.click();

    await expect(page).toHaveURL(/\/workflows(\/|$)/, { timeout: 30000 });
  });
});
