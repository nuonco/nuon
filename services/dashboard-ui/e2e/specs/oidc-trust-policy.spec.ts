import { test, expect } from "../fixtures";

test.describe("OIDC trust policy", () => {
  test("create from the GitHub Actions preset, then edit it", async ({
    page,
    orgId,
  }) => {
    test.setTimeout(60000);
    await page.goto(`/${orgId}/settings/oidc`);
    await page.waitForLoadState("domcontentloaded");

    // --- Create ---
    await page
      .getByRole("button", { name: "Create trust policy" })
      .first()
      .click();

    const createDialog = page.getByRole("dialog");
    await expect(createDialog.getByLabel("Name")).toBeVisible({
      timeout: 10000,
    });

    const name = `e2e-oidc-${Date.now()}`;
    await createDialog.getByLabel("Name").fill(name);
    // The `sub` claim value input (the preset leaves it empty) — targeted by placeholder.
    await createDialog
      .getByPlaceholder("acme/app:main")
      .fill("repo:acme/app:ref:refs/heads/main");

    await createDialog
      .getByRole("button", { name: "Create trust policy" })
      .click();

    await expect(page.getByText("Trust policy created")).toBeVisible({
      timeout: 15000,
    });

    // --- Edit the policy we just created ---
    const row = page.getByRole("row", { name: new RegExp(name) });
    await expect(row).toBeVisible({ timeout: 10000 });
    await row.getByRole("button", { name: "Edit" }).click();

    const editDialog = page.getByRole("dialog");
    const save = editDialog.getByRole("button", { name: "Save changes" });
    await expect(save).toBeVisible({ timeout: 10000 });
    await save.click();

    await expect(page.getByText("Trust policy updated")).toBeVisible({
      timeout: 15000,
    });
  });
});
