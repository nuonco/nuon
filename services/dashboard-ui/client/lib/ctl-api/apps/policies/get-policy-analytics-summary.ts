import { api } from '@/lib/api'
import type { TPolicyAnalyticsSummary } from '@/types'

export const getPolicyAnalyticsSummary = ({
  appId,
  orgId,
  start,
  end,
  installId,
  policyId,
}: {
  appId: string
  orgId: string
  start?: string
  end?: string
  installId?: string
  policyId?: string
}) => {
  const params = new URLSearchParams()
  if (start) params.set('start', start)
  if (end) params.set('end', end)
  if (installId) params.set('install_id', installId)
  if (policyId) params.set('policy_id', policyId)
  const qs = params.toString()
  return api<TPolicyAnalyticsSummary>({
    path: `apps/${appId}/policy-analytics/summary${qs ? `?${qs}` : ''}`,
    orgId,
  })
}
