import { api } from '@/lib/api'

export const triggerAppInstallSync = ({
  appId,
  orgId,
}: {
  appId: string
  orgId: string
}) =>
  api<{ status: string }>({
    path: `apps/${appId}/install-syncs`,
    method: 'POST',
    orgId,
  })
