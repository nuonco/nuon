import { api } from '@/lib/api'

export async function retryNowWorkflowStep({
  orgId,
  workflowId,
  stepId,
}: {
  orgId: string
  workflowId: string
  stepId: string
}) {
  return api<{ workflow_id: string }>({
    method: 'POST',
    orgId,
    path: `workflows/${workflowId}/steps/${stepId}/retry-now`,
  })
}
