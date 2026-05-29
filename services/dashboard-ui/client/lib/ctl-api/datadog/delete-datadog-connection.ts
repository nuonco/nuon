import { api } from '@/lib/api'

export const deleteDatadogConnection = ({
  orgId,
  connectionId,
}: {
  orgId: string
  connectionId: string
}) =>
  api<void>({
    method: 'DELETE',
    orgId,
    path: `orgs/${orgId}/datadog/connections/${connectionId}`,
  })
