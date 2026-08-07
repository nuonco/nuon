import { test, expect } from "../fixtures";

test.describe("Create branch", () => {
  test("create an app branch from a VCS repo", async ({
    page,
    orgId,
    appConfig,
  }) => {
    test.setTimeout(90000);
    test.skip(!appConfig, "No app config available");

    await page.goto(`/${orgId}/apps`);
    await page.waitForLoadState("domcontentloaded");
    const appLink = page.getByRole("link", { name: appConfig! }).first();
    await expect(appLink).toBeVisible({ timeout: 15000 });
    await appLink.click();
    await expect(page).toHaveURL(/\/apps\/app[a-z0-9]+/, { timeout: 15000 });
    const appId = page.url().match(/\/apps\/(app[a-z0-9]+)/)?.[1];
    expect(appId).toBeTruthy();

    await page.goto(`/${orgId}/apps/${appId}/branches`);
    await page.waitForLoadState("domcontentloaded");

    // The branches list (which hosts the create button) only renders when the app
    // has zero branches; once branches exist, /branches redirects to a branch
    // detail whose switcher offers no create action. Skip gracefully in that case.
    const openCreate = page
      .getByRole("button", { name: "Create branch" })
      .first();
    const reachable = await openCreate
      .isVisible({ timeout: 10000 })
      .catch(() => false);
    test.skip(
      !reachable,
      "Branches list not reachable — app already has branches (redirects to branch detail)"
    );
    await openCreate.click();

    const dialog = page.getByRole("dialog");
    await expect(dialog.getByLabel("Branch name")).toBeVisible({
      timeout: 10000,
    });

    const branchName = `e2e-branch-${Date.now()}`;
    await dialog.getByLabel("Branch name").fill(branchName);

    // Repository — pick the first real repo (options render in a body-level portal).
    const repoCombo = dialog.getByRole("combobox", { name: "Repository" });
    await repoCombo.click();
    await expect(page.getByRole("option").first()).toBeVisible({
      timeout: 15000,
    });
    await page.getByRole("option").first().click();

    // Git branch — a searchable Select once branches load, or a text input fallback.
    const gitBranchCombo = dialog.getByRole("combobox", { name: "Git branch" });
    const gitBranchInput = dialog.getByRole("textbox", { name: "Git branch" });
    await expect(gitBranchCombo.or(gitBranchInput).first()).toBeVisible({
      timeout: 15000,
    });
    if ((await gitBranchCombo.count()) > 0) {
      await gitBranchCombo.click();
      await page.getByRole("option").first().click();
    } else {
      await gitBranchInput.fill("main");
    }

    await dialog.getByRole("button", { name: "Create branch" }).click();

    await expect(page.getByText("Branch created")).toBeVisible({
      timeout: 20000,
    });
  });
});
