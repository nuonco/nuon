import { api } from '@/lib/api'

export async function cancelWorkflow({
  orgId,
  workflowId,
}: {
  orgId: string
  workflowId: string
}) {
  return api<boolean>({
    method: 'POST',
    orgId,
    path: `workflows/${workflowId}/cancel`,
  })
}
