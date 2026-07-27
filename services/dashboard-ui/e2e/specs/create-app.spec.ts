import { test, expect } from "../fixtures";

test.describe("Create app", () => {
  test.setTimeout(60000);

  test("create an app and redirect to its branches", async ({
    page,
    orgId,
  }) => {
    await page.goto(`/${orgId}/apps`);
    await page.waitForLoadState("domcontentloaded");
    await expect(page).toHaveTitle(/^Apps \|/, { timeout: 15000 });

    const createButton = page
      .getByRole("button", { name: "Create app" })
      .first();
    await expect(createButton).toBeVisible({ timeout: 15000 });

    const name = `e2e-app-${Date.now()}`;
    await createButton.click();
    const dialog = page.getByRole("dialog");
    await expect(dialog.getByText("Create app")).toBeVisible();
    await dialog.getByPlaceholder("my-app").fill(name);
    await dialog.getByRole("button", { name: "Create", exact: true }).click();
    await expect(page.getByText("App created")).toBeVisible({ timeout: 10000 });
    await expect(page).toHaveURL(/\/apps\/.*\/branches/, { timeout: 15000 });
  });
});
