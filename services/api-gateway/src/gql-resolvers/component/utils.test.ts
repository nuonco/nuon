import {
  formatComponent,
  getConfig,
  getEnvVarsConfig,
  getVcsConfig,
} from "./utils";

const mockDateTimeObject = {
  day: 31,
  hours: 8,
  minutes: 15,
  month: 12,
  seconds: 30,
  utcOffset: {
    nanos: 0,
    seconds: 0,
  },
  year: 1999,
};

const mockComponent = {
  createdAt: mockDateTimeObject,
  id: "test-id",
  name: "test-node",
  updatedAt: mockDateTimeObject,
};

test("getVcsConfig should take a raw grpc component & return the correct vcs config object for public git repos", () => {
  const spec = getVcsConfig({
    vcsCfg: {
      publicGitConfig: {
        directory: "/",
        repo: "test-repo",
      },
    },
  });

  expect(spec).toMatchInlineSnapshot(`
    {
      "__typename": "PublicGitConfig",
      "directory": "/",
      "repo": "test-repo",
    }
  `);
});

test("getVcsConfig should take a raw grpc component & return the correct vcs config object for connected github repos", () => {
  const spec = getVcsConfig({
    vcsCfg: {
      connectedGithubConfig: {
        branch: "main",
        directory: "/",
        repo: "test-repo",
      },
    },
  });

  expect(spec).toMatchInlineSnapshot(`
    {
      "__typename": "ConnectedGithubConfig",
      "branch": "main",
      "directory": "/",
      "repo": "test-repo",
    }
  `);
});

test("getEnvVarsConfig should take a raw grpc component & return the correct environment variables config object", () => {
  const spec = getEnvVarsConfig({
    envVarsConfig: {
      key: "test-key",
      value: "test-value",
    },
  });

  expect(spec).toMatchInlineSnapshot(`
      {
        "__typename": "KeyValuePair",
        "key": "test-key",
        "value": "test-value",
      }
    `);
});

test("getConfig should take a raw grpc component config & return the correct config object", () => {
  const spec = getConfig({
    buildCfg: {
      externalImageCfg: {
        ociImageUrl: "test.io",
      },
    },
    deployCfg: {
      basic: {
        envVars: {
          envList: [
            {
              name: "test",
              sensitive: false,
              value: "test",
            },
            {
              name: "test",
              sensitive: true,
              value: "test",
            },
          ],
        },
        instanceCount: 2,
        listenerCfg: {
          healthCheckPath: "/readyz",
          port: 8080,
        },
      },
    },
  });

  expect(spec).toMatchInlineSnapshot(`
    {
      "__typename": "ComponentConfig",
      "buildConfig": {
        "__typename": "ExternalImageConfig",
        "authConfig": null,
        "ociImageUrl": "test.io",
      },
      "deployConfig": {
        "__typename": "BasicDeployConfig",
        "envVars": [
          {
            "__typename": "KeyValuePair",
            "key": "test",
            "sensitive": false,
            "value": "test",
          },
          {
            "__typename": "KeyValuePair",
            "key": "test",
            "sensitive": true,
            "value": "test",
          },
        ],
        "healthCheckPath": "/readyz",
        "instanceCount": 2,
        "port": undefined,
      },
    }
  `);
});

test("getConfig should take a raw grpc component config & return the correct config object - helm", () => {
  const spec = getConfig({
    buildCfg: {
      externalImageCfg: {
        ociImageUrl: "test.io",
      },
    },
    deployCfg: {
      helmRepo: {
        chartName: "httpbin",
        chartRepo: "httpbin/blah",
        chartVersion: "latest-version",
        imageRepoValuesKey: "repo.com",
        imageTagValuesKey: "latest-tag",
      },
    },
  });

  expect(spec).toMatchInlineSnapshot(`
    {
      "__typename": "ComponentConfig",
      "buildConfig": {
        "__typename": "ExternalImageConfig",
        "authConfig": null,
        "ociImageUrl": "test.io",
      },
      "deployConfig": {
        "__typename": "HelmRepoDeployConfig",
        "chartName": "httpbin",
        "chartRepo": "httpbin/blah",
        "chartVersion": "latest-version",
        "imageRepoValuesKey": "repo.com",
        "imageTagValuesKey": "latest-tag",
      },
    }
  `);
});

test("getConfig should take a raw grpc component config & return the correct config object - helm chart", () => {
  const spec = getConfig({
    buildCfg: {
      helmChartCfg: {
        chartName: "test-chart",
        envVarsConfig: {
          key: "test-env-var-key",
          val: "test-env-var-val",
        },
        vcsCfg: {
          connectedGithubConfig: {
            branch: "main",
            directory: "/",
            repo: "test-repo",
          },
        },
      },
    },
    deployCfg: {
      helmChart: {
        values: {
          valuesList: [
            {
              name: "key",
              sensitive: false,
              value: "value",
            },
            {
              name: "key",
              sensitive: true,
              value: "value",
            },
          ],
        },
      },
    },
  });

  expect(spec).toMatchInlineSnapshot(`
    {
      "__typename": "ComponentConfig",
      "buildConfig": {
        "__typename": "HelmBuildConfig",
        "chartName": "test-chart",
        "envVarsConfig": {
          "__typename": "KeyValuePair",
          "key": "test-env-var-key",
          "val": "test-env-var-val",
        },
        "vcsConfig": {
          "__typename": "ConnectedGithubConfig",
          "branch": "main",
          "directory": "/",
          "repo": "test-repo",
        },
      },
      "deployConfig": {
        "__typename": "HelmDeployConfig",
        "noop": true,
        "values": [
          {
            "__typename": "KeyValuePair",
            "key": "key",
            "sensitive": false,
            "value": "value",
          },
          {
            "__typename": "KeyValuePair",
            "key": "key",
            "sensitive": true,
            "value": "value",
          },
        ],
      },
    }
  `);
});

test("getConfig should take a raw grpc component config & return the correct config object - terraform", () => {
  const spec = getConfig({
    buildCfg: {
      terraformModuleCfg: {
        vcsCfg: {
          connectedGithubConfig: {
            branch: "main",
            directory: "/",
            repo: "test-repo",
          },
        },
      },
    },
    deployCfg: {
      terraformModuleConfig: {
        terraformVersion: 1,
        vars: {
          variablesList: [
            {
              key: "test-env-var-key",
              val: "test-env-var-val",
            },
          ],
        },
      },
    },
  });

  expect(spec).toMatchInlineSnapshot(`
    {
      "__typename": "ComponentConfig",
      "buildConfig": {
        "__typename": "TerraformBuildConfig",
        "vcsConfig": {
          "__typename": "ConnectedGithubConfig",
          "branch": "main",
          "directory": "/",
          "repo": "test-repo",
        },
      },
      "deployConfig": {
        "__typename": "TerraformDeployConfig",
        "terraformVersion": "TERRAFORM_VERSION_LATEST",
        "vars": [
          {
            "__typename": "KeyValuePair",
            "key": undefined,
            "sensitive": undefined,
            "value": undefined,
          },
        ],
      },
    }
  `);
});

test("formatComponent should return a GQL component with correct date format & vcs config field", () => {
  const spec = formatComponent(mockComponent);

  expect(spec).toMatchInlineSnapshot(`
    {
      "config": {
        "__typename": "ComponentConfig",
        "buildConfig": null,
        "deployConfig": null,
      },
      "createdAt": "1999-12-31T08:15:30.000Z",
      "id": "test-id",
      "name": "test-node",
      "updatedAt": "1999-12-31T08:15:30.000Z",
    }
  `);
});
