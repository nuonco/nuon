import { BasicConfig as BasicDeployConfig } from "../../../build/components/deploy/v1/basic_pb";
import { ListenerConfig } from "../../../build/components/deploy/v1/config_pb";
import { EnvVar, EnvVars } from "../../../build/components/variables/v1/env_pb";
import type { BasicDeployConfigInput, TgRPCMessage } from "../../../types";

export function initBasicDeployConfig(
  basicDeployInput: BasicDeployConfigInput
): TgRPCMessage {
  const { healthCheckPath, instanceCount, port } = basicDeployInput;
  const listenerCfg = new ListenerConfig()
    .setListenPort(port)
    .setHealthCheckPath(healthCheckPath);
  const basicDeployCfg = new BasicDeployConfig()
    .setInstanceCount(instanceCount)
    .setListenerCfg(listenerCfg);

  if (basicDeployInput?.envVars) {
    const envVars = new EnvVars().setEnvList(
      basicDeployInput.envVars.map(({ key, sensitive, value }) => {
        return new EnvVar()
          .setName(key)
          .setValue(value)
          .setSensitive(sensitive);
      })
    );

    basicDeployCfg.setEnvVars(envVars);
  }

  return basicDeployCfg;
}
