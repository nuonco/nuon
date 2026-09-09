import { expect, test, type Page } from "@playwright/test";

const BUILDS_STORY =
  "/?story=branches--branchrunbuilds--many-builds&mode=preview";
const EXPAND_STORY = "/?story=common--expand--basic-usage&mode=preview";

const overflow = (page: Page) =>
  page.evaluate(() => {
    const body = document.querySelector(".expand");
    if (!body) return null;
    const inner = body.firstElementChild as HTMLElement;
    return (
      inner.getBoundingClientRect().height -
      body.getBoundingClientRect().height
    );
  });

const expectNothingClipped = async (page: Page) => {
  await expect
    .poll(() => overflow(page), { timeout: 5000 })
    .toBeLessThanOrEqual(1);
};

test("an expanded card shows all of its content, however tall", async ({
  page,
}) => {
  await page.goto(BUILDS_STORY, { waitUntil: "domcontentloaded" });
  const rows = page.locator(".expand .divide-y > div");
  await expect(rows).toHaveCount(26);

  await expect(page.getByText("component-25")).toBeVisible();
  await expectNothingClipped(page);
  await expect(page.locator(".expand")).toHaveCSS("max-height", "none");
});

test("a tall card still collapses and reopens intact", async ({ page }) => {
  await page.goto(BUILDS_STORY, { waitUntil: "domcontentloaded" });
  const toggle = page.locator("button[aria-expanded]").first();
  await expect(page.locator(".expand")).toBeVisible();

  await toggle.click();
  await expect(page.locator(".expand")).toHaveCount(0);

  await toggle.click();
  await expect(page.locator(".expand .divide-y > div")).toHaveCount(26);
  await expectNothingClipped(page);
});

test("a collapsed expand mounts no content and animates on open", async ({
  page,
}) => {
  await page.goto(EXPAND_STORY, { waitUntil: "domcontentloaded" });
  const toggle = page.locator("button[aria-expanded]").first();
  await toggle.waitFor();
  await expect(page.locator(".expand")).toHaveCount(0);

  await toggle.click();
  await expect(page.locator(".expand")).toHaveCSS(
    "animation-name",
    "enter-content",
  );
  await expect(page.locator(".expand")).toHaveCSS("opacity", "1");
  await expectNothingClipped(page);
});
