import { api } from '@/lib/api'
import type { TWorkflowStep } from '@/types'

export const getWorkflowStepById = ({
  workflowId,
  workflowStepId,
  orgId,
}: {
  workflowId: string
  workflowStepId: string
  orgId: string
}) =>
  api<TWorkflowStep>({
    path: `workflows/${workflowId}/steps/${workflowStepId}`,
    orgId,
  })
