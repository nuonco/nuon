import { expect, test, type Page } from "@playwright/test";

const STORY = "/?story=orgs--orgstatusbar--full-context&mode=preview";
const BYOC_STORY =
  "/?story=orgs--orgstatusbar--full-context-with-byoc-badge&mode=preview";

const SINGLE_LINE_MAX_HEIGHT = 34;

const bar = (page: Page) => page.locator("div.bg-code").first();

const setBarWidth = async (page: Page, width: number) => {
  await page.evaluate((w) => {
    const frame = document.querySelector("div.bg-code")?.parentElement
    if (frame) frame.style.width = `${w}px`
  }, width);
  await page.waitForTimeout(100);
};

const measure = async (page: Page) =>
  page.evaluate(() => {
    const el = document.querySelector("div.bg-code") as HTMLElement;
    return {
      height: el.getBoundingClientRect().height,
      clipped: el.scrollWidth - el.clientWidth,
    };
  });

test.describe("org status bar", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await expect(bar(page)).toBeVisible();
  });

  test("stays on one line with nothing clipped as the bar narrows", async ({
    page,
  }) => {
    for (let width = 1600; width >= 480; width -= 40) {
      await setBarWidth(page, width);
      const { height, clipped } = await measure(page);

      expect(height, `bar wrapped at ${width}px`).toBeLessThanOrEqual(
        SINGLE_LINE_MAX_HEIGHT,
      );
      expect(clipped, `bar content clipped at ${width}px`).toBeLessThanOrEqual(
        1,
      );
    }
  });

  test("drops org, then app, then branch, and always keeps the install", async ({
    page,
  }) => {
    const orgName = page.getByText("acme-platform | dev");
    const appName = page.getByText("acme-enterprise-aws");
    const branchName = page.getByText("eng-sandbox");
    const installName = page.getByText("ws-workspace_01m1fbh", {
      exact: false,
    });

    await setBarWidth(page, 1200);
    await expect(orgName).toBeVisible();
    await expect(appName).toBeVisible();
    await expect(branchName).toBeVisible();
    await expect(installName).toBeVisible();

    await setBarWidth(page, 840);
    await expect(orgName).toBeHidden();
    await expect(appName).toBeVisible();
    await expect(branchName).toBeVisible();

    await setBarWidth(page, 720);
    await expect(appName).toBeHidden();
    await expect(branchName).toBeVisible();

    await setBarWidth(page, 560);
    await expect(branchName).toBeHidden();
    await expect(installName).toBeVisible();
  });

  test("keeps the install status icons visible at half width", async ({
    page,
  }) => {
    await setBarWidth(page, 720);

    const icons = bar(page).locator("svg");
    const count = await icons.count();
    expect(count).toBeGreaterThan(5);

    for (let i = 0; i < count; i++) {
      await expect(icons.nth(i)).toBeInViewport();
    }
  });

  test("keeps the BYOC badge intact on a narrow bar", async ({ page }) => {
    await page.goto(BYOC_STORY, { waitUntil: "domcontentloaded" });
    await expect(bar(page)).toBeVisible();
    await setBarWidth(page, 720);

    await expect(page.getByText("acme payments")).toBeVisible();
    const { height, clipped } = await measure(page);
    expect(height).toBeLessThanOrEqual(SINGLE_LINE_MAX_HEIGHT);
    expect(clipped).toBeLessThanOrEqual(1);
  });
});
