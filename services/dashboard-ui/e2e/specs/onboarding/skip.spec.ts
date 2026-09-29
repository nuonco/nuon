import { test, expect } from "@playwright/test";
import { currentOrgId, firstRun, freshAccount, openAs, startOnboarding } from "../../onboarding";

test.describe("Onboarding: skip", () => {
  test.setTimeout(2 * 60_000);

  test("skip completes first_run and lands on the org dashboard", async ({ browser }) => {
    const { token } = await freshAccount();
    const page = await openAs(browser, token);

    await startOnboarding(page);
    const orgId = await currentOrgId(token);

    await page.getByRole("button", { name: /Skip/ }).click();
    await expect(page).toHaveURL(new RegExp(`/${orgId}$`), { timeout: 30000 });

    const journey = await firstRun(token);
    expect(journey?.steps.every((s) => s.complete)).toBe(true);
    expect(journey?.steps.find((s) => s.name === "start")?.metadata?.skipped).toBe("true");

    await page.goto("/");
    await expect(page).toHaveURL(new RegExp(`/${orgId}$`), { timeout: 30000 });
  });
});
