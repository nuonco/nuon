import { expect, test } from "@playwright/test";

const STORY =
  "/?story=serviceaccounts--createserviceaccount--default&mode=preview";

test.describe("CreateServiceAccount form behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("submit stays disabled until the form is valid", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", {
      name: "Create service account",
    });

    await expect(submit).toBeDisabled();

    await dialog.getByLabel("Name").fill("ci-deploy");

    await expect(submit).toBeEnabled();
  });

  test("a required field errors only after it is touched", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const name = dialog.getByLabel("Name");

    await expect(dialog.getByText("Name is required")).toBeHidden();

    await name.click();
    await name.blur();
    await expect(dialog.getByText("Name is required")).toBeVisible();

    await name.fill("ci-deploy");
    await expect(dialog.getByText("Name is required")).toBeHidden();
  });
});
