import { expect, test } from "@playwright/test";

const STORY = "/?story=triggers--revoketriggersecretmodal--default&mode=preview";

test.describe("RevokeTriggerSecretModal behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("warns before revoking and confirms with a danger action", async ({
    page,
  }) => {
    const dialog = page.getByRole("dialog");

    await expect(dialog.getByText("Revoke secret?")).toBeVisible();
    await expect(
      dialog.getByText("stop authenticating immediately", { exact: false })
    ).toBeVisible();

    const revoke = dialog.getByRole("button", { name: "Revoke secret" });
    await expect(revoke).toBeVisible();
    await revoke.click();

    await expect(page.getByRole("dialog")).toBeHidden();
  });
});
