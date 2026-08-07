import { test, expect } from "../fixtures";

test.describe("Edit stack overrides", () => {
  test.setTimeout(60000);

  test("open the editor, set a VPC template URL, and save", async ({
    page,
    orgId,
    installIds,
  }) => {
    const installId = installIds[0];
    test.skip(!installId, "No seed install available");

    await page.goto(`/${orgId}/installs/${installId}/stacks`);
    await page.waitForLoadState("domcontentloaded");
    await expect(page).toHaveTitle(/^Stacks \|/, { timeout: 15000 });

    const openButton = page.getByRole("button", {
      name: "Edit stack overrides",
    });
    await expect(openButton).toBeVisible({ timeout: 15000 });
    test.skip(
      await openButton.isDisabled(),
      "Install is config-managed; stack overrides are read-only"
    );
    await openButton.click();

    const dialog = page.getByRole("dialog");
    await expect(dialog.getByRole("textbox").first()).toBeVisible({
      timeout: 10000,
    });
    await dialog
      .getByRole("textbox")
      .first()
      .fill("https://s3.amazonaws.com/nuon-e2e/vpc-template.yaml");

    await dialog.getByRole("button", { name: "Save overrides" }).click();

    await expect(page.getByText("Stack overrides updated")).toBeVisible({
      timeout: 15000,
    });
  });
});
