import { expect, test } from "@playwright/test";

// This form prefills from the action's config env vars, so it opens valid.
// The meaningful contract here: clearing a required config var disables submit
// and errors on touch; refilling recovers.
const STORY =
  "/?story=actions--installactionmanualrun--default&mode=preview";

test.describe("InstallActionManualRun form behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("submit disables when a required env var is cleared", async ({
    page,
  }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Run action" });

    await expect(submit).toBeEnabled();

    await dialog.getByLabel("API_URL").fill("");

    await expect(submit).toBeDisabled();
  });

  test("a required env var errors only after it is touched", async ({
    page,
  }) => {
    const dialog = page.getByRole("dialog");
    const apiUrl = dialog.getByLabel("API_URL");

    await expect(dialog.getByText("Required")).toBeHidden();

    await apiUrl.fill("");
    await apiUrl.blur();
    await expect(dialog.getByText("Required")).toBeVisible();

    await apiUrl.fill("https://api.example.com");
    await expect(dialog.getByText("Required")).toBeHidden();
  });
});
