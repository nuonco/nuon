import { GlobalRegistrator } from "@happy-dom/global-registrator";
import { afterEach, mock } from "bun:test";

GlobalRegistrator.register();

import "@testing-library/jest-dom";

const { cleanup } = await import("@testing-library/react");

afterEach(cleanup);

const posthogStub = new Proxy(
  {},
  {
    get: () => () => undefined,
  }
);

mock.module("posthog-js", () => ({
  default: posthogStub,
  posthog: posthogStub,
}));
