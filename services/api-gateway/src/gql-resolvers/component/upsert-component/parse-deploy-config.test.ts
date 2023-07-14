import { TerraformVersion } from "../../../types";
import { parseDeployConfigInput } from "./parse-deploy-config";

test("parseDeployConfigInput should return a deploy config for a basic k8s deployment", () => {
  const spec = parseDeployConfigInput({
    basicDeployConfig: {
      healthCheckPath: "/test",
      instanceCount: 1,
      port: 3000,
    },
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "basic": {
        "argsList": [],
        "cpuLimit": "",
        "cpuRequest": "",
        "envVars": undefined,
        "instanceCount": 1,
        "listenerCfg": {
          "healthCheckPath": "/test",
          "listenPort": 3000,
        },
        "memLimit": "",
        "memRequest": "",
      },
      "helmChart": undefined,
      "helmRepo": undefined,
      "noop": undefined,
      "terraformModuleConfig": undefined,
      "timeout": undefined,
    }
  `);
});

test("parseDeployConfigInput should return a deploy config for a helm chart deployment", () => {
  const spec = parseDeployConfigInput({
    helmDeployConfig: {
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
    },
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "basic": undefined,
      "helmChart": {
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
      },
      "helmRepo": undefined,
      "noop": undefined,
      "terraformModuleConfig": undefined,
      "timeout": undefined,
    }
  `);
});

test("parseDeployConfigInput should return a deploy config for a terraform deployment", () => {
  const spec = parseDeployConfigInput({
    terraformDeployConfig: {
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
    },
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "basic": undefined,
      "helmChart": undefined,
      "helmRepo": undefined,
      "noop": undefined,
      "terraformModuleConfig": {
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
      },
      "timeout": undefined,
    }
  `);
});
