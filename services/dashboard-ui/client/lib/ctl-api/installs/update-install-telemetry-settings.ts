import { api } from '@/lib/api'
import type { TInstallTelemetrySettings } from '@/types'

export const updateInstallTelemetrySettings = ({
  installId,
  orgId,
  enabled,
}: {
  installId: string
  orgId: string
  enabled: boolean
}) =>
  api<TInstallTelemetrySettings>({
    path: `installs/${installId}/telemetry`,
    orgId,
    method: 'PATCH',
    body: { enabled },
  })
