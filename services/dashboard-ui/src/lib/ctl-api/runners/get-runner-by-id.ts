import { api } from "@/lib/api";
import type { TRunner } from "@/types";

export const getRunnerById = ({
  runnerId,
  orgId,
}: {
  runnerId: string;
  orgId: string;
}) =>
  api<TRunner>({
    path: `runners/${runnerId}`,
    orgId,
  });
