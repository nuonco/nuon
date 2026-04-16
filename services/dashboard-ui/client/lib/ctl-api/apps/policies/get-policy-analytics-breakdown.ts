import { api } from '@/lib/api'
import type { TPolicyAnalyticsBreakdown } from '@/types'

export const getPolicyAnalyticsBreakdown = ({
  appId,
  orgId,
  dimension,
  start,
  end,
  limit,
  installId,
  policyId,
}: {
  appId: string
  orgId: string
  dimension: string
  start?: string
  end?: string
  limit?: number
  installId?: string
  policyId?: string
}) => {
  const params = new URLSearchParams()
  params.set('dimension', dimension)
  if (start) params.set('start', start)
  if (end) params.set('end', end)
  if (limit) params.set('limit', String(limit))
  if (installId) params.set('install_id', installId)
  if (policyId) params.set('policy_id', policyId)
  return api<TPolicyAnalyticsBreakdown>({
    path: `apps/${appId}/policy-analytics/breakdown?${params.toString()}`,
    orgId,
  })
}
