import { api } from '@/lib/api'

export type TComponentDiffEntry = {
  component_id: string
  component_name?: string
  component_type?: string
  old_checksum?: string
  new_checksum?: string
}

export type TInstallConfigDiff = {
  added: TComponentDiffEntry[]
  removed: TComponentDiffEntry[]
  changed: TComponentDiffEntry[]
  unchanged: TComponentDiffEntry[]
  sandbox_changed?: boolean
  sandbox_old_id?: string
  sandbox_new_id?: string
  stack_changed?: boolean
  stack_old_id?: string
  stack_new_id?: string
}

export const getInstallAppConfigVersionDiff = ({
  installId,
  versionId,
  orgId,
}: {
  installId: string
  versionId: string
  orgId: string
}) =>
  api<TInstallConfigDiff>({
    path: `installs/${installId}/app-config-versions/${versionId}/diff`,
    orgId,
  })
