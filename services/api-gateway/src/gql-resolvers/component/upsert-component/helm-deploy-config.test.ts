import { initHelmDeployConfig } from "./helm-deploy-config";

test("initHelmDeployConfig should return gRPC message for a helm deploy config", () => {
  const spec = initHelmDeployConfig({
    noop: true,
    values: [
      {
        key: "key",
        sensitive: false,
        value: "value",
      },
      {
        key: "key",
        sensitive: true,
        value: "value",
      },
    ],
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "values": {
        "valuesList": [
          {
            "name": "key",
            "sensitive": false,
            "value": "value",
          },
          {
            "name": "key",
            "sensitive": true,
            "value": "value",
          },
        ],
      },
    }
  `);
});
