import { Link } from '@/components/common/Link'
import { Panel } from '@/components/surfaces/Panel'
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
  const hrefFor = (commit?: TInstallOverviewCommit) =>
    org?.id && install?.app_id && commit?.branch_id && commit?.workflow_id
      ? `/${org.id}/apps/${install.app_id}/branches/${commit.branch_id}/runs/${commit.workflow_id}`
      : undefined

  if (!error && !(tracking?.commits_behind && tracking.commits_behind > 0)) {
    return null
  }

  return (
    <div className={className}>
      <InstallBranchBanner
        error={error}
        onRetry={() => void refetch()}
        runsBehind={tracking?.commits_behind}
        branchName={branchName}
        branchHref={branchHref}
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
                runs={tracking?.pending_commits ?? []}
                runsBehind={tracking?.commits_behind}
                selectedCommit={tracking?.selected_commit}
                repo={tracking?.repo}
                hrefFor={hrefFor}
              />
            </Panel>
          )
        }
      />
    </div>
  )
}
