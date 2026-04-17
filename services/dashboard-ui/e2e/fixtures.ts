import { test as base, expect } from "@playwright/test";
import { env } from "./env";

type Fixtures = {
  orgId: string;
};

export const test = base.extend<Fixtures>({
  orgId: async ({}, use) => {
    await use(env.orgId);
  },
});

export { expect };
