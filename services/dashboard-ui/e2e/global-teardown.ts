import { type FullConfig } from "@playwright/test";
import { env } from "./env";
import fs from "node:fs";

const ORG_STATE_PATH = "e2e/.auth/org.json";

async function getToken(): Promise<string> {
  const res = await fetch(`${env.adminApiUrl}/v1/general/seed-user`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-Nuon-Admin-Email": env.email,
    },
    body: JSON.stringify({}),
  });
  const body = await res.text();
  const firstJson = body.match(/^\{[^}]*\}/);
  if (!firstJson) throw new Error(`seed-user unexpected response: ${body}`);
  const { api_token } = JSON.parse(firstJson[0]) as { api_token: string };
  return api_token;
}

async function deleteApps(token: string, orgId: string) {
  const listRes = await fetch(`${env.publicApiUrl}/v1/apps`, {
    headers: {
      Authorization: `Bearer ${token}`,
      "X-Nuon-Org-ID": orgId,
    },
  });
  if (!listRes.ok) return;

  const apps = (await listRes.json()) as { id: string }[];
  for (const app of apps) {
    const delRes = await fetch(`${env.publicApiUrl}/v1/apps/${app.id}`, {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${token}`,
        "X-Nuon-Org-ID": orgId,
      },
    });
    if (!delRes.ok) {
      const body = await delRes.text();
      console.warn(`Failed to delete app ${app.id}: ${body}`);
    }
  }
}

export default async function globalTeardown(_config: FullConfig) {
  if (!fs.existsSync(ORG_STATE_PATH)) return;

  const state = JSON.parse(fs.readFileSync(ORG_STATE_PATH, "utf-8")) as {
    orgId: string;
    createdOrg: boolean;
  };

  if (!state.createdOrg) return;

  const token = await getToken();
  await deleteApps(token, state.orgId);

  const res = await fetch(
    `${env.adminApiUrl}/v1/orgs/${state.orgId}/admin-delete`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Nuon-Admin-Email": env.email,
      },
      body: JSON.stringify({}),
    },
  );

  if (!res.ok) {
    const body = await res.text();
    console.warn(`Failed to delete test org ${state.orgId}: ${body}`);
  }
}
