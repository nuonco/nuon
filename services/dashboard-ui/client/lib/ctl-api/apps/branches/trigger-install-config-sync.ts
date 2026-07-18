import { api } from '@/lib/api'

export const triggerInstallConfigSync = ({
  appId,
  branchId,
  orgId,
  installName,
}: {
  appId: string
  branchId: string
  orgId: string
  installName?: string
}) =>
  api<{ status: string }>({
    path: `apps/${appId}/branches/${branchId}/sync-install-configs`,
    method: 'POST',
    orgId,
    body: { install_name: installName },
  })
