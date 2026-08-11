import { expect, test } from "@playwright/test";

// CreateOrgStep is an inline onboarding wizard step, not a modal — the form
// renders directly on the page (no "Open modal" trigger, no dialog).
const STORY = "/?story=onboarding--v1-steps--createorgstep--default&mode=preview";

test.describe("CreateOrgStep form behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await expect(page.getByLabel("Organization name")).toBeVisible();
  });

  test("submit stays disabled until the form is valid", async ({ page }) => {
    const submit = page.getByRole("button", { name: "Create organization" });

    await expect(submit).toBeDisabled();

    await page.getByLabel("Organization name").fill("my-org");

    await expect(submit).toBeEnabled();
  });

  test("a required field errors only after it is touched", async ({ page }) => {
    const name = page.getByLabel("Organization name");

    await expect(
      page.getByText("Organization name is required")
    ).toBeHidden();

    await name.click();
    await name.blur();
    await expect(
      page.getByText("Organization name is required")
    ).toBeVisible();

    await name.fill("my-org");
    await expect(
      page.getByText("Organization name is required")
    ).toBeHidden();
  });
});
