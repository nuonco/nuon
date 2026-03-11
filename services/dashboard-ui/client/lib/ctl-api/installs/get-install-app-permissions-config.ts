import { api } from '@/lib/api'

export type TInstallPermissionsRolePolicy = {
  id?: string
  name?: string
  managed_policy_name?: string
  contents?: string
}

export type TInstallPermissionsRoleStatus = {
  id: string
  name: string
  display_name: string
  description: string
  type: string
  policies: TInstallPermissionsRolePolicy[]
  permissions_boundary: string
  created_at: string
  enabled: boolean
  arn: string
}

export type TInstallAppPermissionsConfig = {
  provision_role: TInstallPermissionsRoleStatus | null
  deprovision_role: TInstallPermissionsRoleStatus | null
  maintenance_role: TInstallPermissionsRoleStatus | null
  break_glass_roles: TInstallPermissionsRoleStatus[]
  custom_roles: TInstallPermissionsRoleStatus[]
}

export const getInstallAppPermissionsConfig = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<TInstallAppPermissionsConfig>({
    path: `installs/${installId}/app-permissions-config`,
    orgId,
  })
