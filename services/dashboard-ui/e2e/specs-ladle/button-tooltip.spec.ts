import { expect, test } from "@playwright/test";

const STORY = "/?story=common--button--tooltips&mode=preview";

test.beforeEach(async ({ page }) => {
  await page.goto(STORY, { waitUntil: "domcontentloaded" });
  await expect(page.getByRole("button", { name: "Deploy" })).toBeVisible();
});

test("disabled button still shows its reason tooltip on hover", async ({
  page,
}) => {
  const deploy = page.getByRole("button", { name: "Deploy" });
  await expect(deploy).toHaveAttribute("aria-disabled", "true");

  await deploy.hover({ force: true });

  await expect(
    page.getByRole("tooltip", { name: "Sync the app config first" }),
  ).toBeVisible();
});

test("enabled button shows its hint tooltip on hover", async ({ page }) => {
  await page.getByRole("button", { name: "Docs" }).hover();

  await expect(
    page.getByRole("tooltip", { name: "Opens the docs in a new tab" }),
  ).toBeVisible();
});
