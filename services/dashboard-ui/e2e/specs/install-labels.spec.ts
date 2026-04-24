import { test, expect } from "../fixtures";

test.describe("Install labels", () => {
  test.setTimeout(90000);

  test("add labels, verify, edit labels, verify, then filter installs list", async ({
    page,
    orgId,
    installIds,
  }) => {
    const installId = installIds[0];
    test.skip(!installId, "No seed install available");

    // --- Step 1: Navigate to the install page ---
    await page.goto(`/${orgId}/installs/${installId}`);
    await page.waitForLoadState("networkidle");

    // --- Step 2: Open Manage → Edit labels, clear any existing labels ---
    await page.getByRole("button", { name: "Manage" }).click();
    await page.getByRole("button", { name: "Edit labels" }).click();
    await expect(page.getByRole("button", { name: "Add label" })).toBeVisible();

    // Remove any pre-existing label rows
    while ((await page.locator("fieldset").count()) > 0) {
      await page.locator("fieldset button").first().click();
    }

    // --- Step 3: Add 3 labels ---
    for (const [key, value] of [
      ["env", "staging"],
      ["team", "platform"],
      ["region", "us-west-2"],
    ]) {
      await page.getByRole("button", { name: "Add label" }).click();
      const keyInputs = page.locator('input[name^="label:"][name$=":key"]');
      const valueInputs = page.locator('input[name^="label:"][name$=":value"]');
      await keyInputs.last().fill(key);
      await valueInputs.last().fill(value);
    }

    // Save — wait for the labels API response, then modal close
    const saveResponse = page.waitForResponse(
      (r) => r.url().includes("/labels") && r.request().method() === "POST"
    );
    await page.getByRole("button", { name: "Save labels" }).click();
    await saveResponse;
    await expect(
      page.getByRole("button", { name: "Save labels" })
    ).not.toBeVisible({ timeout: 10000 });

    // --- Step 4: Reload and verify all 3 labels ---
    await page.reload();
    await page.waitForLoadState("networkidle");
    await expect(page.getByText("env: staging")).toBeVisible({ timeout: 10000 });
    await expect(page.getByText("team: platform")).toBeVisible();
    await expect(page.getByText("region: us-west-2")).toBeVisible();

    // --- Step 5: Open Edit labels to rename one and remove one ---
    await page.getByRole("button", { name: "Manage" }).click();
    await page.getByRole("button", { name: "Edit labels" }).click();

    const keyInputs = page.locator('input[name^="label:"][name$=":key"]');
    await expect(keyInputs).toHaveCount(3, { timeout: 10000 });

    // Rows are sorted alphabetically: env, region, team
    const envKeyInput = keyInputs.first();
    await expect(envKeyInput).toHaveValue("env");
    await envKeyInput.clear();
    await envKeyInput.fill("environment");

    // Remove last row (team)
    await page.locator("fieldset button").last().click();
    await expect(keyInputs).toHaveCount(2);

    // Save — wait for labels API responses
    const editSaveResponse = page.waitForResponse(
      (r) => r.url().includes("/labels") && r.request().method() === "POST"
    );
    await page.getByRole("button", { name: "Save labels" }).click();
    await editSaveResponse;
    await expect(
      page.getByRole("button", { name: "Save labels" })
    ).not.toBeVisible({ timeout: 10000 });

    // --- Step 6: Reload and verify updated labels ---
    await page.reload();
    await page.waitForLoadState("networkidle");
    await expect(page.getByText("environment: staging")).toBeVisible({
      timeout: 10000,
    });
    await expect(page.getByText("region: us-west-2")).toBeVisible();
    await expect(page.getByText("team: platform")).not.toBeVisible();

    // --- Step 7: Navigate to installs list and test label filter ---
    await page.goto(`/${orgId}/installs`);
    await page.waitForLoadState("networkidle");

    const labelsDropdown = page.getByRole("button", { name: /Labels/ });
    await expect(labelsDropdown).toBeVisible({ timeout: 10000 });
    await labelsDropdown.click();

    await expect(page.getByText("environment:staging")).toBeVisible();
    await expect(page.getByText("region:us-west-2")).toBeVisible();

    // Toggle a checkbox filter
    const envCheckbox = page.locator(
      'input[type="checkbox"][value="environment:staging"]'
    );
    await envCheckbox.click({ force: true });

    await expect(
      page.getByRole("button", { name: /Labels \(1\)/ })
    ).toBeVisible();

    // Filtered install should be visible
    await expect(page.getByText(installId)).toBeVisible({ timeout: 10000 });

    // Use "Only" on the other label
    await page.getByRole("button", { name: /region:us-west-2/ }).click();
    await expect(
      page.getByRole("button", { name: /Labels \(1\)/ })
    ).toBeVisible();

    // Reset — the dropdown content is portaled to body
    await page
      .locator('[id="dropdown-content-labels-filter"]')
      .getByRole("button", { name: "Reset", exact: true })
      .click();

    await expect(
      page.getByRole("button", { name: "Labels" })
    ).toBeVisible();
  });
});
