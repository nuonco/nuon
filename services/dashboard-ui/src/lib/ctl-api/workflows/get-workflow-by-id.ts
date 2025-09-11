import { api } from "@/lib/api";
import type { TWorkflow } from "@/types";

export const getWorkflowById = ({
  workflowId,
  orgId,
}: {
  workflowId: string;
  orgId: string;
}) =>
  api<TWorkflow>({
    path: `workflows/${workflowId}`,
    orgId,
  });
