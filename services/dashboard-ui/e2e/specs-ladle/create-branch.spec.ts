import { expect, test } from "@playwright/test";

const STORY = "/?story=branches--createbranchmodal--default&mode=preview";

test.describe("CreateBranch form behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("submit stays disabled until the branch name is valid", async ({
    page,
  }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Create branch" });

    await expect(submit).toBeDisabled();

    await dialog.getByLabel("Branch name").fill("production");

    await expect(submit).toBeEnabled();
  });

  test("the branch name errors only after it is touched", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const name = dialog.getByLabel("Branch name");

    await expect(dialog.getByText("Branch name is required")).toBeHidden();

    await name.click();
    await name.blur();
    await expect(dialog.getByText("Branch name is required")).toBeVisible();

    await name.fill("production");
    await expect(dialog.getByText("Branch name is required")).toBeHidden();
  });
});
