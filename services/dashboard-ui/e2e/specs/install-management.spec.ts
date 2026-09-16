import { test, expect } from "../fixtures";

test.describe("Install management", () => {
  test.setTimeout(60000);

  test("enable then disable auto-approval", async ({
    page,
    orgId,
    installIds,
  }) => {
    const installId = installIds[0];
    test.skip(!installId, "No seed install available");

    // /workflows is a legacy path now; landing on /history also covers the redirect.
    await page.goto(`/${orgId}/installs/${installId}/workflows`);
    await page.waitForLoadState("domcontentloaded");
    await expect(page).toHaveURL(
      new RegExp(`/installs/${installId}/history$`),
      { timeout: 15000 }
    );
    await expect(page).toHaveTitle(/^History \|/, { timeout: 15000 });

    const toggle = page.getByRole("switch", { name: "Auto approval" });
    await expect(toggle).toBeVisible({ timeout: 15000 });

    // --- Enable ---
    await toggle.click();
    const enableDialog = page.getByRole("dialog");
    await enableDialog
      .getByRole("button", { name: "Enable auto approval" })
      .click();
    await expect(
      page.getByText("Auto approve enabled", { exact: true })
    ).toBeVisible({ timeout: 10000 });

    // --- Disable (restore) ---
    await toggle.click();
    const disableDialog = page.getByRole("dialog");
    await disableDialog
      .getByRole("button", { name: "Disable auto approval" })
      .click();
    await expect(
      page.getByText("Auto approve disabled", { exact: true })
    ).toBeVisible({ timeout: 10000 });
  });
});
