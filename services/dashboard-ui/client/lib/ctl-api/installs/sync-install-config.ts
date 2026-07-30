import { api } from '@/lib/api'

export const syncInstallConfig = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<{ status: string }>({
    path: `installs/${installId}/sync-config`,
    method: 'POST',
    orgId,
    body: {},
  })
