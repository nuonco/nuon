import { Buffer } from "node:buffer";
import { getInput, info, setFailed, setOutput, setSecret } from "@actions/core";
import ensureError from "ensure-error";
import isBase64 from "is-base64";
import { fetchInstallationToken } from "./fetch-token";

const run = async () => {
  try {
    const appId = getInput("app_id", { required: true });
    const installationId = Number(
      getInput("installation_id", { required: true })
    );
    const privateKeyInput = getInput("private_key", { required: true });
    const privateKey = isBase64(privateKeyInput)
      ? Buffer.from(privateKeyInput, "base64").toString("utf8")
      : privateKeyInput;

    // const repositoriesInput = getInput("repositories");

    const installationToken = await fetchInstallationToken({
      appId,
      installationId,
      privateKey,
      // repos,
    });

    setSecret(installationToken);
    setOutput("token", installationToken);
    info("Token generated successfully!");
  } catch (_error: unknown) {
    const error = ensureError(_error);
    setFailed(error);
  }
};

void run();
