import { test, expect } from "../fixtures";

test.describe("API tokens", () => {
  test.setTimeout(60000);

  test("create a token, confirm the reveal modal, then delete it", async ({
    page,
    orgId,
  }) => {
    const tokenName = `e2e-token-${Date.now()}`;

    await page.goto(`/${orgId}/api-tokens`);
    await page.waitForLoadState("domcontentloaded");
    await expect(page).toHaveTitle(/^API tokens \|/, { timeout: 15000 });

    // --- Create (default role + expiry) ---
    await page.getByRole("button", { name: "Create token" }).first().click();
    const createDialog = page.getByRole("dialog");
    await expect(createDialog.getByText("Create API token")).toBeVisible();
    await page.getByPlaceholder("e.g. ci-deploy").fill(tokenName);
    await createDialog.getByRole("button", { name: "Create token" }).click();

    // --- One-time reveal modal ---
    const revealDialog = page.getByRole("dialog");
    await expect(revealDialog.getByText("API token created")).toBeVisible({
      timeout: 10000,
    });
    await revealDialog.getByRole("button", { name: "Done" }).click();

    // --- Token appears in the table ---
    const row = page.getByRole("row").filter({ hasText: tokenName });
    await expect(row).toBeVisible({ timeout: 10000 });

    // --- Delete via the row menu ---
    await row.getByRole("button").last().click();
    await page.getByRole("button", { name: "Delete token" }).click();
    const deleteDialog = page.getByRole("dialog");
    await expect(deleteDialog.getByText("Delete API token?")).toBeVisible();
    await deleteDialog.getByRole("button", { name: "Delete token" }).click();
    await expect(page.getByText("Token deleted")).toBeVisible({
      timeout: 10000,
    });
    await expect(row).not.toBeVisible();
  });
});
