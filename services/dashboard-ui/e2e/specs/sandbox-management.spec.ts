import { test, expect } from "../fixtures";
import { createThrowawayInstall } from "../helpers";

test.describe("Sandbox management", () => {
  test.setTimeout(240000);

  test("drift scan, reprovision, and deprovision the sandbox", async ({
    page,
    orgId,
  }) => {
    const install = await createThrowawayInstall("e2e-sandbox");
    test.skip(!install, "Could not create throwaway install");

    const sandboxUrl = `/${orgId}/installs/${install!.id}/sandbox`;

    const openSandbox = async () => {
      await page.goto(sandboxUrl);
      await page.waitForLoadState("domcontentloaded");
      await expect(page).toHaveTitle(/^Sandbox \|/, { timeout: 15000 });
      await page.getByRole("button", { name: "Sandbox controls" }).click();
    };

    const runAction = async (menuItem: string, confirm: () => Promise<void>) => {
      await openSandbox();
      await page.getByRole("button", { name: menuItem }).click();
      await expect(page.getByRole("dialog")).toBeVisible({ timeout: 10000 });
      await confirm();
      await expect(page).toHaveURL(/\/workflows(\/|$)/, { timeout: 30000 });
    };

    await runAction("Drift scan sandbox", async () => {
      await page
        .getByRole("dialog")
        .getByRole("button", { name: "Drift scan sandbox" })
        .click();
    });

    await runAction("Reprovision sandbox", async () => {
      await page
        .getByRole("dialog")
        .getByRole("button", { name: "Reprovision sandbox" })
        .click();
    });

    await runAction("Deprovision sandbox", async () => {
      const dialog = page.getByRole("dialog");
      await dialog.getByPlaceholder("deprovision").fill("deprovision");
      const button = dialog.getByRole("button", { name: "Deprovision sandbox" });
      await expect(button).toBeEnabled({ timeout: 10000 });
      await button.click();
    });
  });
});
