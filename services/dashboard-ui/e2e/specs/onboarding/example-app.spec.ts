import { test, expect } from "@playwright/test";
import {
  currentOrgId,
  deployToWorkflow,
  firstRun,
  freshAccount,
  openAs,
  publicApi,
  startOnboarding,
} from "../../onboarding";

// The example path: Kitchen Sink deployed to a test cloud account, no GitHub.
const CLOUDS = [
  { label: "AWS", app: "kitchen-sink" },
  { label: "GCP", app: "kitchen-sink-gcp" },
] as const;

test.describe("Onboarding: example app", () => {
  test.setTimeout(8 * 60_000);

  for (const cloud of CLOUDS) {
    test(`deploy Kitchen Sink to ${cloud.label} and land on the workflow`, async ({ browser }) => {
      const { token } = await freshAccount();
      const page = await openAs(browser, token);

      await startOnboarding(page);
      const orgId = await currentOrgId(token);
      // Example path: Start, Deploy, Stack, Provision.
      await expect(page.getByRole("button", { name: /^Go to step/ })).toHaveCount(4);

      await page.getByRole("button", { name: `Deploy to ${cloud.label}` }).click();
      await deployToWorkflow(page);

      const apps = await publicApi<{ name: string }[]>(token, "/v1/apps", { orgId });
      expect(apps.map((a) => a.name)).toContain(cloud.app);

      const journey = await firstRun(token);
      expect(journey?.steps.every((s) => s.complete)).toBe(true);
      const meta = Object.assign({}, ...(journey?.steps ?? []).map((s) => s.metadata ?? {}));
      expect(meta).toMatchObject({ path: "example", cloud: cloud.label.toLowerCase(), app_name: cloud.app });
      expect(meta.install_id).toMatch(/^inl/);
    });
  }
});
