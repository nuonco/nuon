import { initBasicDeployConfig } from "./basic-deploy-config";

test("initBasicDeployConfig should return a gRPC message for a basic deploy config", () => {
  const spec = initBasicDeployConfig({
    envVars: [
      {
        key: "test",
        value: "test",
      },
      {
        key: "test",
        sensitive: true,
        value: "test",
      },
    ],
    healthCheckPath: "/test",
    instanceCount: 2,
    port: 3000,
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "argsList": [],
      "cpuLimit": "",
      "cpuRequest": "",
      "envVars": {
        "envList": [
          {
            "name": "test",
            "sensitive": false,
            "value": "test",
          },
          {
            "name": "test",
            "sensitive": true,
            "value": "test",
          },
        ],
      },
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
