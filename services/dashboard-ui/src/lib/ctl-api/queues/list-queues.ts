import { api } from '@/lib/api'
import type { TQueue } from './get-queue'

type TListQueuesParams = {
  orgId: string
  ownerId?: string
  ownerType?: string
  limit?: number
  offset?: number
}

export const listQueues = ({
  orgId,
  ownerId,
  ownerType,
  limit = 50,
  offset = 0,
}: TListQueuesParams) => {
  const params = new URLSearchParams()
  if (ownerId) params.append('owner_id', ownerId)
  if (ownerType) params.append('owner_type', ownerType)
  params.append('limit', limit.toString())
  params.append('offset', offset.toString())

  return api<TQueue[]>({
    path: `queues?${params.toString()}`,
    orgId,
  })
}
