import { initBasicDeployConfig } from "./basic-deploy-config";

test("initBasicDeployConfig should return a gRPC message for a basic deploy config", () => {
  const spec = initBasicDeployConfig({
    healthCheckPath: "/test",
    instanceCount: 2,
    port: 3000,
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "argsList": [],
      "cpuLimit": "",
      "cpuRequest": "",
      "envVars": undefined,
      "instanceCount": 2,
      "listenerCfg": {
        "healthCheckPath": "/test",
        "listenPort": 3000,
      },
      "memLimit": "",
      "memRequest": "",
    }
  `);
});
