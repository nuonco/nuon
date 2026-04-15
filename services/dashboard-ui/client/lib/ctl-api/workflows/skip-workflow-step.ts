import { api } from '@/lib/api'

export async function skipWorkflowStep({
  orgId,
  workflowId,
  stepId,
}: {
  orgId: string
  workflowId: string
  stepId: string
}) {
  return api<{ workflow_id: string; skippable: boolean }>({
    method: 'POST',
    orgId,
    path: `workflows/${workflowId}/steps/${stepId}/skip`,
  })
}
