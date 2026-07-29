import { test, expect } from "../fixtures";
import { createThrowawayInstall } from "../helpers";

test.describe("Reprovision install", () => {
  test.setTimeout(120000);

  test("reprovision an install and redirect to its workflow", async ({
    page,
    orgId,
  }) => {
    const install = await createThrowawayInstall("e2e-reprovision");
    test.skip(!install, "Could not create throwaway install");

    await page.goto(`/${orgId}/installs/${install!.id}?panel=settings`);
    await page.waitForLoadState("domcontentloaded");

    const panel = page.getByRole("complementary");
    const openButton = panel.getByRole("button", {
      name: "Reprovision install",
    });
    await expect(openButton).toBeVisible({ timeout: 15000 });
    await openButton.click();

    const dialog = page.getByRole("dialog");
    await expect(
      dialog.getByText("Reprovision install?", { exact: true }),
    ).toBeVisible({ timeout: 10000 });

    await dialog.getByRole("button", { name: "Reprovision install" }).click();

    await expect(page).toHaveURL(/\/workflows(\/|$)/, { timeout: 30000 });
  });
});
