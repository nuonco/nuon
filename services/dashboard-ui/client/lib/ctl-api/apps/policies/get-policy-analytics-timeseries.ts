import { api } from '@/lib/api'
import type { TPolicyAnalyticsTimeseries } from '@/types'

export const getPolicyAnalyticsTimeseries = ({
  appId,
  orgId,
  start,
  end,
  groupBy,
  installId,
  policyId,
}: {
  appId: string
  orgId: string
  start?: string
  end?: string
  groupBy?: string
  installId?: string
  policyId?: string
}) => {
  const params = new URLSearchParams()
  if (start) params.set('start', start)
  if (end) params.set('end', end)
  if (groupBy) params.set('group_by', groupBy)
  if (installId) params.set('install_id', installId)
  if (policyId) params.set('policy_id', policyId)
  const qs = params.toString()
  return api<TPolicyAnalyticsTimeseries>({
    path: `apps/${appId}/policy-analytics/timeseries${qs ? `?${qs}` : ''}`,
    orgId,
  })
}
