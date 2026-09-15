import { api } from '@/lib/api'
import type { TInstallTelemetrySettings } from '@/types'

export const getInstallTelemetrySettings = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<TInstallTelemetrySettings>({
    path: `installs/${installId}/telemetry`,
    orgId,
  })
