import { test, expect } from "../fixtures";

test.describe("Run an action", () => {
  test.setTimeout(60000);

  test("trigger the healthcheck action and redirect to its workflow", async ({
    page,
    orgId,
    installIds,
  }) => {
    const installId = installIds[0];
    test.skip(!installId, "No seed install available");

    await page.goto(`/${orgId}/installs/${installId}/actions`);
    await page.waitForLoadState("domcontentloaded");
    await expect(page).toHaveTitle(/^Actions \|/, { timeout: 15000 });

    await page
      .getByRole("link", { name: "healthcheck", exact: true })
      .click();

    const runButton = page.getByRole("button", { name: "Run action" }).first();
    await expect(runButton).toBeVisible({ timeout: 15000 });
    await runButton.click();

    const dialog = page.getByRole("dialog");
    await expect(dialog.getByText("Run action healthcheck")).toBeVisible();
    await dialog.getByRole("button", { name: "Run action" }).click();

    await expect(page).toHaveURL(/\/workflows\//, { timeout: 30000 });
    await expect(page.getByText("Action workflow started")).toBeVisible({
      timeout: 10000,
    });
  });
});
