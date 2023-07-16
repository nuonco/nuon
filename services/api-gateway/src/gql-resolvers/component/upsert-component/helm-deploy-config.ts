import { HelmChartConfig as HelmDeployConfig } from "../../../build/components/deploy/v1/helm_chart_pb";
import {
  HelmValue,
  HelmValues,
} from "../../../build/components/variables/v1/helm_pb";
import { HelmDeployInput, TgRPCMessage } from "../../../types";

export function initHelmDeployConfig(
  helmDeployInput: HelmDeployInput
): TgRPCMessage {
  const helmDeployCfg = new HelmDeployConfig();

  if (helmDeployInput?.values) {
    const { values } = helmDeployInput;
    const helmValues = new HelmValues().setValuesList(
      values.map(({ key, sensitive, value }) => {
        return new HelmValue()
          .setName(key)
          .setValue(value)
          .setSensitive(sensitive);
      })
    );

    helmDeployCfg.setValues(helmValues);
  }

  return helmDeployCfg;
}
