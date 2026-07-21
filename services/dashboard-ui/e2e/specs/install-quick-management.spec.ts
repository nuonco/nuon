import { test, expect, type Page } from "../fixtures";

async function openRowMenu(page: Page, installId: string) {
  const trigger = page.locator(`#dropdown-button-${installId}`);
  await expect(trigger).toBeVisible({ timeout: 15000 });
  await trigger.click();
}

test.describe("Install quick management dropdown", () => {
  test.setTimeout(60000);

  test("edit inputs opens without crashing", async ({
    page,
    orgId,
    installIds,
  }) => {
    const installId = installIds[0];
    test.skip(!installId, "No seed install available");

    await page.goto(`/${orgId}/installs`);
    await page.waitForLoadState("domcontentloaded");
    await expect(page.getByRole("heading", { name: "Installs" })).toBeVisible({
      timeout: 15000,
    });

    await openRowMenu(page, installId);
    await page.getByRole("button", { name: "Edit inputs" }).click();

    await expect(page.getByText("Edit install inputs")).toBeVisible({
      timeout: 15000,
    });
    await expect(page.getByText("Something went wrong.")).not.toBeVisible();
  });

  test("current inputs navigates to the inputs page", async ({
    page,
    orgId,
    installIds,
  }) => {
    const installId = installIds[0];
    test.skip(!installId, "No seed install available");

    await page.goto(`/${orgId}/installs`);
    await page.waitForLoadState("domcontentloaded");

    await openRowMenu(page, installId);
    await page.getByRole("link", { name: "Current inputs" }).click();

    await expect(page).toHaveURL(new RegExp(`/installs/${installId}/inputs`), {
      timeout: 15000,
    });
    await expect(
      page.getByText("The current input values for this install.")
    ).toBeVisible({ timeout: 15000 });
  });

  test("view state navigates to the state page", async ({
    page,
    orgId,
    installIds,
  }) => {
    const installId = installIds[0];
    test.skip(!installId, "No seed install available");

    await page.goto(`/${orgId}/installs`);
    await page.waitForLoadState("domcontentloaded");

    await openRowMenu(page, installId);
    await page.getByRole("link", { name: "View state" }).click();

    await expect(page).toHaveURL(new RegExp(`/installs/${installId}/state`), {
      timeout: 15000,
    });
  });
});
