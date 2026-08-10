import { api } from '@/lib/api'

export const deleteAppInstallsConfig = ({
  appId,
  configId,
  orgId,
}: {
  appId: string
  configId: string
  orgId: string
}) =>
  api<{ status: string }>({
    path: `apps/${appId}/installs-configs/${configId}`,
    method: 'DELETE',
    orgId,
  })
