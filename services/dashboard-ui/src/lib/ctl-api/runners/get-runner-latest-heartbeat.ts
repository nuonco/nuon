import { api } from "@/lib/api";
import type { TRunnerMngHeartbeat } from "@/types";

export const getRunnerLatestHeartbeat = ({
  runnerId,
  orgId,
}: {
  runnerId: string;
  orgId: string;
}) =>
  api<TRunnerMngHeartbeat>({
    path: `runners/${runnerId}/heart-beats/latest`,
    orgId,
  });
