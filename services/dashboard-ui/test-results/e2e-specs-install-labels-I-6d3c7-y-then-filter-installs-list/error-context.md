# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: e2e/specs/install-labels.spec.ts >> Install labels >> add labels, verify, edit labels, verify, then filter installs list
- Location: e2e/specs/install-labels.spec.ts:6:7

# Error details

```
Error: page.goto: Protocol error (Page.navigate): Cannot navigate to invalid URL
Call log:
  - navigating to "/orgeo4ggrwdwtgd15lgjaqcwke/installs/inlcm31l1kvnp6pezlqd7r4jms", waiting until "load"

```

# Test source

```ts
  1   | import { test, expect } from "../fixtures";
  2   | 
  3   | test.describe("Install labels", () => {
  4   |   test.setTimeout(90000);
  5   | 
  6   |   test("add labels, verify, edit labels, verify, then filter installs list", async ({
  7   |     page,
  8   |     orgId,
  9   |     installIds,
  10  |   }) => {
  11  |     const installId = installIds[0];
  12  |     test.skip(!installId, "No seed install available");
  13  | 
  14  |     // --- Step 1: Navigate to the install page ---
> 15  |     await page.goto(`/${orgId}/installs/${installId}`);
      |                ^ Error: page.goto: Protocol error (Page.navigate): Cannot navigate to invalid URL
  16  |     await page.waitForLoadState("networkidle");
  17  | 
  18  |     // --- Step 2: Open Manage → Edit labels ---
  19  |     await page.getByRole("button", { name: "Manage" }).click();
  20  |     await page.getByRole("button", { name: "Edit labels" }).click();
  21  | 
  22  |     // Modal should be visible
  23  |     await expect(page.getByText("Edit labels").first()).toBeVisible();
  24  | 
  25  |     // --- Step 3: Add 3 labels ---
  26  |     for (const [key, value] of [
  27  |       ["env", "staging"],
  28  |       ["team", "platform"],
  29  |       ["region", "us-west-2"],
  30  |     ]) {
  31  |       await page.getByRole("button", { name: "Add label" }).click();
  32  |       // Fill the last row's key and value inputs
  33  |       const keyInputs = page.locator('input[name^="label:"][name$=":key"]');
  34  |       const valueInputs = page.locator('input[name^="label:"][name$=":value"]');
  35  |       const lastKey = keyInputs.last();
  36  |       const lastValue = valueInputs.last();
  37  |       await lastKey.fill(key);
  38  |       await lastValue.fill(value);
  39  |     }
  40  | 
  41  |     // Save
  42  |     await page.getByRole("button", { name: "Save labels" }).click();
  43  | 
  44  |     // Wait for modal to close and labels to appear
  45  |     await expect(
  46  |       page.getByRole("button", { name: "Save labels" })
  47  |     ).not.toBeVisible({ timeout: 10000 });
  48  | 
  49  |     // --- Step 4: Verify all 3 labels render in the install header ---
  50  |     const header = page.locator("header").first();
  51  |     await expect(header.getByText("env: staging")).toBeVisible({ timeout: 10000 });
  52  |     await expect(header.getByText("team: platform")).toBeVisible();
  53  |     await expect(header.getByText("region: us-west-2")).toBeVisible();
  54  | 
  55  |     // --- Step 5: Open Edit labels again to rename one and remove one ---
  56  |     await page.getByRole("button", { name: "Manage" }).click();
  57  |     await page.getByRole("button", { name: "Edit labels" }).click();
  58  |     await expect(
  59  |       page.locator('input[name^="label:"][name$=":key"]').first()
  60  |     ).toBeVisible();
  61  | 
  62  |     // The modal should have 3 rows pre-populated (sorted: env, region, team)
  63  |     const keyInputs = page.locator('input[name^="label:"][name$=":key"]');
  64  |     const valueInputs = page.locator('input[name^="label:"][name$=":value"]');
  65  |     await expect(keyInputs).toHaveCount(3);
  66  | 
  67  |     // Rename "env: staging" to "environment: staging" (first row after sort)
  68  |     const envKeyInput = keyInputs.first();
  69  |     await expect(envKeyInput).toHaveValue("env");
  70  |     await envKeyInput.clear();
  71  |     await envKeyInput.fill("environment");
  72  | 
  73  |     // Remove the 3rd label (team: platform — last row after sort)
  74  |     const removeButtons = page.locator("fieldset button");
  75  |     await removeButtons.last().click();
  76  |     await expect(keyInputs).toHaveCount(2);
  77  | 
  78  |     // Save
  79  |     await page.getByRole("button", { name: "Save labels" }).click();
  80  |     await expect(
  81  |       page.getByRole("button", { name: "Save labels" })
  82  |     ).not.toBeVisible({ timeout: 10000 });
  83  | 
  84  |     // --- Step 6: Verify updated labels in header ---
  85  |     await expect(header.getByText("environment: staging")).toBeVisible({
  86  |       timeout: 10000,
  87  |     });
  88  |     await expect(header.getByText("region: us-west-2")).toBeVisible();
  89  |     // Removed labels should be gone
  90  |     await expect(header.getByText("team: platform")).not.toBeVisible();
  91  |     await expect(header.getByText("env: staging")).not.toBeVisible();
  92  | 
  93  |     // --- Step 7: Navigate to installs list and test label filter ---
  94  |     await page.goto(`/${orgId}/installs`);
  95  |     await page.waitForLoadState("networkidle");
  96  | 
  97  |     // The Labels filter dropdown should be visible
  98  |     const labelsDropdown = page.getByRole("button", { name: /Labels/ });
  99  |     await expect(labelsDropdown).toBeVisible({ timeout: 10000 });
  100 |     await labelsDropdown.click();
  101 | 
  102 |     // Should see our labels as filter options
  103 |     await expect(page.getByText("environment:staging")).toBeVisible();
  104 |     await expect(page.getByText("region:us-west-2")).toBeVisible();
  105 | 
  106 |     // Select "environment:staging" checkbox
  107 |     const envCheckbox = page.locator(
  108 |       'input[type="checkbox"][value="environment:staging"]'
  109 |     );
  110 |     await envCheckbox.check();
  111 | 
  112 |     // The dropdown button should show the count
  113 |     await expect(
  114 |       page.getByRole("button", { name: /Labels \(1\)/ })
  115 |     ).toBeVisible();
```