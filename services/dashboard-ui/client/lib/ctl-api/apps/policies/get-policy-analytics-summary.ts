import { api } from '@/lib/api'
import type { TPolicyAnalyticsSummary } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

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
}) =>
  api<TPolicyAnalyticsSummary>({
    path: `apps/${appId}/policy-analytics/summary${buildQueryParams({ start, end, install_id: installId, policy_id: policyId })}`,
    orgId,
  })
