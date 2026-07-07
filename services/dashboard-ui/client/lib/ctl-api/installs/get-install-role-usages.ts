import { api } from '@/lib/api'
import type { TInstallRoleUsage, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallRoleUsages = ({
  installId,
  orgId,
  roleName,
  limit,
  offset,
}: {
  installId: string
  orgId: string
  roleName: string
} & TPaginationParams) =>
  api<TInstallRoleUsage[]>({
    path: `installs/${installId}/roles/usages${buildQueryParams({
      role_name: roleName,
      limit,
      offset,
    })}`,
    orgId,
    paginated: true,
  })
