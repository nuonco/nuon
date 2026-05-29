import { api } from '@/lib/api'
import type { TDatadogConnection } from '@/types'

export const getDatadogConnection = ({
  orgId,
  connectionId,
}: {
  orgId: string
  connectionId: string
}) =>
  api<TDatadogConnection>({
    orgId,
    path: `orgs/${orgId}/datadog/connections/${connectionId}`,
  })
