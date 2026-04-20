import { api } from '@/lib/api'
import type { TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'
import type { TInstallPermissionsRoleStatus } from './get-install-app-permissions-config'

export type TInstallRole = {
  id: string
  install_id: string
  app_role_config_id: string
  app_role_config: TInstallPermissionsRoleStatus
  enabled: boolean
  provisioned: boolean
  role_id: string
  created_at: string
}

export const getLatestInstallRoles = ({
  installId,
  orgId,
  limit,
  offset,
  q,
}: {
  installId: string
  orgId: string
  q?: string
} & TPaginationParams) =>
  api<TInstallRole[]>({
    path: `installs/${installId}/roles/latest${buildQueryParams({ limit, offset, q })}`,
    orgId,
    paginated: true,
  })
