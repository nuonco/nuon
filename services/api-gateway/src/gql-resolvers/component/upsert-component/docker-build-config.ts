import { DockerConfig } from "../../../build/components/build/v1/docker_pb";
import type { DockerBuildInput, TgRPCMessage } from "../../../types";
import { initVcsConfig } from "./vcs-config";

export function initDockerBuildConfig(
  dockerBuildInput: DockerBuildInput
): TgRPCMessage {
  const { dockerfile, vcsConfig: vcsInput } = dockerBuildInput;
  const vcsConfig = initVcsConfig(vcsInput);

  return new DockerConfig().setDockerfile(dockerfile).setVcsCfg(vcsConfig);
}
