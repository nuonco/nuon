import { expect, test } from "@playwright/test";

// Rename prefills the current name, so Save is disabled until it changes AND is
// non-empty; clearing it errors on touch.
const STORY = "/?story=serviceaccounts--renameserviceaccount--default&mode=preview";

test.describe("RenameServiceAccount form behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("save is disabled until the name changes", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Save" });

    await expect(submit).toBeDisabled();

    await dialog.getByLabel("Name", { exact: true }).fill("ci-deploy-2");

    await expect(submit).toBeEnabled();
  });

  test("the name errors only after it is touched", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const name = dialog.getByLabel("Name", { exact: true });

    await expect(dialog.getByText("Name is required")).toBeHidden();

    await name.fill("");
    await name.blur();
    await expect(dialog.getByText("Name is required")).toBeVisible();

    await name.fill("ci-deploy-2");
    await expect(dialog.getByText("Name is required")).toBeHidden();
  });
});
