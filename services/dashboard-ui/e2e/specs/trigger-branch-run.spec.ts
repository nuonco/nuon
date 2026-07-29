import { test, expect } from "../fixtures";
import { createTriggerableBranch } from "../helpers";

test.describe("Trigger branch run", () => {
  test.setTimeout(90000);

  test("trigger a run for a branch with a deployment plan", async ({
    page,
    orgId,
  }) => {
    const branch = await createTriggerableBranch();
    test.skip(!branch, "Could not seed a triggerable branch");

    await page.goto(
      `/${orgId}/apps/${branch!.appId}/branches/${branch!.branchId}`,
    );
    await page.waitForLoadState("domcontentloaded");

    const triggerButton = page
      .getByRole("button", { name: "Trigger run" })
      .first();
    await expect(triggerButton).toBeVisible({ timeout: 15000 });
    await expect(triggerButton).toBeEnabled({ timeout: 15000 });
    await triggerButton.click();

    const dialog = page.getByRole("dialog");
    const confirmButton = dialog.getByRole("button", { name: "Trigger run" });
    await expect(confirmButton).toBeVisible({ timeout: 10000 });
    await confirmButton.click();

    await expect(page.getByText("Run triggered", { exact: true })).toBeVisible({
      timeout: 15000,
    });
  });
});
