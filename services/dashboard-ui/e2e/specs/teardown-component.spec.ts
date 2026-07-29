import { test, expect } from "../fixtures";
import { createThrowawayInstall } from "../helpers";

test.describe("Teardown component", () => {
  test.setTimeout(180000);

  test("teardown a component and redirect to its workflow", async ({
    page,
    orgId,
  }) => {
    const install = await createThrowawayInstall("e2e-teardown");
    test.skip(!install, "Could not create throwaway install");

    await page.goto(`/${orgId}/installs/${install!.id}/components`);
    await page.waitForLoadState("domcontentloaded");
    await expect(page).toHaveTitle(/^Components \|/, { timeout: 15000 });

    const componentLink = page
      .locator(`a[href*="/installs/${install!.id}/components/"]`)
      .first();
    await expect(componentLink).toBeVisible({ timeout: 15000 });
    await componentLink.click();
    await expect(page).toHaveURL(/\/components\/[^/]+$/, { timeout: 15000 });

    await page.getByRole("button", { name: "Component controls" }).click();
    await page.getByRole("button", { name: "Teardown component" }).click();

    const dialog = page.getByRole("dialog");
    const headingText =
      (await dialog.getByText(/^Teardown .+\?$/).textContent()) ?? "";
    const componentName = headingText
      .replace(/^Teardown\s+/, "")
      .replace(/\?\s*$/, "")
      .trim();
    expect(componentName.length).toBeGreaterThan(0);

    await dialog.locator("#confirm-component-name").fill(componentName);

    const teardownButton = dialog.getByRole("button", {
      name: "Teardown component",
    });
    await expect(teardownButton).toBeEnabled({ timeout: 10000 });
    await teardownButton.click();

    await expect(page).toHaveURL(/\/workflows(\/|$)/, { timeout: 30000 });
  });
});
