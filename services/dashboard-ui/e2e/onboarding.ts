import { expect, type Browser, type Page } from "@playwright/test";
import { env } from "./env";

// Helpers for the first-run onboarding specs. Each spec gets a brand-new
// account with no org, the way a real sign-up arrives at /onboarding.

async function admin(path: string, body: unknown) {
  const res = await fetch(`${env.adminApiUrl}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "X-Nuon-Admin-Email": env.email },
    body: JSON.stringify(body),
  });
  const text = await res.text();
  if (!res.ok) throw new Error(`Admin API ${path} failed (${res.status}): ${text}`);
  // Some admin routes append a second JSON object; the first is the response.
  const first = text.match(/^\{[^}]*\}/);
  return JSON.parse(first ? first[0] : text);
}

export async function publicApi<T = any>(
  token: string,
  path: string,
  opts: { method?: string; body?: unknown; orgId?: string } = {},
): Promise<T> {
  const res = await fetch(`${env.publicApiUrl}${path}`, {
    method: opts.method ?? "GET",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
      ...(opts.orgId ? { "X-Nuon-Org-ID": opts.orgId } : {}),
    },
    body: opts.body ? JSON.stringify(opts.body) : undefined,
  });
  const text = await res.text();
  if (!res.ok) throw new Error(`Public API ${path} failed (${res.status}): ${text}`);
  return (text ? JSON.parse(text) : undefined) as T;
}

// Integration users are fresh accounts; their own token lasts 10 minutes, so a
// longer static token is minted for the same account.
export async function freshAccount(): Promise<{ token: string; email: string }> {
  const user = await admin("/v1/general/integration-user", {});
  const { api_token } = await admin("/v1/general/admin-static-token", {
    duration: "2h",
    email_or_subject: user.email,
  });
  return { token: api_token, email: user.email };
}

export async function openAs(browser: Browser, token: string): Promise<Page> {
  const context = await browser.newContext({ storageState: { cookies: [], origins: [] } });
  const host = new URL(env.baseUrl).hostname;
  await context.addCookies([
    { name: "X-Nuon-Auth", value: token, domain: host, path: "/", httpOnly: true, sameSite: "Lax" },
  ]);
  return context.newPage();
}

export async function currentOrgId(token: string): Promise<string> {
  const orgs = await publicApi<{ id: string }[]>(token, "/v1/orgs");
  expect(orgs, "onboarding should have created exactly one org").toHaveLength(1);
  return orgs[0].id;
}

export async function firstRun(token: string) {
  const account = await publicApi<{ user_journeys?: any[] }>(token, "/v1/account");
  return account.user_journeys?.find((j) => j.name === "first_run") as
    | { steps: { name: string; complete: boolean; metadata?: Record<string, string> }[] }
    | undefined;
}

// Intro → wizard.
export async function startOnboarding(page: Page) {
  await page.goto("/onboarding");
  await page.waitForLoadState("domcontentloaded");
  await expect(page.getByRole("heading", { name: "Your account is set up" })).toBeVisible({
    timeout: 30000,
  });
  await page.getByRole("button", { name: /Create your first app template/ }).click();
  await expect(page.getByRole("heading", { name: "Create your first app template" })).toBeVisible();
}

// Create install → Stack (launch and Continue if it has not reported back) →
// Provision → the deploy workflow page.
export async function deployToWorkflow(page: Page) {
  await expect(page.getByRole("heading", { name: "Your app is ready for BYOC" })).toBeVisible({
    timeout: 60000,
  });
  await page.getByRole("button", { name: /Create install/ }).click();
  const stack = page.getByRole("heading", { name: "Create the install stack" });
  const provision = page.getByRole("heading", { name: "Your first BYOC install is deploying" });
  await expect(stack.or(provision)).toBeVisible({ timeout: 300000 });

  if (await stack.isVisible()) {
    // Launch opens the console in a new tab on AWS; the step only needs the click.
    const launch = page.getByRole("link", { name: /Open the CloudFormation stack/ }).or(
      page.getByRole("button", { name: /Get the Terraform stack|Get the Azure commands/ }),
    );
    // Stack generation usually takes under a minute but can take several.
    await expect(launch.or(provision)).toBeVisible({ timeout: 300000 });
    if (await launch.isVisible()) {
      const popup = page.context().waitForEvent("page", { timeout: 5000 }).catch(() => null);
      await launch.click();
      await (await popup)?.close();
      const next = page.getByRole("button", { name: /^Continue/ });
      await expect(next.or(provision)).toBeVisible({ timeout: 10000 });
      if (await next.isVisible()) await next.click();
    }
  }

  await expect(provision).toBeVisible({ timeout: 30000 });
  await expect(page.getByText(/nuon installs workflows watch -i inl/)).toBeVisible();
  await page.getByRole("button", { name: /Go to deploy workflow/ }).click();
  await expect(page).toHaveURL(/\/installs\/[^/]+\/history\/[^/]+/, { timeout: 30000 });
}
