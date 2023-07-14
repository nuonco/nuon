import { parseBuildConfigInput } from "./parse-build-config";

test("parseBuildConfigInput should return a gRPC message with a noop config", () => {
  const spec = parseBuildConfigInput({
    noop: true,
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "dockerCfg": undefined,
      "externalImageCfg": undefined,
      "helmChartCfg": undefined,
      "noop": {},
      "terraformModuleCfg": undefined,
      "timeout": undefined,
    }
  `);
});

test("parseBuildConfigInput should return a gRPC message with an external image config", () => {
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

test("parseBuildConfigInput should return a gRPC message with a docker build config", () => {
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

test("parseBuildConfigInput should return a gRPC message with a helm build config", () => {
  const spec = parseBuildConfigInput({
    helmBuildConfig: {
      chartName: "test-chart",
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
      "dockerCfg": undefined,
      "externalImageCfg": undefined,
      "helmChartCfg": {
        "chartName": "test-chart",
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
      "noop": undefined,
      "terraformModuleCfg": undefined,
      "timeout": undefined,
    }
  `);
});

test("parseBuildConfigInput should return a gRPC message with a terraform build config", () => {
  const spec = parseBuildConfigInput({
    terraformBuildConfig: {
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
      "dockerCfg": undefined,
      "externalImageCfg": undefined,
      "helmChartCfg": undefined,
      "noop": undefined,
      "terraformModuleCfg": {
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
      "timeout": undefined,
    }
  `);
});
