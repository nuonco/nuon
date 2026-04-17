import { type FullConfig } from "@playwright/test";
import { env } from "./env";
import fs from "node:fs";

const ORG_STATE_PATH = "e2e/.auth/org.json";

export default async function globalTeardown(_config: FullConfig) {
  if (!fs.existsSync(ORG_STATE_PATH)) return;

  const state = JSON.parse(fs.readFileSync(ORG_STATE_PATH, "utf-8")) as {
    orgId: string;
    createdOrg: boolean;
  };

  if (!state.createdOrg) return;

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
