import { expect, test } from "@playwright/test";

// Edit prefills the branch name, so it opens valid. Clearing the required name
// disables Save and errors on touch; refilling recovers.
const STORY = "/?story=branches--editbranchnamemodal--default&mode=preview";

test.describe("EditBranchName form behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("save disables when the branch name is cleared", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Save changes" });

    await expect(submit).toBeEnabled();

    await dialog.getByLabel("Branch name").fill("");

    await expect(submit).toBeDisabled();
  });

  test("the branch name errors only after it is touched", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const name = dialog.getByLabel("Branch name");

    await expect(dialog.getByText("Branch name cannot be empty")).toBeHidden();

    await name.fill("");
    await name.blur();
    await expect(dialog.getByText("Branch name cannot be empty")).toBeVisible();

    await name.fill("production");
    await expect(dialog.getByText("Branch name cannot be empty")).toBeHidden();
  });
});
