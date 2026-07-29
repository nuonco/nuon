import { test, expect } from "../fixtures";

test.describe("Service accounts", () => {
  test.setTimeout(60000);

  test("create, mint token, rename, change role, then delete", async ({
    page,
    orgId,
  }) => {
    const name = `e2e-sa-${Date.now()}`;
    const renamed = `${name}-renamed`;

    await page.goto(`/${orgId}/service-accounts`);
    await page.waitForLoadState("domcontentloaded");
    await expect(page).toHaveTitle(/^Service accounts \|/, { timeout: 15000 });

    // --- Create ---
    await page
      .getByRole("button", { name: "Create service account" })
      .first()
      .click();
    const createDialog = page.getByRole("dialog");
    const nameInput = createDialog.getByPlaceholder("e.g. ci-deploy");
    await expect(nameInput).toBeVisible();
    await nameInput.fill(name);
    await createDialog.getByRole("combobox").click();
    await page.getByRole("option").first().click();
    await createDialog
      .getByRole("button", { name: "Create service account" })
      .click();
    await expect(page.getByText("Service account created")).toBeVisible({
      timeout: 10000,
    });

    const row = page.getByRole("row").filter({ hasText: name });
    await expect(row).toBeVisible({ timeout: 10000 });
    const identity = ((await row.locator("td").nth(1).textContent()) ?? "").trim();

    // --- Mint a token ---
    await row.getByRole("button").last().click();
    await page.getByRole("button", { name: "Create token" }).click();
    const tokenDialog = page.getByRole("dialog");
    await tokenDialog.getByRole("button", { name: "Create token" }).click();
    await expect(
      tokenDialog.getByRole("button", { name: "Done" })
    ).toBeVisible({ timeout: 10000 });
    await tokenDialog.getByRole("button", { name: "Done" }).click();

    // --- Rename ---
    await row.getByRole("button").last().click();
    await page.getByRole("button", { name: "Rename" }).click();
    const renameDialog = page.getByRole("dialog");
    await expect(renameDialog.getByText("Rename service account")).toBeVisible();
    await renameDialog.getByPlaceholder("e.g. ci-deploy").fill(renamed);
    await renameDialog.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Service account renamed")).toBeVisible({
      timeout: 10000,
    });

    const renamedRow = page.getByRole("row").filter({ hasText: renamed });
    await expect(renamedRow).toBeVisible({ timeout: 10000 });

    // --- Change role (Save is disabled until the role actually changes) ---
    await renamedRow.getByRole("button").last().click();
    await page.getByRole("button", { name: "Change role" }).click();
    const roleDialog = page.getByRole("dialog");
    await expect(roleDialog.getByText("Change role")).toBeVisible();
    await roleDialog.getByRole("combobox").click();
    await page.getByRole("option", { selected: false }).first().click();
    await roleDialog.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Role updated")).toBeVisible({ timeout: 10000 });

    // --- Delete (type-to-confirm the identity) ---
    await renamedRow.getByRole("button").last().click();
    await page.getByRole("button", { name: "Delete service account" }).click();
    const deleteDialog = page.getByRole("dialog");
    await expect(deleteDialog.getByText("Delete service account?")).toBeVisible();
    await deleteDialog.locator("#confirm-service-account-identity").fill(identity);
    await deleteDialog
      .getByRole("button", { name: "Delete service account" })
      .click();
    await expect(page.getByText("Service account deleted")).toBeVisible({
      timeout: 10000,
    });
    await expect(renamedRow).not.toBeVisible();
  });
});
