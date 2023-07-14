import { NoopConfig as NoopBuildConfig } from "@buf/nuon_components.grpc_node/build/v1/noop_pb";
import type { TgRPCMessage } from "../../../types";

export function initNoopBuildConfig(): TgRPCMessage {
  return new NoopBuildConfig();
}
