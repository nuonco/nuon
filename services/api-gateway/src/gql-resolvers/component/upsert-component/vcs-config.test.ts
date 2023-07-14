import { initVcsConfig } from "./vcs-config";

test("initVcsConfig should return a gRPC message for a connect github config", () => {
  const spec = initVcsConfig({
    connectedGithub: {
      branch: "main",
      directory: "/",
      repo: "org/repo",
    },
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "connectedGithubConfig": {
        "directory": "/",
        "gitRef": "",
        "githubAppKeyId": "",
        "githubAppKeySecretName": "",
        "githubInstallId": "",
        "repo": "org/repo",
      },
      "publicGitConfig": undefined,
    }
  `);
});

test("initVcsConfig should return a gRPC message for a public git config", () => {
  const spec = initVcsConfig({
    publicGit: {
      directory: "/",
      repo: "org/repo",
    },
  });

  expect(spec.toObject()).toMatchInlineSnapshot(`
    {
      "connectedGithubConfig": undefined,
      "publicGitConfig": {
        "directory": "/",
        "gitRef": "",
        "repo": "org/repo",
      },
    }
  `);
});
