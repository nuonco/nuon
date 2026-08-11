import { expect, test } from "@playwright/test";

const STORY = "/?story=runners--updaterunner--default&mode=preview";

test.describe("UpdateRunner form behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("submit stays disabled until the form is valid", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", {
      name: "Update runner version",
    });

    await expect(submit).toBeDisabled();

    await dialog.getByPlaceholder("runner tag").fill("v1.2.3");

    await expect(submit).toBeEnabled();
  });

  test("a required field errors only after it is touched", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const tag = dialog.getByPlaceholder("runner tag");

    await expect(dialog.getByText("Enter a value to update to")).toBeHidden();

    await tag.click();
    await tag.blur();
    await expect(dialog.getByText("Enter a value to update to")).toBeVisible();

    await tag.fill("v1.2.3");
    await expect(dialog.getByText("Enter a value to update to")).toBeHidden();
  });
});
