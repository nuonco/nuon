import fs from "node:fs";
import { env } from "./env";

const ORG_STATE_PATH = "e2e/.auth/org.json";

type OrgState = {
  orgId: string;
  apiToken: string;
  appConfig?: string;
  installIds?: string[];
};

function orgState(): OrgState {
  return JSON.parse(fs.readFileSync(ORG_STATE_PATH, "utf-8"));
}

async function api(apiPath: string, options: RequestInit = {}) {
  const { orgId, apiToken } = orgState();
  const res = await fetch(`${env.publicApiUrl}${apiPath}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${apiToken}`,
      "X-Nuon-Org-ID": orgId,
      ...options.headers,
    },
  });
  if (!res.ok) {
    throw new Error(`Public API ${apiPath} failed (${res.status}): ${await res.text()}`);
  }
  return res;
}

type SeededApp = {
  id: string;
  name: string;
  runner_config?: { app_runner_type?: string };
};

async function findSeededApp(): Promise<SeededApp | undefined> {
  const { appConfig } = orgState();
  const apps = (await (await api("/v1/apps")).json()) as SeededApp[];
  return apps.find((a) => a.name === appConfig) ?? apps[0];
}

export async function getSeededAppId(): Promise<string | undefined> {
  return (await findSeededApp())?.id;
}

export async function createThrowawayInstall(
  namePrefix: string,
): Promise<{ id: string; name: string } | null> {
  const app = await findSeededApp();
  if (!app) return null;

  const name = `${namePrefix}-${Date.now()}`;
  const body: Record<string, unknown> = {
    name,
    install_config: { approval_option: "prompt" },
    metadata: { managed_by: "nuon/e2e" },
  };
  if (app.runner_config?.app_runner_type === "aws") {
    body.aws_account = { iam_role_arn: "", region: "us-west-2" };
  }

  const install = (await (
    await api(`/v1/apps/${app.id}/installs`, {
      method: "POST",
      body: JSON.stringify(body),
    })
  ).json()) as { id: string };

  return { id: install.id, name };
}

export async function createTriggerableBranch(): Promise<{
  appId: string;
  branchId: string;
} | null> {
  const app = await findSeededApp();
  if (!app) return null;

  const install = await createThrowawayInstall("e2e-branch-install");
  if (!install) return null;
  const installId = install.id;

  const branch = (await (
    await api(`/v1/apps/${app.id}/branches`, {
      method: "POST",
      body: JSON.stringify({ name: `e2e-branch-${Date.now()}` }),
    })
  ).json()) as { id: string };

  await api(`/v1/apps/${app.id}/branches/${branch.id}/configs`, {
    method: "POST",
    body: JSON.stringify({
      install_groups: [{ name: "default", install_ids: [installId], order: 0 }],
    }),
  });

  return { appId: app.id, branchId: branch.id };
}

export async function waitForComponentBuildActive(
  timeoutMs = 60_000,
): Promise<boolean> {
  const app = await findSeededApp();
  if (!app) return false;

  const comps = (await (
    await api(`/v1/apps/${app.id}/components`)
  ).json()) as { id: string }[];
  const componentId = comps?.[0]?.id;
  if (!componentId) return false;

  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    const res = (await (
      await api(`/v1/builds?component_id=${componentId}&limit=10`)
    ).json()) as
      | { data?: { status_v2?: { status?: string } }[] }
      | { status_v2?: { status?: string } }[];
    const builds = Array.isArray(res) ? res : (res.data ?? []);
    if (builds.some((b) => b?.status_v2?.status === "active")) return true;
    await new Promise((r) => setTimeout(r, 3_000));
  }
  return false;
}
