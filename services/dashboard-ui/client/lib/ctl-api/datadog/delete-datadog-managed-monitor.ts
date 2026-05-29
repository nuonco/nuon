import { api } from '@/lib/api'

export const deleteDatadogManagedMonitor = ({
  orgId,
  monitorId,
}: {
  orgId: string
  monitorId: string
}) =>
  api<void>({
    method: 'DELETE',
    orgId,
    path: `orgs/${orgId}/datadog/managed-monitors/${monitorId}`,
  })
