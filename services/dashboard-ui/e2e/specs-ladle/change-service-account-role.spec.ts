import { expect, test } from "@playwright/test";

// Change-role opens on the current role, so Save is disabled until a different
// role is picked.
const STORY =
  "/?story=serviceaccounts--changeserviceaccountrole--default&mode=preview";

test.describe("ChangeServiceAccountRole form behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("save is disabled until a different role is selected", async ({
    page,
  }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Save" });

    await expect(submit).toBeDisabled();

    await dialog.getByText("Runner", { exact: true }).click();
    await page.getByText("Admin", { exact: true }).click();

    await expect(submit).toBeEnabled();
  });
});
