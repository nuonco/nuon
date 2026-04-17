import { chromium, type FullConfig } from "@playwright/test";
import { env } from "./env";
import fs from "node:fs";
import path from "node:path";

const AUTH_STATE_PATH = "e2e/.auth/user.json";
const ORG_STATE_PATH = "e2e/.auth/org.json";

async function adminFetch(path: string, options: RequestInit = {}) {
  const res = await fetch(`${env.adminApiUrl}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      "X-Nuon-Admin-Email": env.email,
      ...options.headers,
    },
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`Admin API ${path} failed (${res.status}): ${body}`);
  }
  return res;
}

async function apiFetch(
  token: string,
  apiPath: string,
  options: RequestInit = {},
) {
  const res = await fetch(`${env.publicApiUrl}${apiPath}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
      ...options.headers,
    },
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`Public API ${apiPath} failed (${res.status}): ${body}`);
  }
  return res;
}

export default async function globalSetup(_config: FullConfig) {
  // seed-user creates the account if it doesn't exist and returns a token.
  // The admin middleware may append an error to the response body if the
  // X-Nuon-Admin-Email account doesn't exist yet, resulting in two concatenated
  // JSON objects. We parse only the first one.
  const seedRes = await fetch(
    `${env.adminApiUrl}/v1/general/seed-user`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Nuon-Admin-Email": env.email,
      },
      body: JSON.stringify({}),
    },
  );
  const seedBody = await seedRes.text();
  const firstJson = seedBody.match(/^\{[^}]*\}/);
  if (!firstJson) {
    throw new Error(`seed-user returned unexpected response: ${seedBody}`);
  }
  const { api_token } = JSON.parse(firstJson[0]) as { api_token: string };
  if (!api_token) {
    throw new Error(`seed-user did not return a token: ${seedBody}`);
  }

  let orgId = env.orgId;
  let createdOrg = false;

  if (!orgId) {
    const orgName = `e2e-test-${Date.now()}`;
    const orgRes = await apiFetch(api_token, "/v1/orgs", {
      method: "POST",
      body: JSON.stringify({ name: orgName, use_sandbox_mode: true }),
    });
    const org = (await orgRes.json()) as { id: string };
    orgId = org.id;
    createdOrg = true;
  }

  fs.mkdirSync(path.dirname(ORG_STATE_PATH), { recursive: true });
  fs.writeFileSync(
    ORG_STATE_PATH,
    JSON.stringify({ orgId, createdOrg }),
  );

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
  await page.goto(`${env.baseUrl}/${orgId}`);
  await page.waitForLoadState("networkidle");

  await context.storageState({ path: AUTH_STATE_PATH });
  await browser.close();
}
