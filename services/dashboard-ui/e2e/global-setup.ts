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
        "X-Nuon-Admin-Email": env.adminEmail,
      },
      body: JSON.stringify({
        email_or_subject: env.userEmail,
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

  const { token } = (await res.json()) as { token: string };

  const baseUrl = new URL(env.baseUrl);
  const browser = await chromium.launch();
  const context = await browser.newContext();

  await context.addCookies([
    {
      name: "X-Nuon-Auth",
      value: token,
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
