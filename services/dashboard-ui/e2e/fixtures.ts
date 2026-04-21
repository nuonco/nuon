import { test as base, expect } from "@playwright/test";
import fs from "node:fs";

const ORG_STATE_PATH = "e2e/.auth/org.json";

function getOrgState(): { orgId: string; appConfig?: string } {
  return JSON.parse(fs.readFileSync(ORG_STATE_PATH, "utf-8"));
}

type Fixtures = {
  orgId: string;
  appConfig: string | undefined;
};

export const test = base.extend<Fixtures>({
  orgId: async ({}, use) => {
    await use(getOrgState().orgId);
  },
  appConfig: async ({}, use) => {
    await use(getOrgState().appConfig);
  },
});

export { expect };
