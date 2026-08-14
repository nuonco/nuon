import { expect, test } from "@playwright/test";

const CREATE_STORY = "/?story=webhooks--webhookform--create&mode=preview";
const EDIT_STORY = "/?story=webhooks--webhookform--edit&mode=preview";

test.describe("WebhookForm create behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(CREATE_STORY, { waitUntil: "domcontentloaded" });
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

test.describe("WebhookForm edit behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(EDIT_STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("rotating the secret gates submit until a value is entered", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Save changes" });

    await expect(submit).toBeEnabled();

    await dialog.getByText("Rotate to a new secret").click();
    await expect(submit).toBeDisabled();

    await dialog.getByPlaceholder("New signing secret").fill("s3cr3t");
    await expect(submit).toBeEnabled();
  });
});
