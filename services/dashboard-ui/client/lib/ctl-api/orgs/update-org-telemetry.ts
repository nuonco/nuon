import { api } from '@/lib/api'
import type { TOrg } from '@/types'

export const updateOrgTelemetry = ({
  orgId,
  enabled,
}: {
  orgId: string
  enabled: boolean
}) =>
  api<TOrg>({
    path: 'orgs/current/telemetry',
    orgId,
    method: 'PATCH',
    body: { enabled },
  })
