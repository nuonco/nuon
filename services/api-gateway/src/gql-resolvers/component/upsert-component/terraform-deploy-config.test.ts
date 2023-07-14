import { TerraformVersion } from "../../../types";
import { initTerraformDeployConfig } from "./terraform-deploy-config";

test("initTerraformDeployConfig should return a gRPC message for a terraform deploy config", () => {
  const spec = initTerraformDeployConfig({
    terraformVersion: TerraformVersion.TerraformVersionLatest,
    vars: [
      {
        key: "TEST",
        value: "test",
      },
      {
        key: "TEST",
        sensitive: true,
        value: "test",
      },
    ],
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "terraformVersion": 1,
      "vars": {
        "variablesList": [
          {
            "name": "TEST",
            "sensitive": false,
            "value": "test",
          },
          {
            "name": "TEST",
            "sensitive": true,
            "value": "test",
          },
        ],
      },
    }
  `);
});
