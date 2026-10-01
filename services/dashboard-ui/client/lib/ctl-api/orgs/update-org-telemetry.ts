import { api } from '@/lib/api'
import type { TOrg } from '@/types'

export type TOrgTelemetryUpdate = {
  enabled?: boolean
  relay_endpoint?: string | null
}

export const updateOrgTelemetry = ({
  orgId,
  ...settings
}: {
  orgId: string
} & TOrgTelemetryUpdate) =>
  api<TOrg>({
    path: 'orgs/current/telemetry',
    orgId,
    method: 'PATCH',
    body: settings,
  })
