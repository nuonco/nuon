import { api } from '@/lib/api'

export async function retryWorkflowStep({
  orgId,
  workflowId,
  stepId,
}: {
  orgId: string
  workflowId: string
  stepId: string
}) {
  return api<{ workflow_id: string; retryable: boolean }>({
    method: 'POST',
    orgId,
    path: `workflows/${workflowId}/steps/${stepId}/retry`,
  })
}
