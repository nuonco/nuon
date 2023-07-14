import { initHelmBuildConfig } from "./helm-build-config";

test("initHelmBuildConfig should return a gRPC message for a helm build with a public git vcs config", () => {
  const spec = initHelmBuildConfig({
    chartName: "test-chart",
    vcsConfig: {
      publicGit: {
        directory: "/",
        repo: "org/repo",
      },
    },
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "chartName": "test-chart",
      "vcsCfg": {
        "connectedGithubConfig": undefined,
        "publicGitConfig": {
          "directory": "/",
          "gitRef": "",
          "repo": "org/repo",
        },
      },
    }
  `);
});

test("initHelmBuildConfig should return a gRPC message for a helm build with a connected github vcs config", () => {
  const spec = initHelmBuildConfig({
    chartName: "test-chart",
    vcsConfig: {
      connectedGithub: {
        branch: "main",
        directory: "/",
        repo: "org/repo",
      },
    },
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
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
    }
  `);
});
