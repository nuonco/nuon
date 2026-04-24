import { chromium, type FullConfig } from "@playwright/test";
import { execSync } from "node:child_process";
import { env } from "./env";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";

function log(message: string) {
  const elapsed = ((performance.now() - setupStart) / 1000).toFixed(1);
  console.log(`[e2e setup +${elapsed}s] ${message}`);
}

let setupStart = performance.now();

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

function seedApp(token: string, orgId: string, configName: string) {
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "e2e-app-config-"));
  try {
    log(`cloning example-app-configs (sparse: ${configName})...`);
    execSync(
      `git clone --depth 1 --filter=blob:none --sparse https://github.com/nuonco/example-app-configs.git .`,
      { cwd: tmpDir, stdio: "inherit" },
    );
    execSync(`git sparse-checkout set ${configName}`, {
      cwd: tmpDir,
      stdio: "inherit",
    });

    const cliEnv = {
      ...process.env,
      NUON_API_TOKEN: token,
      NUON_API_URL: env.publicApiUrl,
      NUON_ORG_ID: orgId,
      NUON_NO_TTY: "true",
    };

    log(`running: nuon apps create --name ${configName}`);
    execSync(`nuon apps create --name ${configName}`, {
      env: cliEnv,
      stdio: "inherit",
    });
    log(`running: nuon apps sync ${configName}`);
    execSync(`nuon apps sync ${path.join(tmpDir, configName)} || true`, {
      env: cliEnv,
      stdio: "inherit",
      shell: "/bin/sh",
    });
    log("app seeded");
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

export default async function globalSetup(_config: FullConfig) {
  setupStart = performance.now();
  log("starting global setup");

  log("seeding user account...");
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
  log("user seeded, token obtained");

  let orgId = env.orgId;
  let createdOrg = false;

  if (!orgId) {
    const orgName = `e2e-test-${Date.now()}`;
    log(`creating org: ${orgName}...`);
    const orgRes = await apiFetch(api_token, "/v1/orgs", {
      method: "POST",
      body: JSON.stringify({ name: orgName, use_sandbox_mode: true }),
    });
    const org = (await orgRes.json()) as { id: string };
    orgId = org.id;
    createdOrg = true;
    log(`org created: ${orgId}`);
  } else {
    log(`using existing org: ${orgId}`);
  }

  if (createdOrg) {
    log(`waiting for org ${orgId} to provision...`);
    const maxWait = 120_000;
    const pollInterval = 2_000;
    const start = Date.now();
    let orgActive = false;
    while (Date.now() - start < maxWait) {
      const orgRes = await apiFetch(api_token, `/v1/orgs/current`, {
        headers: { "X-Nuon-Org-ID": orgId },
      });
      const org = (await orgRes.json()) as { status: string; status_description: string };
      log(`org status: ${org.status} — ${org.status_description}`);
      if (org.status === "active") {
        orgActive = true;
        break;
      }
      if (org.status === "error" || org.status === "failed") {
        throw new Error(`Org provisioning failed: ${org.status} — ${org.status_description}`);
      }
      await new Promise((r) => setTimeout(r, pollInterval));
    }
    if (!orgActive) {
      throw new Error(`Org ${orgId} did not become active within ${maxWait / 1000}s`);
    }

    seedApp(api_token, orgId, env.appConfig);
  }

  fs.mkdirSync(path.dirname(ORG_STATE_PATH), { recursive: true });
  fs.writeFileSync(
    ORG_STATE_PATH,
    JSON.stringify({ orgId, createdOrg, appConfig: env.appConfig, apiToken: api_token }),
  );

  log("injecting auth cookie...");
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
  log("global setup complete");
}
