import { NoopConfig as NoopDeployConfig } from "../../../build/components/deploy/v1/noop_pb";
import type { TgRPCMessage } from "../../../types";

export function initNoopDeployConfig(): TgRPCMessage {
  return new NoopDeployConfig();
}
