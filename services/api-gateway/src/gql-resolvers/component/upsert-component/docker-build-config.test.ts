import { initDockerBuildConfig } from "./docker-build-config";

test("initDockerBuildConfig should return a gRPC message for a docker build with a public git vcs config", () => {
  const spec = initDockerBuildConfig({
    dockerfile: "Dockerfile",
    vcsConfig: {
      publicGit: {
        directory: "/",
        repo: "org/repo",
      },
    },
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
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
    }
  `);
});

test("initDockerBuildConfig should return a gRPC message for a docker build with a connected github vcs config", () => {
  const spec = initDockerBuildConfig({
    dockerfile: "Dockerfile",
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
    }
  `);
});
