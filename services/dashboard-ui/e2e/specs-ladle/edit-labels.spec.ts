import { expect, test } from "@playwright/test";

// Labels prefill valid, so it opens valid. Clearing a required key disables
// Save and errors on touch; refilling recovers.
const STORY = "/?story=installs--editlabels--default&mode=preview";

test.describe("EditLabels form behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("save disables when a label key is cleared", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Save labels" });

    await expect(submit).toBeEnabled();

    await dialog.getByPlaceholder("e.g. env").first().fill("");

    await expect(submit).toBeDisabled();
  });

  test("a label key errors only after it is touched", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const key = dialog.getByPlaceholder("e.g. env").first();

    await expect(dialog.getByText("Key is required")).toBeHidden();

    await key.fill("");
    await key.blur();
    await expect(dialog.getByText("Key is required")).toBeVisible();

    await key.fill("env");
    await expect(dialog.getByText("Key is required")).toBeHidden();
  });
});
