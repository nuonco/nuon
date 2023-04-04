import { env } from "node:process";
import { createAppAuth } from "@octokit/auth-app";
import { request } from "@octokit/request";

export const fetchInstallationToken = async ({
  appId,
  installationId,
  privateKey,
}: Readonly<{
  appId: string;
  installationId: number;
  privateKey: string;
}>): Promise<string> => {
  const app = createAppAuth({
    appId,
    privateKey,
    request: request.defaults({
      // GITHUB_API_URL is part of GitHub Actions' built-in environment variables.
      // See https://docs.github.com/en/actions/reference/environment-variables#default-environment-variables.
      baseUrl: env.GITHUB_API_URL,
    }),
  });

  const installation = await app({
    installationId,
    type: "installation",
  });
  return installation.token;
};
