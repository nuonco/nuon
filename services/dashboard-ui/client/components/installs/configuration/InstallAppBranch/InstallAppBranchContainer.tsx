import { useQuery } from '@tanstack/react-query'
import { InstallVersionsTimeline } from '@/components/install-versions/InstallVersionsTimeline'
import { useCurrentAppBranchRun } from '@/hooks/use-current-app-branch-run'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getAppBranch } from '@/lib'
import { InstallAppBranch } from './InstallAppBranch'

const branchRunHref = ({
  appId,
  branchId,
  orgId,
  runId,
}: {
  appId?: string
  branchId?: string
  orgId?: string
  runId?: string
}) =>
  appId && branchId && orgId && runId
    ? `/${orgId}/apps/${appId}/branches/${branchId}/runs/${runId}`
    : undefined

export const InstallAppBranchContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { run: appliedRun, isLoading: appliedRunLoading } =
    useCurrentAppBranchRun()
  const branchId = install?.app_branch?.id

  const { data: branch, isLoading: branchLoading } = useQuery({
    queryKey: [
      'install-configuration-app-branch',
      org?.id,
      install?.app_id,
      branchId,
    ],
    queryFn: () =>
      getAppBranch({
        orgId: org!.id,
        appId: install!.app_id!,
        branchId: branchId!,
        latestConfig: true,
      }),
    enabled: !!org?.id && !!install?.app_id && !!branchId,
  })

  const latestRun = branch?.latest_run
  const latestConfig = latestRun?.app_branch_config ?? branch?.configs?.at(0)
  const branchHref =
    org?.id && install?.app_id && branchId
      ? `/${org.id}/apps/${install.app_id}/branches/${branchId}`
      : undefined

  return (
    <InstallAppBranch
      appliedConfigId={install?.app_config_id}
      branchName={install?.app_branch?.name}
      branchHref={branchHref}
      branchConfig={latestConfig}
      latestRun={latestRun}
      latestRunHref={branchRunHref({
        orgId: org?.id,
        appId: install?.app_id,
        branchId,
        runId: latestRun?.id,
      })}
      appliedRun={appliedRun}
      appliedRunHref={branchRunHref({
        orgId: org?.id,
        appId: install?.app_id,
        branchId: appliedRun?.app_branch?.id ?? branchId,
        runId: appliedRun?.id,
      })}
      isLoading={branchLoading || appliedRunLoading}
      history={<InstallVersionsTimeline />}
    />
  )
}
