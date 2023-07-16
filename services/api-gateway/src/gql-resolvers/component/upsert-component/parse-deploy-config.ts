import { Config as DeployConfig } from "../../../build/components/deploy/v1/config_pb";
import type { DeployConfigInput, TgRPCMessage } from "../../../types";
import { initBasicDeployConfig } from "./basic-deploy-config";
import { initHelmDeployConfig } from "./helm-deploy-config";
import { initNoopDeployConfig } from "./noop-deploy-config";
import { initTerraformDeployConfig } from "./terraform-deploy-config";

export function parseDeployConfigInput(
  deployConfigInput: DeployConfigInput
): TgRPCMessage {
  const deployCfg = new DeployConfig();

  if (deployConfigInput?.noop) {
    const noopCfg = initNoopDeployConfig();

    deployCfg.setNoop(noopCfg);
  }

  if (deployConfigInput?.basicDeployConfig) {
    const basicDeployCfg = initBasicDeployConfig(
      deployConfigInput.basicDeployConfig
    );

    deployCfg.setBasic(basicDeployCfg);
  }

  if (deployConfigInput?.helmDeployConfig) {
    const helmDeployCfg = initHelmDeployConfig(
      deployConfigInput.helmDeployConfig
    );

    deployCfg.setHelmChart(helmDeployCfg);
  }

  if (deployConfigInput?.terraformDeployConfig) {
    const terraformDeployCfg = initTerraformDeployConfig(
      deployConfigInput.terraformDeployConfig
    );

    deployCfg.setTerraformModuleConfig(terraformDeployCfg);
  }

  return deployCfg;
}
