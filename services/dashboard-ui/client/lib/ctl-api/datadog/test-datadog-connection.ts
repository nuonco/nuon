import { api } from '@/lib/api'
import type { TDatadogTestConnectionResponse } from '@/types'

export const testDatadogConnection = ({
  orgId,
  connectionId,
}: {
  orgId: string
  connectionId: string
}) =>
  api<TDatadogTestConnectionResponse>({
    method: 'POST',
    orgId,
    path: `orgs/${orgId}/datadog/connections/${connectionId}/test`,
  })
