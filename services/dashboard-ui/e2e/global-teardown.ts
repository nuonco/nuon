import { type FullConfig } from "@playwright/test";
import { env } from "./env";
import fs from "node:fs";

const ORG_STATE_PATH = "e2e/.auth/org.json";

let teardownStart = performance.now();

function log(message: string) {
  const elapsed = ((performance.now() - teardownStart) / 1000).toFixed(1);
  console.log(`[e2e teardown +${elapsed}s] ${message}`);
}

export default async function globalTeardown(_config: FullConfig) {
  teardownStart = performance.now();

  if (!fs.existsSync(ORG_STATE_PATH)) {
    log("no org state file found, skipping teardown");
    return;
  }

  const state = JSON.parse(fs.readFileSync(ORG_STATE_PATH, "utf-8")) as {
    orgId: string;
    createdOrg: boolean;
    apiToken?: string;
  };

  if (!state.createdOrg) {
    log(`org ${state.orgId} was not created by setup, skipping teardown`);
    return;
  }

  log(`deleting org ${state.orgId} (force)...`);

  // seed-user creates accounts as "seed@nuon.co", not the E2E_EMAIL.
  const adminEmail = "seed@nuon.co";

  const res = await fetch(
    `${env.adminApiUrl}/v1/orgs/${state.orgId}/admin-delete`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Nuon-Admin-Email": adminEmail,
      },
      body: JSON.stringify({ force: true }),
    },
  );

  if (!res.ok) {
    const body = await res.text();
    log(`failed to delete org ${state.orgId}: ${body}`);
  } else {
    log(`org ${state.orgId} deleted`);
  }

  log("teardown complete");
}
