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
  let seedInstallIds: string[] = [];

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

    log("adding support users...");
    const supportRes = await fetch(
      `${env.adminApiUrl}/v1/orgs/${orgId}/admin-support-users`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-Nuon-Admin-Email": "seed@nuon.co",
        },
      },
    );
    if (!supportRes.ok) {
      const body = await supportRes.text();
      log(`warning: failed to add support users: ${body}`);
    } else {
      log("support users added");
    }
  } else {
    log(`using existing org: ${orgId}`);

    log("fetching existing installs...");
    const installsRes = await apiFetch(api_token, "/v1/installs", {
      headers: { "X-Nuon-Org-ID": orgId },
    });
    const installsBody = (await installsRes.json()) as { data?: { id: string }[]; id?: string }[] | { data: { id: string }[] };
    if (Array.isArray(installsBody)) {
      seedInstallIds = installsBody.map((i: any) => i.id).filter(Boolean);
    } else if (installsBody.data) {
      seedInstallIds = installsBody.data.map((i) => i.id);
    }
    log(`found ${seedInstallIds.length} existing installs`);
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

    // Create 2 installs for tests that need pre-existing installs (e.g. label filtering)
    log("listing apps to find seeded app...");
    const appsRes = await apiFetch(api_token, "/v1/apps", {
      headers: { "X-Nuon-Org-ID": orgId },
    });
    const apps = (await appsRes.json()) as { id: string; name: string; runner_config?: { app_runner_type?: string } }[];
    const seededApp = apps.find((a) => a.name === env.appConfig);
    if (seededApp) {
      log(`found app ${seededApp.id}, creating 2 seed installs...`);
      const installIds: string[] = [];
      for (const name of ["e2e-install-alpha", "e2e-install-beta"]) {
        const body: Record<string, unknown> = {
          name,
          install_config: { approval_option: "prompt" },
          metadata: { managed_by: "nuon/e2e" },
        };
        const platform = seededApp.runner_config?.app_runner_type;
        if (platform === "aws") {
          body.aws_account = { iam_role_arn: "", region: "us-west-2" };
        }
        const installRes = await apiFetch(api_token, `/v1/apps/${seededApp.id}/installs`, {
          method: "POST",
          headers: { "X-Nuon-Org-ID": orgId },
          body: JSON.stringify(body),
        });
        const install = (await installRes.json()) as { id: string };
        installIds.push(install.id);
        log(`created install: ${name} (${install.id})`);
      }
      seedInstallIds = installIds;
    } else {
      log("warning: could not find seeded app, skipping install creation");
    }
  }

  fs.mkdirSync(path.dirname(ORG_STATE_PATH), { recursive: true });
  fs.writeFileSync(
    ORG_STATE_PATH,
    JSON.stringify({
      orgId,
      createdOrg,
      appConfig: env.appConfig,
      apiToken: api_token,
      installIds: seedInstallIds,
    }),
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
