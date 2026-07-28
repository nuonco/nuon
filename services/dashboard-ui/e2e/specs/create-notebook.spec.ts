import { test, expect } from "../fixtures";

test.describe("Create notebook", () => {
  test.setTimeout(60000);

  test("create a notebook and redirect to it", async ({
    page,
    orgId,
    installIds,
  }) => {
    const installId = installIds[0];
    test.skip(!installId, "No seed install available");

    await page.goto(`/${orgId}/installs/${installId}/notebooks`);
    await page.waitForLoadState("domcontentloaded");

    const createButton = page
      .getByRole("button", { name: "Create notebook" })
      .first();
    await expect(createButton).toBeVisible({ timeout: 15000 });

    await expect(page).toHaveTitle(/^Notebooks \|/, { timeout: 15000 });

    const name = `e2e-notebook-${Date.now()}`;
    await createButton.click();
    const dialog = page.getByRole("dialog");
    const nameInput = dialog.getByPlaceholder("e.g. Debug pods");
    await expect(nameInput).toBeVisible();
    await nameInput.fill(name);
    await dialog.getByRole("button", { name: "Create notebook" }).click();
    await expect(page.getByText("Notebook created")).toBeVisible({
      timeout: 10000,
    });
    await expect(page).toHaveURL(/\/notebooks\//, { timeout: 15000 });
  });
});
