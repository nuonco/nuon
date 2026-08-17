import { expect, test } from "@playwright/test";

// Stack overrides have no required fields (all optional), so submit is always
// enabled. The meaningful behavior here is the custom-stacks array field:
// Add stack reveals a row, Remove clears it.
const STORY =
  "/?story=installs--management--editstackoverrides--empty&mode=preview";

test.describe("EditStackOverrides form behavior", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: "domcontentloaded" });
    await page
      .getByRole("button", { name: "Edit stack overrides (empty)" })
      .click();
    await expect(page.getByRole("dialog")).toBeVisible();
  });

  test("save is enabled and add/remove stack toggles a row", async ({
    page,
  }) => {
    const dialog = page.getByRole("dialog");
    const submit = dialog.getByRole("button", { name: "Save overrides" });

    await expect(submit).toBeEnabled();
    await expect(
      dialog.getByText("No custom nested stack overrides configured.")
    ).toBeVisible();

    await dialog.getByRole("button", { name: "Add stack" }).click();
    await expect(dialog.getByPlaceholder("e.g. k8s_namespaces")).toBeVisible();

    await dialog.getByRole("button", { name: "Remove stack 1" }).click();
    await expect(
      dialog.getByText("No custom nested stack overrides configured.")
    ).toBeVisible();
  });
});
