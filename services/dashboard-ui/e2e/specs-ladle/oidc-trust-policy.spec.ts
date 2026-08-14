import { expect, test } from "@playwright/test";

const CREATE_STORY =
  "/?story=oidctrustpolicies--oidctrustpolicyform--create&mode=preview";
const EDIT_STORY =
  "/?story=oidctrustpolicies--oidctrustpolicyform--edit&mode=preview";

test.describe("OIDCTrustPolicyForm create behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(CREATE_STORY, { waitUntil: "domcontentloaded" });
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
      .getByPlaceholder("repo:acme/app:ref:refs/heads/main")
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

test.describe("OIDCTrustPolicyForm edit behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(EDIT_STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("edit opens valid and shows the enabled toggle; clearing name disables save", async ({
    page,
  }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Save changes" });

    await expect(dialog.getByText("Enabled")).toBeVisible();
    await expect(submit).toBeEnabled();

    await dialog.getByLabel("Name", { exact: true }).fill("");
    await expect(submit).toBeDisabled();

    await dialog.getByLabel("Name", { exact: true }).fill("GitHub Actions CI");
    await expect(submit).toBeEnabled();
  });
});
