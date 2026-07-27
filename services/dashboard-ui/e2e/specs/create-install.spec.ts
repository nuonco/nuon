import { test, expect } from "../fixtures";

test.describe("Create install", () => {
  test("navigate to installs page", async ({ page, orgId }) => {
    await page.goto(`/${orgId}/installs`);
    await page.waitForLoadState("domcontentloaded");
    await expect(page.getByRole("heading", { name: "Installs" })).toBeVisible();
  });

  test("open modal, select app, fill form, and submit", async ({
    page,
    orgId,
  }) => {
    test.setTimeout(60000);
    await page.goto(`/${orgId}/installs`);
    await page.waitForLoadState("domcontentloaded");

    // Open create modal (first match is the page header button)
    await page.getByRole("button", { name: "Create install" }).first().click();
    await expect(
      page.getByText("Select an app to create an install")
    ).toBeVisible();

    // App list loads with search
    await expect(
      page.getByPlaceholder("Search apps...")
    ).toBeVisible();

    // Select first app
    const firstRadio = page.locator('input[name="app-selection"]').first();
    await firstRadio.click({ force: true });

    // Form loads — wait for the install name input to appear
    const nameInput = page.getByPlaceholder("Enter install name");
    await expect(nameInput).toBeVisible({ timeout: 10000 });
    await expect(
      page.getByText("Select an app to create an install")
    ).not.toBeVisible();

    // Fill install name
    const installName = `e2e-test-${Date.now()}`;
    await nameInput.fill(installName);

    // Select AWS region (if the region selector is present)
    const awsSettings = page.getByRole("group", { name: /AWS settings/i });
    if (await awsSettings.isVisible({ timeout: 8000 }).catch(() => false)) {
      await awsSettings.getByRole("combobox").click();
      const searchInput = page.getByPlaceholder("Search...");
      await expect(searchInput).toBeVisible();
      await searchInput.fill("us-west-2");
      await page.getByRole("option", { name: /us-west-2/ }).first().click();
    }

    // Enable auto-approve
    const autoApprove = page.getByText("Auto-approve changes");
    await autoApprove.click();

    // Submit the form
    await page.getByRole("button", { name: "Create install" }).last().click();

    // Redirected to provision workflow
    await expect(page).toHaveURL(/\/workflows\//, { timeout: 30000 });
  });
});
