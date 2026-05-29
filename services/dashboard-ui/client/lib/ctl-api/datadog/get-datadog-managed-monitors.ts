import { api } from '@/lib/api'
import type { TDatadogManagedMonitor } from '@/types'

export const getDatadogManagedMonitors = ({
  orgId,
  connectionId,
  targetId,
}: {
  orgId: string
  connectionId?: string
  targetId?: string
}) => {
  const search = new URLSearchParams()
  if (connectionId) search.set('connection_id', connectionId)
  if (targetId) search.set('target_id', targetId)
  const query = search.toString()
  const path =
    `orgs/${orgId}/datadog/managed-monitors` + (query ? `?${query}` : '')
  return api<TDatadogManagedMonitor[]>({ orgId, path })
}
