import { api } from '@/lib/api'
import type { TAPIError, TManagedServiceAccount } from '@/types'

export const getInstallTelemetryCollector = async ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) => {
  try {
    return await api<TManagedServiceAccount>({
      path: `installs/${installId}/telemetry/collector`,
      orgId,
    })
  } catch (error) {
    if ((error as TAPIError).status === 404) return null
    throw error
  }
}

export const ensureInstallTelemetryCollector = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<TManagedServiceAccount>({
    path: `installs/${installId}/telemetry/collector`,
    method: 'POST',
    orgId,
  })

export const deleteInstallTelemetryCollector = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<void>({
    path: `installs/${installId}/telemetry/collector`,
    method: 'DELETE',
    orgId,
  })
