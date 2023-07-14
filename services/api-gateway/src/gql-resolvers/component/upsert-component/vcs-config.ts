import { Config as VcsConfig } from "@buf/nuon_components.grpc_node/vcs/v1/config_pb";
import { ConnectedGithubConfig } from "@buf/nuon_components.grpc_node/vcs/v1/connected_github_pb";
import { PublicGitConfig } from "@buf/nuon_components.grpc_node/vcs/v1/public_git_pb";
import type { TgRPCMessage, VcsConfigInput } from "../../../types";

export function initVcsConfig(vcsInput: VcsConfigInput): TgRPCMessage {
  const vcsConfig = new VcsConfig();

  if (vcsInput?.connectedGithub) {
    const connectedGithubCfg = new ConnectedGithubConfig()
      .setRepo(vcsInput.connectedGithub.repo)
      .setDirectory(vcsInput.connectedGithub.directory);

    vcsConfig.setConnectedGithubConfig(connectedGithubCfg);
  } else if (vcsInput?.publicGit) {
    const publicGitCfg = new PublicGitConfig()
      .setRepo(vcsInput.publicGit.repo)
      .setDirectory(vcsInput.publicGit.directory);

    vcsConfig.setPublicGitConfig(publicGitCfg);
  }

  return vcsConfig;
}
