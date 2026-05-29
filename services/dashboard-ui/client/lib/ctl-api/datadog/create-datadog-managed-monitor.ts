import { api } from '@/lib/api'
import type {
  TCreateDatadogManagedMonitorBody,
  TDatadogManagedMonitor,
} from '@/types'

export const createDatadogManagedMonitor = ({
  orgId,
  body,
}: {
  orgId: string
  body: TCreateDatadogManagedMonitorBody
}) =>
  api<TDatadogManagedMonitor>({
    body,
    method: 'POST',
    orgId,
    path: `orgs/${orgId}/datadog/managed-monitors`,
  })
