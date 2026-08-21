import { GlobalRegistrator } from "@happy-dom/global-registrator";
import { mock } from "bun:test";

GlobalRegistrator.register();

import "@testing-library/jest-dom";

// posthog-js runs DOM access as an import side-effect (it async-loads its
// toolbar/surveys chunks and reads window.location.hash), which throws an
// unhandled rejection under the test DOM. Stub it so the real module never loads.
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
