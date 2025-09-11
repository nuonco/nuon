import { api } from "@/lib/api";
import type { TWorkflow, TPaginationParams } from "@/types";
import { buildQueryParams } from "@/utils/build-query-params";

export const getInstallWorkflows = ({
  installId,
  limit,
  offset,
  orgId,
}: { installId: string; orgId: string } & TPaginationParams) =>
  api<TWorkflow[]>({
    path: `installs/${installId}/workflows${buildQueryParams({ limit, offset })}`,
    orgId,
  });
