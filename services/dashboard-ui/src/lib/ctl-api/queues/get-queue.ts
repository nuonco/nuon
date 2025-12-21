import { api } from '@/lib/api'

export type TQueue = {
  id: string
  name: string
  owner_id: string
  owner_type: string
  org_id: string
  created_by_id: string
  created_at: string
  updated_at: string
}

export const getQueue = ({
  queueId,
  orgId,
}: {
  queueId: string
  orgId: string
}) =>
  api<TQueue>({
    path: `queues/${queueId}`,
    orgId,
  })
