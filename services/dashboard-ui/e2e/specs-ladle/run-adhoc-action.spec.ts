import { expect, test } from "@playwright/test";

const STORY = "/?story=installs--runadhocaction--default&mode=preview";
const COMMAND_PLACEHOLDER = "echo 'Hello, world!'";

test.describe("RunAdhocAction form behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("submit stays disabled until the form is valid", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Run action" });

    await expect(submit).toBeDisabled();

    await dialog.getByPlaceholder(COMMAND_PLACEHOLDER).fill("echo hello");

    await expect(submit).toBeEnabled();
  });

  test("the command field errors only after it is touched", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const command = dialog.getByPlaceholder(COMMAND_PLACEHOLDER);

    await expect(dialog.getByText("Command is required")).toBeHidden();

    await command.click();
    await command.blur();
    await expect(dialog.getByText("Command is required")).toBeVisible();

    await command.fill("echo hello");
    await expect(dialog.getByText("Command is required")).toBeHidden();
  });
});
