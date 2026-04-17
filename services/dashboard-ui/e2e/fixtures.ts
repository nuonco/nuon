import { test as base, expect } from "@playwright/test";
import fs from "node:fs";

const ORG_STATE_PATH = "e2e/.auth/org.json";

function getOrgId(): string {
  const state = JSON.parse(fs.readFileSync(ORG_STATE_PATH, "utf-8")) as {
    orgId: string;
  };
  return state.orgId;
}

type Fixtures = {
  orgId: string;
};

export const test = base.extend<Fixtures>({
  orgId: async ({}, use) => {
    await use(getOrgId());
  },
});

export { expect };
