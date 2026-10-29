import { Link } from '@/components/common/Link'
import { Panel } from '@/components/surfaces/Panel'
import { ReconcileAppVersionButton } from '@/components/installs/management/ReconcileAppVersion'
import { useInstall } from '@/hooks/use-install'
import { useInstallOverview } from '@/hooks/use-install-overview'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TInstallOverviewCommit } from '@/types'
import { InstallBranchBanner } from './InstallBranchBanner'
import { SuccessfulBranchRuns } from './SuccessfulBranchRuns'

export const InstallBranchBannerContainer = ({
  className,
}: {
  className?: string
}) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { addPanel } = useSurfaces()
  const { data, error, refetch } = useInstallOverview()
  const tracking = data?.branch_tracking
  const branchName = tracking?.target_branch ?? install?.app_branch?.name
  const branchId = tracking?.branch_id ?? install?.app_branch?.id
  const branchHref =
    org?.id && install?.app_id && branchId
      ? `/${org.id}/apps/${install.app_id}/branches/${branchId}`
      : undefined
  const getRunHref = (run?: TInstallOverviewCommit) =>
    org?.id && install?.app_id && run?.branch_id && run?.workflow_id
      ? `/${org.id}/apps/${install.app_id}/branches/${run.branch_id}/runs/${run.workflow_id}`
      : undefined

  if (!error && !(tracking?.runs_behind && tracking.runs_behind > 0)) {
    return null
  }

  return (
    <div className={className}>
      <InstallBranchBanner
        error={error}
        onRetry={() => void refetch()}
        runsBehind={tracking?.runs_behind}
        branchName={branchName}
        branchHref={branchHref}
        reconcileAction={<ReconcileAppVersionButton size="sm" variant="secondary" />}
        onViewRuns={() =>
          addPanel(
            <Panel
              heading={`Successful app branch runs on ${branchName ?? 'the app branch'}`}
              size="half"
              data-testid="successful-runs-panel"
              footer={
                branchHref ? (
                  <Link className="mr-auto" href={`${branchHref}/runs`}>
                    View all runs
                  </Link>
                ) : undefined
              }
            >
              <SuccessfulBranchRuns
                branchName={branchName}
                runs={tracking?.successful_runs ?? []}
                runsBehind={tracking?.runs_behind}
                selectedCommit={tracking?.selected_commit}
                repo={tracking?.repo}
                getRunHref={getRunHref}
              />
            </Panel>
          )
        }
      />
    </div>
  )
}
