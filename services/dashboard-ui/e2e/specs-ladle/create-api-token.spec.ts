import { expect, test } from "@playwright/test";

// Behavior-test template for TanStack Form + Zod forms: drives the Ladle story
// (no backend) and asserts the interaction contract the wrappers guarantee —
// disabled-submit-until-valid and error-on-touch. New migrated forms get a
// sibling spec following this shape.
const STORY = "/?story=apitokens--createapitoken--default&mode=preview";

test.describe("CreateApiToken form behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("submit stays disabled until the form is valid", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Create token" });

    await expect(submit).toBeDisabled();

    await dialog.getByLabel("Name").fill("ci-deploy");

    await expect(submit).toBeEnabled();
  });

  test("a required field errors only after it is touched", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const name = dialog.getByLabel("Name");

    // Error is not shown on open (untouched), even though the field is invalid.
    await expect(dialog.getByText("Name is required")).toBeHidden();

    await name.click();
    await name.blur();
    await expect(dialog.getByText("Name is required")).toBeVisible();

    await name.fill("ci-deploy");
    await expect(dialog.getByText("Name is required")).toBeHidden();
  });
});
