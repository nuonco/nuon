import { api } from '@/lib/api'

export type TRetryWorkflowStepBody = {
  operation: 'retry-step' | 'skip-step'
  step_id: string
}

export async function retryWorkflowStep({
  body,
  orgId,
  workflowId,
}: {
  body: TRetryWorkflowStepBody
  orgId: string
  workflowId: string
}) {
  return api<{ workflow_id: string }>({
    body,
    method: 'POST',
    orgId,
    path: `workflows/${workflowId}/retry`,
  })
}
