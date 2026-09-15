import { expect, test } from "@playwright/test";

const STORY = "/?story=installs--appselect--default&mode=preview";
const FORM_STORY =
  "/?story=installs--installform--create-aws-stack-only&mode=preview";

test.describe("AppSelect readiness", () => {
  test("keeps apps without a runner config disabled", async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    const radio = page.getByRole("radio", { name: /Dev App/ });
    await expect(radio).toBeDisabled();
    await expect(page.getByText("Not provisionable")).toBeVisible();
  });

  test("lets the user pick an app with no components", async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    const radio = page.getByRole("radio", { name: /Sandbox only/ });
    await expect(page.getByText("No components")).toBeVisible();
    await expect(radio).toBeEnabled();
    await radio.check();
    await expect(radio).toBeChecked();
  });

  test("lets the user pick an app with no component builds", async ({
    page,
  }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    const radio = page.getByRole("radio", { name: /Unbuilt components/ });
    await expect(page.getByText("No component builds")).toBeVisible();
    await expect(radio).toBeEnabled();
    await radio.check();
    await expect(radio).toBeChecked();
  });
});

test.describe("InstallForm stack-only default", () => {
  test("checks stack and runner only when defaulted", async ({ page }) => {
    await page.goto(FORM_STORY, { waitUntil: "domcontentloaded" });
    await expect(page.getByText("Install name")).toBeVisible();
    await expect(page.locator('input[name="stackOnly"]')).toBeChecked();
  });
});
