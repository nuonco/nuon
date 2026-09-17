import { expect, test } from "@playwright/test";

const STORY = "/?story=installs--installform--create-modal&mode=preview";

test.describe("InstallForm create behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("submit stays disabled until the form is valid", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Create install" });

    await expect(submit).toBeDisabled();

    await dialog.getByPlaceholder("Enter install name").fill("my-install");

    await expect(submit).toBeEnabled();
  });

  test("the name field errors only after it is touched", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const name = dialog.getByPlaceholder("Enter install name");

    await expect(dialog.getByText("Install name is required")).toBeHidden();

    await name.click();
    await name.blur();
    await expect(dialog.getByText("Install name is required")).toBeVisible();

    await name.fill("my-install");
    await expect(dialog.getByText("Install name is required")).toBeHidden();
  });

  test("stack and runner only lives in the footer, not the form body", async ({
    page,
  }) => {
    const dialog = page.getByRole("dialog");
    const stackOnly = dialog.locator('input[name="stackOnly"]');
    const cancel = dialog.getByRole("button", { name: "Cancel" });

    await expect(dialog.getByText("Provisioning scope")).toHaveCount(0);
    await expect(stackOnly).toBeVisible();

    const stackBox = await stackOnly.boundingBox();
    const cancelBox = await cancel.boundingBox();
    expect(stackBox && cancelBox).toBeTruthy();
    expect(Math.abs(stackBox!.y - cancelBox!.y)).toBeLessThan(80);
  });
});
