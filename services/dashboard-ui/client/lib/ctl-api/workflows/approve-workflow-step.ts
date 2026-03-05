import { api } from '@/lib/api'
import type { TWorkflowStepApprovalResponse } from '@/types'

export type TApproveWorkflowStepBody = {
  note: string
  response_type:
    | 'approve'
    | 'deny'
    | 'retry'
    | 'deny-skip-current'
    | 'deny-skip-current-and-dependents'
}

export async function approveWorkflowStep({
  approvalId,
  body,
  orgId,
  workflowId,
  workflowStepId,
}: {
  approvalId: string
  body: TApproveWorkflowStepBody
  orgId: string
  workflowId: string
  workflowStepId: string
}) {
  return api<TWorkflowStepApprovalResponse>({
    body,
    method: 'POST',
    orgId,
    path: `workflows/${workflowId}/steps/${workflowStepId}/approvals/${approvalId}/response`,
  })
}
