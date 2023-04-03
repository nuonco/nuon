import { AwsRegion } from "../../types";
import {
  parseBuildConfigInput,
  parseConfigInput,
  parseDeployConfigInput,
  upsertComponent,
} from "./upsert-component";

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
      "noop": undefined,
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
      "noop": undefined,
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
      "noop": undefined,
      "timeout": undefined,
    }
  `);
});

test("parseBuildConfigInput should return a build config for a docker build with public git repo", () => {
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
      "noop": undefined,
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
        "instanceCount": 1,
        "listenerCfg": {
          "healthCheckPath": "/test",
          "listenPort": 3000,
        },
        "memLimit": "",
        "memRequest": "",
      },
      "helmRepo": undefined,
      "timeout": undefined,
    }
  `);
});

test("parseDeployConfigInput should return a deploy config for a helm repo deployment", () => {
  const spec = parseDeployConfigInput({
    helmRepoDeployConfig: {
      chartName: "httpbin",
      chartRepo: "rgnu/httpbin",
      chartVersion: "1.0.0",
      imageRepoValuesKey: "httpbin/blah",
      imageTagValuesKey: "latest",
    },
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "basic": undefined,
      "helmRepo": {
        "chartName": "httpbin",
        "chartRepo": "rgnu/httpbin",
        "chartVersion": "1.0.0",
        "imageRepoValuesKey": "httpbin/blah",
        "imageTagValuesKey": "latest",
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
        "noop": undefined,
        "timeout": undefined,
      },
      "deployCfg": {
        "basic": {
          "argsList": [],
          "cpuLimit": "",
          "cpuRequest": "",
          "instanceCount": 2,
          "listenerCfg": {
            "healthCheckPath": "/test",
            "listenPort": 8080,
          },
          "memLimit": "",
          "memRequest": "",
        },
        "helmRepo": undefined,
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
