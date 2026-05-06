import { api } from '@/lib/api'

type TCancelWorkflowsResponse = {
  cancelled: string[]
  errors?: { workflow_id: string; error: string }[]
}

export async function cancelWorkflows({
  orgId,
  workflowIds,
}: {
  orgId: string
  workflowIds: string[]
}) {
  return api<TCancelWorkflowsResponse>({
    method: 'POST',
    orgId,
    path: 'workflows/cancel',
    body: { workflow_ids: workflowIds },
  })
}
