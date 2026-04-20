import { api } from '@/lib/api'
import type { TInstallRoleUsage } from '@/types'

export const getInstallRoleUsages = ({
  installId,
  orgId,
  roleName,
}: {
  installId: string
  orgId: string
  roleName: string
}) =>
  api<TInstallRoleUsage[]>({
    path: `installs/${installId}/roles/usages?role_name=${encodeURIComponent(roleName)}`,
    orgId,
  })
