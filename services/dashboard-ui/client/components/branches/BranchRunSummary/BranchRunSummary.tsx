import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TComponentBuild, TInstallAppConfigVersion, TAppBranchRun } from '@/types'

interface IBranchRunSummary {
  branchRun?: TAppBranchRun
  builds: TComponentBuild[]
  installUpdates: TInstallAppConfigVersion[]
  orgId: string
  appId: string
  branchId: string
  runStatus: string
}

const CommitSection = ({ branchRun }: { branchRun?: TAppBranchRun }) => {
  const commit = branchRun?.vcs_connection_commit
  if (!commit) return null

  return (
    <div className="flex flex-col gap-1.5">
      <Text variant="subtext" weight="strong" theme="neutral">
        Commit
      </Text>
      <div className="flex items-start gap-3">
        <Icon variant="GitCommitIcon" size={16} className="mt-0.5 shrink-0 text-cool-grey-400" />
        <div className="flex flex-col gap-0.5 min-w-0">
          <Text variant="body" weight="strong" className="truncate">
            {commit.message?.split('\n')[0]?.trim()}
          </Text>
          <div className="flex items-center gap-2 flex-wrap">
            {commit.sha && (
              <Badge size="sm" variant="code">
                {commit.sha.slice(0, 8)}
              </Badge>
            )}
            {commit.author_name && (
              <Text variant="subtext" theme="neutral">
                {commit.author_name}
              </Text>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

const BuildsSection = ({
  builds,
  orgId,
}: {
  builds: TComponentBuild[]
  orgId: string
}) => {
  if (builds.length === 0) return null

  return (
    <div className="flex flex-col gap-1.5">
      <Text variant="subtext" weight="strong" theme="neutral">
        Builds ({builds.length})
      </Text>
      <div className="flex flex-col gap-1">
        {builds.map((build) => (
          <div
            key={build.id}
            className="flex items-center justify-between gap-3 px-3 py-2 rounded-md bg-cool-grey-50 dark:bg-dark-grey-700"
          >
            <div className="flex items-center gap-2 min-w-0">
              <Icon variant="PackageIcon" size={14} className="shrink-0 text-cool-grey-400" />
              <Text variant="body" className="truncate">
                {build.component_name || build.component_id}
              </Text>
            </div>
            <div className="flex items-center gap-2 shrink-0">
              <Status status={build.status_v2?.status || 'unknown'} />
              {build.id && (
                <Badge size="sm" variant="code">
                  {build.id.slice(0, 10)}
                </Badge>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

const InstallUpdatesSection = ({
  installUpdates,
  orgId,
}: {
  installUpdates: TInstallAppConfigVersion[]
  orgId: string
}) => {
  if (installUpdates.length === 0) return null

  return (
    <div className="flex flex-col gap-1.5">
      <Text variant="subtext" weight="strong" theme="neutral">
        Installs updated ({installUpdates.length})
      </Text>
      <div className="flex flex-col gap-1">
        {installUpdates.map((update) => (
          <div
            key={update.id}
            className="flex items-center justify-between gap-3 px-3 py-2 rounded-md bg-cool-grey-50 dark:bg-dark-grey-700"
          >
            <div className="flex items-center gap-2 min-w-0">
              <Icon variant="CloudIcon" size={14} className="shrink-0 text-cool-grey-400" />
              {update.install_id ? (
                <Link href={`/${orgId}/installs/${update.install_id}`} className="truncate">
                  {update.install_id}
                </Link>
              ) : (
                <Text variant="body" theme="neutral">Unknown</Text>
              )}
            </div>
            <div className="flex items-center gap-2 shrink-0">
              <Status status={update.status?.status || 'unknown'} />
              {update.workflow_id && (
                <Link
                  href={`/${orgId}/installs/${update.install_id}/workflows/${update.workflow_id}`}
                  className="text-xs"
                >
                  Workflow
                </Link>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export const BranchRunSummary = ({
  branchRun,
  builds,
  installUpdates,
  orgId,
  appId,
  branchId,
  runStatus,
}: IBranchRunSummary) => {
  const statusTheme = runStatus === 'success' ? 'success' : runStatus === 'failed' ? 'error' : 'neutral'

  return (
    <Card className="!p-5 !gap-5 border-l-4" style={{
      borderLeftColor: statusTheme === 'success'
        ? 'var(--color-green-500, #22c55e)'
        : statusTheme === 'error'
          ? 'var(--color-red-500, #ef4444)'
          : 'var(--color-cool-grey-300, #d1d5db)',
    }}>
      <div className="flex items-center justify-between gap-3">
        <Text variant="base" weight="strong">
          Run summary
        </Text>
        <Status status={runStatus} variant="badge" />
      </div>

      <div className="flex flex-col gap-4">
        <CommitSection branchRun={branchRun} />
        <BuildsSection builds={builds} orgId={orgId} />
        <InstallUpdatesSection installUpdates={installUpdates} orgId={orgId} />

        {builds.length === 0 && installUpdates.length === 0 && !branchRun?.vcs_connection_commit && (
          <Text variant="subtext" theme="neutral">
            No details available for this run.
          </Text>
        )}
      </div>
    </Card>
  )
}
