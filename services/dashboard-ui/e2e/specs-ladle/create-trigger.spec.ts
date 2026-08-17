import { expect, test } from "@playwright/test";

const STORY = "/?story=triggers--create-trigger--default&mode=preview";

test.describe("CreateTrigger form behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("submit stays disabled until the form is valid", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Create trigger" });

    await expect(submit).toBeDisabled();

    await dialog.getByLabel("Name", { exact: true }).fill("gh-events");

    await expect(submit).toBeEnabled();
  });

  test("a required field errors only after it is touched", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const name = dialog.getByLabel("Name", { exact: true });

    await expect(dialog.getByText("Name is required")).toBeHidden();

    await name.click();
    await name.blur();
    await expect(dialog.getByText("Name is required")).toBeVisible();

    await name.fill("gh-events");
    await expect(dialog.getByText("Name is required")).toBeHidden();
  });
});
