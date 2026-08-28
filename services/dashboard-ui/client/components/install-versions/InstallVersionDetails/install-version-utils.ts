import type { TInstallAppConfigVersion } from '@/types'

export const resolveInstallVersionStatus = (
  version: TInstallAppConfigVersion
): string => {
  const versionStatus = version?.status?.status
  const workflowStatus = version?.workflow?.status?.status
  if (
    (versionStatus === 'pending' || versionStatus === 'in-progress') &&
    (workflowStatus === 'error' || workflowStatus === 'cancelled')
  ) {
    return workflowStatus
  }
  return versionStatus || 'unknown'
}

export const installVersionSource = (
  version: TInstallAppConfigVersion
): string =>
  version?.metadata?.triggered_by ||
  (version?.app_branch_run_id ? 'app-branch' : 'sync')
