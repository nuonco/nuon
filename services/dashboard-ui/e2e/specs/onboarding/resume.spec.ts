import { test, expect } from "@playwright/test";
import { firstRun, freshAccount, openAs, startOnboarding } from "../../onboarding";

test.describe("Onboarding: resume", () => {
  test.setTimeout(3 * 60_000);

  test("a reload returns to the step in progress with its data", async ({ browser }) => {
    const { token } = await freshAccount();
    const page = await openAs(browser, token);

    await startOnboarding(page);
    await page.getByRole("button", { name: "Deploy to AWS" }).click();
    const deploy = page.getByRole("heading", { name: "Your app is ready for BYOC" });
    await expect(deploy).toBeVisible({ timeout: 60000 });

    await page.reload();
    await page.waitForLoadState("domcontentloaded");
    // No intro on resume, straight back to Deploy with the example app's facts.
    await expect(deploy).toBeVisible({ timeout: 30000 });
    await expect(page.getByText("nuonco/kitchen-sink")).toBeVisible();
    await expect(page.getByText("nuonco/aws-eks-auto-sandbox")).toBeVisible();

    // The completed step keeps its check; going back shows the Start step.
    await page.getByRole("button", { name: /Back/ }).click();
    await expect(page.getByRole("heading", { name: "Create your first app template" })).toBeVisible();

    const journey = await firstRun(token);
    expect(journey?.steps.find((s) => s.name === "start")?.complete).toBe(true);
    expect(journey?.steps.find((s) => s.name === "deploy")?.complete).toBe(false);
  });
});
