import { test, expect } from "../fixtures";

test.describe("Forget install", () => {
  test.setTimeout(60000);

  test("forget the second seed install", async ({ page, orgId, installIds }) => {
    const installId = installIds[1];
    test.skip(!installId, "No second seed install available");

    await page.goto(`/${orgId}/installs/${installId}?panel=settings`);
    await page.waitForLoadState("domcontentloaded");

    const panel = page.getByRole("complementary");
    const openButton = panel.getByRole("button", { name: "Forget install" });
    await expect(openButton).toBeVisible({ timeout: 15000 });
    await openButton.click();

    const dialog = page.getByRole("dialog");
    const headingText =
      (await dialog.getByText(/^Forget .+\?$/).textContent()) ?? "";
    const installName = headingText
      .replace(/^Forget\s+/, "")
      .replace(/\?\s*$/, "")
      .trim();
    expect(installName.length).toBeGreaterThan(0);

    await dialog.locator("#confirm-install-name").fill(installName);
    await dialog.getByRole("button", { name: "Forget install" }).click();

    await expect(page).toHaveURL(/\/installs$/, { timeout: 15000 });
  });
});
