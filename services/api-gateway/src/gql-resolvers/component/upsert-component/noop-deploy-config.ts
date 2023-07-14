import { NoopConfig as NoopDeployConfig } from "@buf/nuon_components.grpc_node/deploy/v1/noop_pb";
import type { TgRPCMessage } from "../../../types";

export function initNoopDeployConfig(): TgRPCMessage {
  return new NoopDeployConfig();
}
