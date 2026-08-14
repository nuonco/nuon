import { expect, test } from "@playwright/test";

const STORY = "/?story=common--button--nudge&mode=preview";

test.beforeEach(async ({ page }) => {
  await page.goto(STORY, { waitUntil: "domcontentloaded" });
  await expect(
    page.getByRole("button", { name: "Simulate app config synced" }),
  ).toBeVisible();
});

test("the nudge opens when triggered and closes on click", async ({ page }) => {
  const nudge = page.getByRole("tooltip", {
    name: "Trigger a run to deploy this branch",
  });
  await expect(nudge).toBeHidden();

  await page.getByRole("button", { name: "Simulate app config synced" }).click();
  await expect(nudge).toBeVisible();

  await page.getByRole("button", { name: "Trigger run" }).click();
  await expect(nudge).toBeHidden();
});
