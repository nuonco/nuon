import { test, expect } from "../fixtures";

test.describe("Webhooks CRUD", () => {
  test.setTimeout(60000);

  test("create, edit secret, then delete a webhook", async ({
    page,
    orgId,
  }) => {
    const url = `https://example.com/e2e-${Date.now()}`;

    await page.goto(`/${orgId}/webhooks`);
    await page.waitForLoadState("domcontentloaded");
    await expect(page).toHaveTitle(/^Webhooks \|/, { timeout: 15000 });

    // --- Create (default scope + all events, only the URL is required) ---
    await page.getByRole("button", { name: "Create webhook" }).first().click();
    const createDialog = page.getByRole("dialog");
    const urlInput = page.getByPlaceholder("https://example.com/webhooks/nuon");
    await expect(urlInput).toBeVisible();
    await urlInput.fill(url);
    await createDialog
      .getByRole("button", { name: "Create webhook" })
      .click();
    await expect(page.getByText("Webhook created")).toBeVisible({
      timeout: 10000,
    });

    const row = page.getByRole("row").filter({ hasText: url });
    await expect(row).toBeVisible({ timeout: 10000 });

    // --- Edit: set a new signing secret ---
    await row.getByRole("button", { name: "Edit" }).click();
    const editDialog = page.getByRole("dialog");
    await expect(editDialog.getByText("Edit webhook")).toBeVisible();
    await editDialog.getByRole("radio", { name: "Set a new secret" }).click();
    await editDialog.locator("#webhook-secret").fill("e2e-secret-value");
    await editDialog.getByRole("button", { name: "Save changes" }).click();
    await expect(page.getByText("Webhook updated")).toBeVisible({
      timeout: 10000,
    });

    // --- Delete ---
    await row.getByRole("button", { name: "Delete" }).click();
    const deleteDialog = page.getByRole("dialog");
    await expect(deleteDialog.getByText("Delete webhook?")).toBeVisible();
    await deleteDialog
      .getByRole("button", { name: "Delete webhook" })
      .click();
    await expect(page.getByText("Webhook deleted")).toBeVisible({
      timeout: 10000,
    });
    await expect(page.getByText(url)).not.toBeVisible();
  });
});
