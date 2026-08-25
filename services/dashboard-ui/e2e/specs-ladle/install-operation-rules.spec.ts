import { expect, test } from "@playwright/test";

const ALL_DISABLED_STORY =
  "/?story=operation-queues--install-operation-rules--all-disabled&mode=preview";
const ALL_ENABLED_STORY =
  "/?story=operation-queues--install-operation-rules--all-enabled&mode=preview";

test.describe("InstallOperationRules enable all", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(ALL_DISABLED_STORY, { waitUntil: "domcontentloaded" });
    await expect(
      page.getByRole("heading", { name: "Install operation rules" }),
    ).toBeVisible();
  });

  test("enable all turns every rule on and back off", async ({ page }) => {
    const enableAll = page.getByRole("switch", { name: "Enable all" });

    await expect(page.getByText("0 of 5 operation types")).toBeVisible();

    await enableAll.click();
    await expect(page.getByText("5 of 5 operation types")).toBeVisible();
    await expect(page.getByText("Rule enabled")).toHaveCount(5);

    await enableAll.click();
    await expect(page.getByText("0 of 5 operation types")).toBeVisible();
    await expect(page.getByText("Rule enabled")).toHaveCount(0);
  });
});

test.describe("InstallOperationRules window behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(ALL_DISABLED_STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("switch", { name: "Actions rule" }).click();
  });

  test("an anytime window has nothing to fall outside of", async ({ page }) => {
    await expect(
      page.getByRole("button", { name: "Weekly" }).first(),
    ).toBeVisible();
    await expect(
      page.getByText("When triggered outside of this window"),
    ).toBeHidden();
  });

  test("save stays disabled until a weekly window picks a day", async ({
    page,
  }) => {
    const save = page.getByRole("button", { name: "Save rules" });

    await page.getByRole("button", { name: "Weekly" }).first().click();
    await expect(page.getByText("Pick at least one day.")).toBeVisible();
    await expect(save).toBeDisabled();

    await page.getByRole("button", { name: "Wed" }).click();
    await expect(page.getByText("Pick at least one day.")).toBeHidden();
    await expect(save).toBeEnabled();
  });

  test("the outside-window checkbox gates the policy options", async ({
    page,
  }) => {
    await page.getByRole("button", { name: "Weekly" }).first().click();
    await page.getByRole("button", { name: "Wed" }).click();

    const askForPermission = page.getByRole("radio", {
      name: /Ask for permission/,
    });
    await expect(askForPermission).toBeVisible();

    await page.getByRole("checkbox").first().click();
    await expect(askForPermission).toBeHidden();
    await expect(
      page.getByText("Operations triggered outside the window run normally."),
    ).toBeVisible();
  });

  test("switching to monthly swaps the day picker", async ({ page }) => {
    await page.getByRole("button", { name: "Weekly" }).first().click();
    await expect(page.getByRole("button", { name: "Wed" })).toBeVisible();

    await page.getByRole("button", { name: "Monthly" }).first().click();
    await expect(page.getByRole("button", { name: "Wed" })).toBeHidden();
  });
});

test.describe("InstallOperationRules impact", () => {
  test("summarizes the enforced windows and policies", async ({ page }) => {
    await page.goto(ALL_ENABLED_STORY, { waitUntil: "domcontentloaded" });

    await expect(page.getByText(/Outside their window, /)).toBeVisible();
    await expect(page.getByText("Break glass: anytime.")).toBeVisible();
  });
});
