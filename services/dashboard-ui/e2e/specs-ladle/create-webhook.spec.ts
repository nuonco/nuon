import { expect, test } from "@playwright/test";

const STORY = "/?story=webhooks--createwebhook--default&mode=preview";

test.describe("CreateWebhook form behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("submit is gated on a valid URL", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Create webhook" });
    const url = dialog.getByLabel("Webhook URL");

    await expect(submit).toBeDisabled();

    await url.fill("not-a-url");
    await url.blur();
    await expect(dialog.getByText("Enter a valid http or https URL")).toBeVisible();
    await expect(submit).toBeDisabled();

    await url.fill("https://example.com/hook");
    await expect(dialog.getByText("Enter a valid http or https URL")).toBeHidden();
    await expect(submit).toBeEnabled();
  });
});
