import { formatComponent, getConfig, getVcsConfig } from "./utils";

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

test("getConfig should take a raw grpc component config & return the correct config object", () => {
  const spec = getConfig({
    buildCfg: {
      externalImageCfg: {
        ociImageUrl: "test.io",
      },
    },
    deployCfg: {
      basic: {
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

test("formatComponent should return a GQL component with correct date format & vcs config field", () => {
  const spec = formatComponent(mockComponent);

  expect(spec).toMatchInlineSnapshot(
    `
    {
      "config": {
        "__typename": "ComponentConfig",
        "buildConfig": null,
        "deployConfig": null,
      },
      "createdAt": "1999-12-31T08:15:30.000",
      "id": "test-id",
      "name": "test-node",
      "updatedAt": "1999-12-31T08:15:30.000",
    }
  `
  );
});
