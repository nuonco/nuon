import { expect, test } from "@playwright/test";

const STORY = "/?story=installs--editinputs--with-name-field&mode=preview";

test.describe("EditInstall (edit mode) behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("edit prefills valid and re-validates on clear", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Update install" });
    const name = dialog.getByPlaceholder("Enter install name");

    await expect(name).toHaveValue("prod-acme");
    await expect(submit).toBeEnabled();
    await expect(dialog.getByText("Install name is required")).toBeHidden();

    await name.fill("");
    await name.blur();
    await expect(submit).toBeDisabled();
    await expect(dialog.getByText("Install name is required")).toBeVisible();

    await name.fill("prod-acme");
    await expect(submit).toBeEnabled();
    await expect(dialog.getByText("Install name is required")).toBeHidden();
  });
});
