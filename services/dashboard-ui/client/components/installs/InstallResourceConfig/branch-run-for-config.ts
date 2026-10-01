import type { TAppBranchRun, TInstallAppConfigVersion } from '@/types'

export const branchRunForConfig = (
  versions: TInstallAppConfigVersion[] | undefined,
  appConfigId?: string
): TAppBranchRun | undefined => {
  if (!appConfigId || !versions?.length) return undefined

  return [...versions]
    .filter(
      (version) =>
        version.new_app_config_id === appConfigId && version.app_branch_run
    )
    .sort((a, b) => (b.created_at ?? '').localeCompare(a.created_at ?? ''))[0]
    ?.app_branch_run
}

export const branchRunHref = ({
  appId,
  orgId,
  run,
}: {
  appId?: string
  orgId?: string
  run?: TAppBranchRun
}) => {
  const branchId = run?.app_branch?.id
  if (!orgId || !appId || !branchId || !run?.id) return undefined
  return `/${orgId}/apps/${appId}/branches/${branchId}/runs/${run.id}`
}
