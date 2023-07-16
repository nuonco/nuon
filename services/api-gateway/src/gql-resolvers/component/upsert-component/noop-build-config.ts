import { NoopConfig as NoopBuildConfig } from "../../../build/components/build/v1/noop_pb";
import type { TgRPCMessage } from "../../../types";

export function initNoopBuildConfig(): TgRPCMessage {
  return new NoopBuildConfig();
}
