import { expect, test } from "@playwright/test";

const STORY = "/?story=sandbox--deprovisionsandbox--default&mode=preview";

test.describe("DeprovisionSandbox behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("deprovision is gated on typing the install name", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Deprovision sandbox" });
    const confirm = dialog.getByPlaceholder("install name");

    await expect(submit).toBeDisabled();

    await confirm.fill("deprovision");
    await expect(dialog.getByText("Install name doesn't match")).toBeVisible();
    await expect(submit).toBeDisabled();

    await confirm.fill("prod-acme");
    await expect(dialog.getByText("Install name doesn't match")).toBeHidden();
    await expect(submit).toBeEnabled();
  });
});
