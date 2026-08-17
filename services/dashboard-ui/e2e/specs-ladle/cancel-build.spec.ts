import { expect, test } from "@playwright/test";

const STORY = "/?story=builds--cancelbuildmodal--default&mode=preview";

test.describe("CancelBuildModal behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("confirms with a danger action and a non-ambiguous dismiss", async ({
    page,
  }) => {
    const dialog = page.getByRole("dialog");

    await expect(dialog.getByText("Cancel build?")).toBeVisible();
    await expect(
      dialog.getByRole("button", { name: "Cancel build" })
    ).toBeVisible();

    const keepBuilding = dialog.getByRole("button", { name: "Keep building" });
    await expect(keepBuilding).toBeVisible();
    await keepBuilding.click();

    await expect(page.getByRole("dialog")).toBeHidden();
  });
});
