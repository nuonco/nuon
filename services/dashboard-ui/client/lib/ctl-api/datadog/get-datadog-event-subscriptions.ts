import { api } from '@/lib/api'
import type { TDatadogEventSubscription } from '@/types'

export const getDatadogEventSubscriptions = ({
  orgId,
  connectionId,
}: {
  orgId: string
  connectionId?: string
}) => {
  const search = new URLSearchParams()
  if (connectionId) search.set('connection_id', connectionId)
  const query = search.toString()
  const path =
    `orgs/${orgId}/datadog/event-subscriptions` + (query ? `?${query}` : '')
  return api<TDatadogEventSubscription[]>({ orgId, path })
}
