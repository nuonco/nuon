import { api } from '@/lib/api'

export async function refreshInstallHealthClusterAccess({
  installId,
  orgId,
  roleName,
}: {
  installId: string
  orgId: string
  roleName?: string
}) {
  return api<{
    cluster_found: boolean
    cluster_id?: string
    role_name?: string
  }>({
    body: { role_name: roleName ?? '' },
    method: 'POST',
    orgId,
    path: `installs/${installId}/health/cluster-access`,
  })
}
