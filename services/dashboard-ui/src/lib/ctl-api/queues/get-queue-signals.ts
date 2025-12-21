import { api } from '@/lib/api'

export type TCompositeStatus = {
  status: string
  message?: string
  error?: string
  user_error?: string
}

export type TQueueSignal = {
  id: string
  created_by_id: string
  created_at: string
  updated_at: string
  org_id: string
  queue_id: string
  emitter_id?: string
  owner_id: string
  owner_type: string
  status: TCompositeStatus
  type: string
  signal: Record<string, any>
  workflow: {
    id: string
    id_template?: string
  }
}

export const getQueueSignals = ({
  queueId,
  ownerId,
  ownerType,
  status,
  type,
  limit,
  offset,
  orgId,
}: {
  queueId: string
  ownerId?: string
  ownerType?: string
  status?: string
  type?: string
  limit?: number
  offset?: number
  orgId: string
}) =>
  api<TQueueSignal[]>({
    path: `queues/${queueId}/signals`,
    orgId,
    params: {
      owner_id: ownerId,
      owner_type: ownerType,
      status,
      type,
      limit,
      offset,
    },
  })
