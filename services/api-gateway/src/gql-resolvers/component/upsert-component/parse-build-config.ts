import { Config as BuildConfig } from "../../../build/components/build/v1/build_pb";
import type { BuildConfigInput, TgRPCMessage } from "../../../types";
import { initDockerBuildConfig } from "./docker-build-config";
import { initExternalImageConfig } from "./external-image-config";
import { initHelmBuildConfig } from "./helm-build-config";
import { initNoopBuildConfig } from "./noop-build-config";
import { initTerraformBuildConfig } from "./terraform-build-config";

export function parseBuildConfigInput(
  buildConfigInput: BuildConfigInput
): TgRPCMessage {
  const buildCfg = new BuildConfig();

  if (buildConfigInput?.noop) {
    const noopCfg = initNoopBuildConfig();

    buildCfg.setNoop(noopCfg);
  }

  if (buildConfigInput?.externalImageConfig) {
    const externalImageCfg = initExternalImageConfig(
      buildConfigInput.externalImageConfig
    );

    buildCfg.setExternalImageCfg(externalImageCfg);
  }

  if (buildConfigInput?.dockerBuildConfig) {
    const dockerCfg = initDockerBuildConfig(buildConfigInput.dockerBuildConfig);

    buildCfg.setDockerCfg(dockerCfg);
  }

  if (buildConfigInput?.helmBuildConfig) {
    const helmBuildCfg = initHelmBuildConfig(buildConfigInput.helmBuildConfig);

    buildCfg.setHelmChartCfg(helmBuildCfg);
  }

  if (buildConfigInput?.terraformBuildConfig) {
    const terraformBuildCfg = initTerraformBuildConfig(
      buildConfigInput.terraformBuildConfig
    );

    buildCfg.setTerraformModuleCfg(terraformBuildCfg);
  }

  return buildCfg;
}
