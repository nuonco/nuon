import { expect, test } from "@playwright/test";

// Create needs a name AND a sub claim-condition value (issuer/audience prefill
// from the github preset). The repo picker autofills both, but here we fill
// them manually to exercise the validation directly.
const STORY =
  "/?story=oidctrustpolicies--createoidctrustpolicy--default&mode=preview";

test.describe("CreateOIDCTrustPolicy form behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("submit enables only once name and sub condition are set", async ({
    page,
  }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Create trust policy" });

    await expect(submit).toBeDisabled();

    await dialog.getByLabel("Name", { exact: true }).fill("ci-deploy");
    await expect(submit).toBeDisabled();

    await dialog
      .getByPlaceholder("acme/app:main")
      .fill("repo:acme/app:ref:refs/heads/main");
    await expect(submit).toBeEnabled();
  });

  test("the name field errors only after it is touched", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const name = dialog.getByLabel("Name", { exact: true });

    await expect(dialog.getByText("Name is required")).toBeHidden();

    await name.click();
    await name.blur();
    await expect(dialog.getByText("Name is required")).toBeVisible();

    await name.fill("ci-deploy");
    await expect(dialog.getByText("Name is required")).toBeHidden();
  });
});
