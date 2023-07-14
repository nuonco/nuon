import { initTerraformBuildConfig } from "./terraform-build-config";

test("initTerraformBuildConfig should return a gRPC message for a terraform build with a public git vcs config", () => {
  const spec = initTerraformBuildConfig({
    vcsConfig: {
      publicGit: {
        directory: "/",
        repo: "org/repo",
      },
    },
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
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

test("initTerraformBuildConfig should return a gRPC message for a terraform build with a connected github vcs config", () => {
  const spec = initTerraformBuildConfig({
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
