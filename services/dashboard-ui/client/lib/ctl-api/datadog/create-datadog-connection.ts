import { api } from '@/lib/api'
import type {
  TCreateDatadogConnectionBody,
  TDatadogConnection,
} from '@/types'

export const createDatadogConnection = ({
  body,
  orgId,
}: {
  body: TCreateDatadogConnectionBody
  orgId: string
}) =>
  api<TDatadogConnection>({
    body,
    method: 'POST',
    orgId,
    path: `orgs/${orgId}/datadog/connections`,
  })
