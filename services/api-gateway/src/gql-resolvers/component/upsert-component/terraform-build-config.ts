import { TerraformModuleConfig as TerraformBuildConfig } from "@buf/nuon_components.grpc_node/build/v1/terraform_module_pb";
import type { TerraformBuildInput, TgRPCMessage } from "../../../types";
import { initVcsConfig } from "./vcs-config";

export function initTerraformBuildConfig(
  terraformBuildInput: TerraformBuildInput
): TgRPCMessage {
  const { vcsConfig: vcsInput } = terraformBuildInput;
  const vcsConfig = initVcsConfig(vcsInput);

  return new TerraformBuildConfig().setVcsCfg(vcsConfig);
}
