import { expect, test } from "@playwright/test";

// RunRunbook is a multi-page wizard. On the inputs page, Next is gated on the
// inputs being valid (required inputs filled); the required input errors on touch.
const STORY = "/?story=runbooks--runrunbook--with-inputs&mode=preview";

test.describe("RunRunbook wizard behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("Next stays disabled until required inputs are valid", async ({
    page,
  }) => {
    const dialog = page.getByRole("dialog");
    const next = dialog.getByRole("button", { name: "Next" });

    await expect(next).toBeDisabled();

    await dialog.getByLabel("Target *").fill("prod");

    await expect(next).toBeEnabled();
  });

  test("a required input errors only after it is touched", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const target = dialog.getByLabel("Target *");

    await expect(dialog.getByText("Target is required")).toBeHidden();

    await target.click();
    await target.blur();
    await expect(dialog.getByText("Target is required")).toBeVisible();

    await target.fill("prod");
    await expect(dialog.getByText("Target is required")).toBeHidden();
  });
});
