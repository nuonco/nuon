import { api } from "@/lib/api";
import type { TDeploy, TPaginationParams } from "@/types";
import { buildQueryParams } from "@/utils/build-query-params";

export const getInstallComponentDeploys = ({
  installId,
  componentId,
  orgId,
  limit,
  offset,
  q,
}: {
  installId: string;
  componentId: string;
  orgId: string;
  q?: string;
} & TPaginationParams) =>
  api<TDeploy[]>({
    path: `installs/${installId}/components/${componentId}/deploys${buildQueryParams({ limit, offset, q })}`,
    orgId,
  });
