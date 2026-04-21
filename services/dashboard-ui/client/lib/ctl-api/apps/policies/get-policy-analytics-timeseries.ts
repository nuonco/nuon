import { api } from '@/lib/api'
import type { TPolicyAnalyticsTimeseries } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

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
}) =>
  api<TPolicyAnalyticsTimeseries>({
    path: `apps/${appId}/policy-analytics/timeseries${buildQueryParams({ start, end, group_by: groupBy, install_id: installId, policy_id: policyId })}`,
    orgId,
  })
