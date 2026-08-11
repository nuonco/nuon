import { expect, test } from "@playwright/test";

// Edit prefills a valid policy, so it opens valid. Clearing the required name
// disables Save and errors on touch; refilling recovers.
const STORY =
  "/?story=oidctrustpolicies--editoidctrustpolicy--default&mode=preview";

test.describe("EditOIDCTrustPolicy form behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("save disables when the name is cleared", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Save changes" });

    await expect(submit).toBeEnabled();

    await dialog.getByLabel("Name", { exact: true }).fill("");

    await expect(submit).toBeDisabled();
  });

  test("the name field errors only after it is touched", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const name = dialog.getByLabel("Name", { exact: true });

    await expect(dialog.getByText("Name is required")).toBeHidden();

    await name.fill("");
    await name.blur();
    await expect(dialog.getByText("Name is required")).toBeVisible();

    await name.fill("GitHub Actions CI");
    await expect(dialog.getByText("Name is required")).toBeHidden();
  });
});
