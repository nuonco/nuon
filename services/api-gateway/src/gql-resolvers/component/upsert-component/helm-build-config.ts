import { HelmChartConfig } from "../../../build/components/build/v1/helm_chart_pb";
import type { HelmBuildInput, TgRPCMessage } from "../../../types";
import { initVcsConfig } from "./vcs-config";

export function initHelmBuildConfig(
  helmBuildInput: HelmBuildInput
): TgRPCMessage {
  const { chartName, vcsConfig: vcsInput } = helmBuildInput;
  const vcsConfig = initVcsConfig(vcsInput);

  return new HelmChartConfig().setChartName(chartName).setVcsCfg(vcsConfig);
}
