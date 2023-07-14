import { initNoopBuildConfig } from "./noop-build-config";

test("initNoopBuildConfig should return gRPC message for a noop build", () => {
  const spec = initNoopBuildConfig();

  expect(spec.toObject()).toMatchInlineSnapshot(`{}`);
});
