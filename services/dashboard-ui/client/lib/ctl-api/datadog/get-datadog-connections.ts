import { api } from '@/lib/api'
import type { TDatadogConnection } from '@/types'

export const getDatadogConnections = ({ orgId }: { orgId: string }) =>
  api<TDatadogConnection[]>({
    orgId,
    path: `orgs/${orgId}/datadog/connections`,
  })
