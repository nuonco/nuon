import { api } from '@/lib/api'
import type { TSlackChannelsResponse } from '@/types'

export const getSlackChannels = ({
  orgId,
  installationId,
  cursor,
  limit,
  types,
}: {
  orgId: string
  installationId: string
  cursor?: string
  limit?: number
  types?: string
}) => {
  const search = new URLSearchParams()
  if (cursor) search.set('cursor', cursor)
  if (limit) search.set('limit', String(limit))
  if (types) search.set('types', types)

  const query = search.toString()
  const path =
    `orgs/${orgId}/slack/installations/${installationId}/channels` +
    (query ? `?${query}` : '')

  return api<TSlackChannelsResponse>({ orgId, path })
}
