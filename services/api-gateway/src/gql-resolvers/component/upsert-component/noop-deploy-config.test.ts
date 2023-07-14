import { initNoopDeployConfig } from "./noop-deploy-config";

test("initNoopDeployConfig should return gRPC message for a noop deploy", () => {
  const spec = initNoopDeployConfig();

  expect(spec.toObject()).toMatchInlineSnapshot(`{}`);
});
