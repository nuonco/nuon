import { test, expect } from "../fixtures";

test.describe("Run a runbook", () => {
  test.setTimeout(60000);

  test("run the verify_status runbook and redirect to its workflow", async ({
    page,
    orgId,
    installIds,
  }) => {
    const installId = installIds[0];
    test.skip(!installId, "No seed install available");

    await page.goto(`/${orgId}/installs/${installId}/runbooks`);
    await page.waitForLoadState("domcontentloaded");
    await expect(page).toHaveTitle(/^Runbooks \|/, { timeout: 15000 });

    const row = page.getByRole("row").filter({ hasText: "verify_status" });
    await expect(row).toBeVisible({ timeout: 15000 });
    await row.getByRole("button").last().click();

    await page.getByRole("button", { name: "Run runbook" }).click();

    const dialog = page.getByRole("dialog");
    await expect(dialog.getByText("Run verify_status")).toBeVisible({
      timeout: 10000,
    });
    await dialog.getByRole("button", { name: "Run runbook" }).click();

    await expect(page).toHaveURL(/\/workflows\//, { timeout: 30000 });
    await expect(page.getByText("Runbook run started")).toBeVisible({
      timeout: 10000,
    });
  });
});
