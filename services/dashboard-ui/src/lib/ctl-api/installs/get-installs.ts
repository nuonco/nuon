import { api } from "@/lib/api";
import type { TInstall, TPaginationParams } from "@/types";
import { buildQueryParams } from "@/utils/build-query-params";

export const getInstalls = ({
  limit,
  offset,
  orgId,
  q,
}: { orgId: string; q?: string } & TPaginationParams) =>
  api<TInstall[]>({
    path: `installs${buildQueryParams({ limit, offset, q })}`,
    orgId,
  });
