import { api } from '@/lib/api'

export type TWorkflowQueueItem = {
  workflow_id: string
  workflow_type: string
  status: string
  created_at: string
  metadata?: Record<string, string>
}

export type TWorkflowQueuePosition = {
  position: number
  queue_depth: number
  signals_ahead: TWorkflowQueueItem[]
}

export const getWorkflowQueuePosition = ({
  workflowId,
  orgId,
}: {
  workflowId: string
  orgId: string
}) =>
  api<TWorkflowQueuePosition>({
    path: `workflows/${workflowId}/queue-position`,
    orgId,
  })
