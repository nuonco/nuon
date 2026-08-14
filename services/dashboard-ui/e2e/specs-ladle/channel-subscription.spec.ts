import { expect, test } from "@playwright/test";

const CREATE_STORY =
  "/?story=slack--channelsubscriptionform--create&mode=preview";
const EDIT_STORY =
  "/?story=slack--channelsubscriptionform--edit-org-wide&mode=preview";

test.describe("ChannelSubscriptionForm create behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(CREATE_STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("submit stays disabled until a channel is selected", async ({ page }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Subscribe channel" });

    await expect(submit).toBeDisabled();

    await dialog.getByRole("combobox", { name: "Channel" }).click();
    await page.getByRole("option", { name: "#deploys" }).click();

    await expect(submit).toBeEnabled();
  });

  test("composite pickers render with their default field values", async ({
    page,
  }) => {
    const dialog = page.getByRole("dialog");

    await expect(dialog.getByText("Everything in this org")).toBeVisible();
    await expect(dialog.getByText("All events")).toBeVisible();
  });
});

test.describe("ChannelSubscriptionForm edit behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(EDIT_STORY, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Open modal" }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("identity is read-only and scope/events prefill; submit is enabled", async ({
    page,
  }) => {
    const dialog = page.getByRole("dialog");

    await expect(dialog.getByText("#deploys")).toBeVisible();
    await expect(dialog.getByRole("combobox", { name: "Channel" })).toHaveCount(
      0
    );
    await expect(dialog.getByText("All events")).toBeVisible();
    await expect(
      dialog.getByRole("button", { name: "Save changes" })
    ).toBeEnabled();
  });
});
