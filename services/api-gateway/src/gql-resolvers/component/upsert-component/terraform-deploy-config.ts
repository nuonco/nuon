import { TerraformModuleConfig as TerraformDeployConfig } from "@buf/nuon_components.grpc_node/deploy/v1/terraform_module_pb";
import {
  TerraformVariable,
  TerraformVariables,
} from "@buf/nuon_components.grpc_node/variables/v1/terraform_pb";
import type { TerraformDeployConfigInput, TgRPCMessage } from "../../../types";

export function initTerraformDeployConfig(
  terraformDeployInput: TerraformDeployConfigInput
): TgRPCMessage {
  const terraformDeployCfg = new TerraformDeployConfig().setTerraformVersion(
    1 // terraformDeployInput?.terraformVersion
  );

  if (terraformDeployInput?.vars) {
    const { vars } = terraformDeployInput;
    const terraformVars = new TerraformVariables().setVariablesList(
      vars.map(({ key, sensitive, value }) => {
        return new TerraformVariable()
          .setName(key)
          .setValue(value)
          .setSensitive(sensitive);
      })
    );

    terraformDeployCfg.setVars(terraformVars);
  }

  return terraformDeployCfg;
}
