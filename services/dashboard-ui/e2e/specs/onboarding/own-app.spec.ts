import { test, expect } from "@playwright/test";
import { env } from "../../env";
import {
  currentOrgId,
  deployToWorkflow,
  firstRun,
  freshAccount,
  openAs,
  publicApi,
  startOnboarding,
} from "../../onboarding";

// The own-app path against a fixture repo per cloud. The repo already has a
// config at its root, so the branch's first run syncs it and no push is needed;
// the spec continues past the push warning and the stack by hand.
const CLOUDS = [
  { key: "aws", label: "AWS", envVar: "E2E_ONBOARDING_REPO_AWS" },
  { key: "gcp", label: "GCP", envVar: "E2E_ONBOARDING_REPO_GCP" },
  { key: "azure", label: "Azure", envVar: "E2E_ONBOARDING_REPO_AZURE" },
] as const;

test.describe("Onboarding: own app", () => {
  test.setTimeout(8 * 60_000);

  for (const cloud of CLOUDS) {
    test(`connect a ${cloud.label} app template and land on the workflow`, async ({ browser }) => {
      const repo = env.onboardingRepo[cloud.key];
      test.skip(!env.githubInstallId, "E2E_GITHUB_INSTALL_ID is not set");
      test.skip(!repo, `${cloud.envVar} is not set`);
      const appName = repo!.split("/")[1];

      const { token } = await freshAccount();
      const page = await openAs(browser, token);
      await startOnboarding(page);
      const orgId = await currentOrgId(token);

      // What the GitHub App callback does, without leaving for github.com.
      const connection = await publicApi<{ id: string }>(token, "/v1/vcs/connection-callback", {
        method: "POST",
        orgId,
        body: { github_install_id: env.githubInstallId, org_id: orgId },
      });
      await page.goto(`/onboarding?vcs-connected=${connection.id}`);
      await page.getByRole("button", { name: /Start with your app/ }).click({ timeout: 30000 });
      await expect(page.getByText("Connected as")).toBeVisible({ timeout: 30000 });
      await expect(page.getByRole("button", { name: /^Go to step/ })).toHaveCount(5);

      // The name is validated as typed and never rewritten.
      const name = page.locator("#first-run-app-name");
      await name.fill("My App");
      await expect(page.locator("#first-run-app-name-description")).toHaveText(
        "Lowercase letters, numbers, underscores, and hyphens only.",
      );
      await expect(page.getByRole("button", { name: /^Next/ })).toBeDisabled();
      await expect(name).toHaveValue("My App");

      await name.fill(appName);
      await page.locator(`label[title="${cloud.label}"]`).click();
      await page.getByRole("button", { name: /^Next/ }).click();

      await expect(page.getByRole("heading", { name: "Fill in your app template" })).toBeVisible({
        timeout: 60000,
      });
      await expect(page.getByText(`in your ${cloud.label} test account`)).toBeVisible();
      await expect(page.getByRole("link", { name: /See full prompt/ })).toHaveAttribute(
        "href",
        "https://nuon.co/loop.md",
      );

      await page.getByRole("button", { name: /Set up your first install/ }).click();
      await expect(page.getByText("Nuon does not have your app config yet")).toBeVisible();
      await page.getByRole("button", { name: "Continue anyway" }).click();

      await deployToWorkflow(page);

      const journey = await firstRun(token);
      const meta = Object.assign({}, ...(journey?.steps ?? []).map((s) => s.metadata ?? {}));
      expect(meta).toMatchObject({ path: "own", cloud: cloud.key, app_name: appName, repo });
      const app = await publicApi<{ config_repo: string; config_directory: string }>(
        token,
        `/v1/apps/${meta.app_id}`,
        { orgId },
      );
      expect(app).toMatchObject({ config_repo: repo, config_directory: "." });
    });
  }
});
