import { api } from "@/lib/api";
import type { TWorkflowStep, TPaginationParams } from "@/types";
import { buildQueryParams } from "@/utils/build-query-params";

export const getWorkflowSteps = ({
  workflowId,
  limit,
  offset,
  orgId,
}: { workflowId: string; orgId: string } & TPaginationParams) =>
  api<TWorkflowStep[]>({
    path: `workflows/${workflowId}/steps${buildQueryParams({ limit, offset })}`,
    orgId,
  });
