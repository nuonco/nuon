import { test, expect } from "../fixtures";

test.describe("Create trigger", () => {
  test("open modal, fill name, submit, redirect to detail", async ({
    page,
    orgId,
  }) => {
    test.setTimeout(60000);
    await page.goto(`/${orgId}/settings/triggers`);
    await page.waitForLoadState("domcontentloaded");

    await page.getByRole("button", { name: "Create trigger" }).first().click();

    const dialog = page.getByRole("dialog");
    await expect(dialog.getByLabel("Name")).toBeVisible({ timeout: 10000 });

    const name = `e2e-trigger-${Date.now()}`;
    await dialog.getByLabel("Name").fill(name);
    await dialog.getByRole("button", { name: "Create trigger" }).click();

    await expect(page).toHaveURL(/\/settings\/triggers\/.+/, { timeout: 15000 });
  });
});
