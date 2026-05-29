import { api } from '@/lib/api'
import type {
  TDatadogConnection,
  TUpdateDatadogConnectionBody,
} from '@/types'

export const updateDatadogConnection = ({
  orgId,
  connectionId,
  body,
}: {
  orgId: string
  connectionId: string
  body: TUpdateDatadogConnectionBody
}) =>
  api<TDatadogConnection>({
    body,
    method: 'PATCH',
    orgId,
    path: `orgs/${orgId}/datadog/connections/${connectionId}`,
  })
