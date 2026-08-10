import { expect, test } from "@playwright/test";

// Behavior test for the composite-widget pilot: proves the ChannelSelect /
// MatchPicker / InterestsPicker wrappers bind to form state (a composite field
// drives canSubmit) with no backend — the default story ships mock channels.
const STORY = "/?story=slack--createchannelsubscription--default&mode=preview";

test.describe("CreateChannelSubscription form behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
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

    // MatchPicker default (undefined match → "all" mode)
    await expect(
      dialog.getByText("Everything in this org")
    ).toBeVisible();
    // InterestsPicker default (allEvents → "All events" summary)
    await expect(dialog.getByText("All events")).toBeVisible();
  });
});
