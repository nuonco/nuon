import { chromium, type FullConfig } from "@playwright/test";
import { env } from "./env";

const AUTH_STATE_PATH = "e2e/.auth/user.json";

export default async function globalSetup(_config: FullConfig) {
  const res = await fetch(
    `${env.adminApiUrl}/v1/general/admin-static-token`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Nuon-Admin-Email": env.email,
      },
      body: JSON.stringify({
        email_or_subject: env.email,
        duration: "1h",
      }),
    },
  );

  if (!res.ok) {
    const body = await res.text();
    throw new Error(
      `Failed to generate static token (${res.status}): ${body}`,
    );
  }

  const { api_token } = (await res.json()) as { api_token: string };

  const baseUrl = new URL(env.baseUrl);
  const browser = await chromium.launch();
  const context = await browser.newContext();

  await context.addCookies([
    {
      name: "X-Nuon-Auth",
      value: api_token,
      domain: baseUrl.hostname,
      path: "/",
      httpOnly: true,
      sameSite: "Lax",
    },
  ]);

  const page = await context.newPage();
  await page.goto(`${env.baseUrl}/${env.orgId}`);
  await page.waitForLoadState("networkidle");

  await context.storageState({ path: AUTH_STATE_PATH });
  await browser.close();
}
