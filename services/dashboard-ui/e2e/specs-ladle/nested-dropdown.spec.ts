import { expect, test } from "@playwright/test";

const STORY = "/?story=common--dropdown--nested-dropdowns&mode=preview";

test.beforeEach(async ({ page }) => {
  await page.goto(STORY, { waitUntil: "domcontentloaded" });
  await expect(
    page.getByRole("button", { name: "Main Menu" }),
  ).toBeVisible();
});

test("opening a nested dropdown keeps the parent open", async ({ page }) => {
  await page.getByRole("button", { name: "Main Menu" }).click();
  await expect(page.getByRole("button", { name: "Preferences" })).toBeVisible();

  await page.getByRole("button", { name: "Preferences" }).click();

  await expect(page.getByRole("button", { name: "Theme" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Preferences" })).toBeVisible();

  await page.waitForTimeout(400);
  await expect(page.getByRole("button", { name: "Theme" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Preferences" })).toBeVisible();
});

test("opening a third-level dropdown keeps both ancestors open", async ({
  page,
}) => {
  await page.getByRole("button", { name: "Main Menu" }).click();
  await page.getByRole("button", { name: "Preferences" }).click();
  await expect(page.getByRole("button", { name: "Advanced" })).toBeVisible();

  await page.getByRole("button", { name: "Advanced" }).click();

  await expect(page.getByRole("button", { name: "Debug Mode" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Theme" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Preferences" })).toBeVisible();
});

test("deep nesting keeps every ancestor open at each level", async ({
  page,
}) => {
  await page.getByRole("button", { name: "Level 1", exact: true }).click();

  for (let level = 2; level <= 6; level++) {
    await page
      .getByRole("button", { name: `Level ${level}`, exact: true })
      .click();

    for (let open = 1; open < level; open++) {
      await expect(
        page.getByRole("button", { name: `Level ${open}`, exact: true }),
      ).toBeVisible();
    }
  }
});

test("clicking a leaf item closes the whole chain", async ({ page }) => {
  await page.getByRole("button", { name: "Main Menu" }).click();
  await page.getByRole("button", { name: "Preferences" }).click();
  await expect(page.getByRole("button", { name: "Theme" })).toBeVisible();

  await page.getByRole("button", { name: "Theme" }).click();

  await expect(page.getByRole("button", { name: "Theme" })).toBeHidden();
  await expect(page.getByRole("button", { name: "Preferences" })).toBeHidden();
});

test("clicking outside closes the whole chain", async ({ page }) => {
  await page.getByRole("button", { name: "Main Menu" }).click();
  await page.getByRole("button", { name: "Preferences" }).click();
  await expect(page.getByRole("button", { name: "Theme" })).toBeVisible();

  await page.mouse.click(5, 5);

  await expect(page.getByRole("button", { name: "Theme" })).toBeHidden();
  await expect(page.getByRole("button", { name: "Preferences" })).toBeHidden();
});
