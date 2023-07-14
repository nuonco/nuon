import { BasicConfig as BasicDeployConfig } from "@buf/nuon_components.grpc_node/deploy/v1/basic_pb";
import { ListenerConfig } from "@buf/nuon_components.grpc_node/deploy/v1/config_pb";
import type { BasicDeployConfigInput, TgRPCMessage } from "../../../types";

export function initBasicDeployConfig(
  basicDeployInput: BasicDeployConfigInput
): TgRPCMessage {
  const { healthCheckPath, instanceCount, port } = basicDeployInput;
  const listenerCfg = new ListenerConfig()
    .setListenPort(port)
    .setHealthCheckPath(healthCheckPath);
  return new BasicDeployConfig()
    .setInstanceCount(instanceCount)
    .setListenerCfg(listenerCfg);
}
