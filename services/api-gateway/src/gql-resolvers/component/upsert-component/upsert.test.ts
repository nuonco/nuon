import { AwsRegion, TerraformVersion } from "../../../types";
import {
  parseBuildConfigInput,
  parseConfigInput,
  parseDeployConfigInput,
  upsertComponent,
} from "./upsert";

test("parseBuildConfigInput should return a build config for an external image", () => {
  const spec = parseBuildConfigInput({
    externalImageConfig: {
      ociImageUrl: "some-place.io/test/image",
      tag: "latest",
    },
  });
  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "dockerCfg": undefined,
      "externalImageCfg": {
        "authCfg": {
          "awsIamAuthCfg": undefined,
          "publicAuthCfg": {},
        },
        "ociImageUrl": "some-place.io/test/image",
        "tag": "latest",
      },
      "helmChartCfg": undefined,
      "noop": undefined,
      "terraformModuleCfg": undefined,
      "timeout": undefined,
    }
  `);
});

test("parseBuildConfigInput should return a build config for a private ECR external image", () => {
  const spec = parseBuildConfigInput({
    externalImageConfig: {
      authConfig: {
        region: AwsRegion.UsEast_1,
        role: "test",
      },
      ociImageUrl: "some-place.io/test/image",
      tag: "latest",
    },
  });
  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "dockerCfg": undefined,
      "externalImageCfg": {
        "authCfg": {
          "awsIamAuthCfg": {
            "awsRegion": "US_EAST_1",
            "iamRoleArn": "test",
          },
          "publicAuthCfg": undefined,
        },
        "ociImageUrl": "some-place.io/test/image",
        "tag": "latest",
      },
      "helmChartCfg": undefined,
      "noop": undefined,
      "terraformModuleCfg": undefined,
      "timeout": undefined,
    }
  `);
});

test("parseBuildConfigInput should return a build config for a docker build with connected github repo", () => {
  const spec = parseBuildConfigInput({
    dockerBuildConfig: {
      dockerfile: "Dockerfile",
      vcsConfig: {
        connectedGithub: {
          branch: "main",
          directory: "/",
          repo: "org/repo",
        },
      },
    },
  });
  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "dockerCfg": {
        "buildArgsList": [],
        "dockerfile": "Dockerfile",
        "envVars": undefined,
        "target": "",
        "vcsCfg": {
          "connectedGithubConfig": {
            "directory": "/",
            "gitRef": "",
            "githubAppKeyId": "",
            "githubAppKeySecretName": "",
            "githubInstallId": "",
            "repo": "org/repo",
          },
          "publicGitConfig": undefined,
        },
      },
      "externalImageCfg": undefined,
      "helmChartCfg": undefined,
      "noop": undefined,
      "terraformModuleCfg": undefined,
      "timeout": undefined,
    }
  `);
});

test("parseBuildConfigInput should return a build config for a docker build with public git repo and one environment variable", () => {
  const spec = parseBuildConfigInput({
    dockerBuildConfig: {
      dockerfile: "Dockerfile",
      vcsConfig: {
        publicGit: {
          directory: "/",
          repo: "org/repo",
        },
      },
    },
  });
  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "dockerCfg": {
        "buildArgsList": [],
        "dockerfile": "Dockerfile",
        "envVars": undefined,
        "target": "",
        "vcsCfg": {
          "connectedGithubConfig": undefined,
          "publicGitConfig": {
            "directory": "/",
            "gitRef": "",
            "repo": "org/repo",
          },
        },
      },
      "externalImageCfg": undefined,
      "helmChartCfg": undefined,
      "noop": undefined,
      "terraformModuleCfg": undefined,
      "timeout": undefined,
    }
  `);
});

test("parseBuildConfigInput should return a build config for a helm chart build with public git repo", () => {
  const spec = parseBuildConfigInput({
    helmBuildConfig: {
      chartName: "test-chart",
      vcsConfig: {
        publicGit: {
          directory: "/",
          repo: "org/repo",
        },
      },
    },
  });
  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "dockerCfg": undefined,
      "externalImageCfg": undefined,
      "helmChartCfg": {
        "chartName": "test-chart",
        "vcsCfg": {
          "connectedGithubConfig": undefined,
          "publicGitConfig": {
            "directory": "/",
            "gitRef": "",
            "repo": "org/repo",
          },
        },
      },
      "noop": undefined,
      "terraformModuleCfg": undefined,
      "timeout": undefined,
    }
  `);
});

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

test("parseConfigInput should return a component config", () => {
  const spec = parseConfigInput({
    buildConfig: {
      externalImageConfig: {
        ociImageUrl: "test.io/test/iagme",
      },
    },
    deployConfig: {
      basicDeployConfig: {
        healthCheckPath: "/test",
        instanceCount: 2,
        port: 8080,
      },
    },
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "buildCfg": {
        "dockerCfg": undefined,
        "externalImageCfg": {
          "authCfg": {
            "awsIamAuthCfg": undefined,
            "publicAuthCfg": {},
          },
          "ociImageUrl": "test.io/test/iagme",
          "tag": "",
        },
        "helmChartCfg": undefined,
        "noop": undefined,
        "terraformModuleCfg": undefined,
        "timeout": undefined,
      },
      "connections": undefined,
      "deployCfg": {
        "basic": {
          "argsList": [],
          "cpuLimit": "",
          "cpuRequest": "",
          "envVars": undefined,
          "instanceCount": 2,
          "listenerCfg": {
            "healthCheckPath": "/test",
            "listenPort": 8080,
          },
          "memLimit": "",
          "memRequest": "",
        },
        "helmChart": undefined,
        "helmRepo": undefined,
        "noop": undefined,
        "terraformModuleConfig": undefined,
        "timeout": undefined,
      },
      "id": "",
    }
  `);
});

test("upsertComponent resolver should return error if service client doesn't exist", async () => {
  await expect(
    upsertComponent(undefined, { input: {} }, { clients: {} })
  ).rejects.toThrowErrorMatchingInlineSnapshot(`"Service isn't available"`);
});
