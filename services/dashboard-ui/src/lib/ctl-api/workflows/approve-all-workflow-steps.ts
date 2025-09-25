import { api } from '@/lib/api'
import type { TWorkflow } from "@/types"

export type TApproveAllWorkflowStepsBody = {
  approval_option: "approve-all" | "prompt"
}

export async function approveAllWorkflowSteps({
  body,
  orgId,
  workflowId,
}: {
  body: TApproveAllWorkflowStepsBody
  orgId: string
  workflowId: string
}) {
  return api<TWorkflow>({
    body,
    method: 'PATCH',
    orgId,
    path: `workflows/${workflowId}`,
  })
}
