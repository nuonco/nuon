import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type {
  TInstallWorkflow,
  TAppBranchRun,
  TComponentBuild,
  TInstallAppConfigVersion,
} from '@/types'

export interface IBranchEntry {
  branchId: string
  branchName: string
  active: boolean
  activatedAt?: string
  latestRun?: TInstallWorkflow
  branchRun?: TAppBranchRun
  builds: TComponentBuild[]
  installUpdates: TInstallAppConfigVersion[]
  appConfigId?: string
  configVersions?: TInstallAppConfigVersion[]
}

interface IBranchCard extends IBranchEntry {
  orgId: string
  appId: string
  installId: string
}

const CommitSection = ({ branchRun }: { branchRun?: TAppBranchRun }) => {
  const commit = branchRun?.vcs_connection_commit
  if (!commit) return null

  return (
    <div className="flex items-start gap-2.5">
      <Icon variant="GitCommitIcon" size={14} className="mt-0.5 shrink-0 text-cool-grey-400" />
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
  )
}

const BuildsList = ({ builds, orgId }: { builds: TComponentBuild[]; orgId: string }) => {
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
            className="flex items-center justify-between gap-3 px-3 py-1.5 rounded-md bg-cool-grey-50 dark:bg-dark-grey-700"
          >
            <div className="flex items-center gap-2 min-w-0">
              <Icon variant="PackageIcon" size={12} className="shrink-0 text-cool-grey-400" />
              <Text variant="subtext" className="truncate">
                {build.component_name || build.component_id}
              </Text>
            </div>
            <Status status={build.status_v2?.status || 'unknown'} />
          </div>
        ))}
      </div>
    </div>
  )
}

const InstallUpdatesList = ({
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
            className="flex items-center justify-between gap-3 px-3 py-1.5 rounded-md bg-cool-grey-50 dark:bg-dark-grey-700"
          >
            <div className="flex items-center gap-2 min-w-0">
              <Icon variant="CloudIcon" size={12} className="shrink-0 text-cool-grey-400" />
              {update.install_id ? (
                <Link href={`/${orgId}/installs/${update.install_id}`} className="truncate">
                  {update.install_id}
                </Link>
              ) : (
                <Text variant="subtext" theme="neutral">Unknown</Text>
              )}
            </div>
            <div className="flex items-center gap-2 shrink-0">
              <Status status={update.status?.status || 'unknown'} />
              {update.workflow_id && (
                <Link
                  href={`/${orgId}/installs/${update.install_id}/workflows/${update.workflow_id}`}
                  className="shrink-0"
                >
                  View workflow
                </Link>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

const ConfigVersionSummary = ({
  version,
  orgId,
  installId,
  appId,
}: {
  version: TInstallAppConfigVersion
  orgId: string
  installId: string
  appId: string
}) => {
  const status = version.status?.status || 'unknown'
  const branchRun = (version as any).app_branch_run
  const branchRunWorkflowId = branchRun?.workflow_id
  const branchId = branchRun?.app_branch_id

  return (
    <div className="flex items-center justify-between gap-3 px-3 py-1.5 rounded-md bg-cool-grey-50 dark:bg-dark-grey-700">
      <div className="flex items-center gap-2 min-w-0">
        <Icon variant="ArrowsClockwiseIcon" size={12} className="shrink-0 text-cool-grey-400" />
        <Text variant="subtext" className="truncate">
          {version.metadata?.triggered_by || 'config update'}
        </Text>
        <Time time={version.created_at} format="relative" variant="subtext" />
      </div>
      <div className="flex items-center gap-2 shrink-0">
        <Status status={status} />
        {branchId && branchRunWorkflowId && (
          <Link
            href={`/${orgId}/apps/${appId}/branches/${branchId}/runs/${branchRunWorkflowId}`}
          >
            View run
          </Link>
        )}
        {version.workflow_id && (
          <Link
            href={`/${orgId}/installs/${installId}/workflows/${version.workflow_id}`}
          >
            Workflow
          </Link>
        )}
      </div>
    </div>
  )
}

const BranchCard = ({
  branchId,
  branchName,
  active,
  activatedAt,
  latestRun,
  branchRun,
  builds,
  installUpdates,
  appConfigId,
  configVersions,
  orgId,
  appId,
  installId,
}: IBranchCard) => {
  const runStatus = latestRun?.status?.status || 'unknown'

  return (
    <Card className="!p-4 !gap-4">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2 min-w-0">
          <Icon variant="GitBranchIcon" size={16} className="shrink-0 text-cool-grey-400" />
          <Link
            href={`/${orgId}/apps/${appId}/branches/${branchId}`}
            className="truncate font-strong"
          >
            {branchName}
          </Link>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          <Badge size="sm" theme={active ? 'success' : 'neutral'}>
            {active ? 'Active' : 'Inactive'}
          </Badge>
          {activatedAt && (
            <Time time={activatedAt} format="relative" variant="subtext" />
          )}
        </div>
      </div>

      {latestRun ? (
        <div className="flex flex-col gap-3 border-t pt-3">
          <div className="flex items-center justify-between gap-3">
            <div className="flex items-center gap-2">
              <Text variant="subtext" weight="strong" theme="neutral">
                Latest run
              </Text>
              <Status status={runStatus} variant="badge" />
            </div>
            <div className="flex items-center gap-2">
              {latestRun.finished_at && (
                <Time time={latestRun.finished_at} format="relative" variant="subtext" />
              )}
              <Link
                href={`/${orgId}/apps/${appId}/branches/${branchId}/runs/${latestRun.id}`}
                className="shrink-0"
              >
                View run
              </Link>
            </div>
          </div>

          <CommitSection branchRun={branchRun} />
          <BuildsList builds={builds} orgId={orgId} />
          <InstallUpdatesList installUpdates={installUpdates} orgId={orgId} />
          {(configVersions?.length ?? 0) > 0 && (
            <div className="flex flex-col gap-1.5">
              <Text variant="subtext" weight="strong" theme="neutral">
                Config updates ({configVersions!.length})
              </Text>
              <div className="flex flex-col gap-1">
                {configVersions!.map((version) => (
                  <ConfigVersionSummary
                    key={version.id}
                    version={version}
                    orgId={orgId}
                    installId={installId}
                    appId={appId}
                  />
                ))}
              </div>
            </div>
          )}
        </div>
      ) : (
        <div className="border-t pt-3">
          <Text variant="subtext" theme="neutral">
            No runs yet
          </Text>
        </div>
      )}
    </Card>
  )
}

interface IInstallBranches {
  branches: IBranchEntry[]
  orgId: string
  appId: string
  installId: string
}

export const InstallBranches = ({ branches, orgId, appId, installId }: IInstallBranches) => {
  if (branches.length === 0) {
    return (
      <EmptyState
        variant="diagram"
        emptyTitle="No app branches connected"
        emptyMessage="Add labels to this install to connect it to an app branch install group."
      />
    )
  }

  return (
    <div className="flex flex-col gap-3">
      <Text variant="base" weight="strong">
        Connected branches
      </Text>
      <div className="flex flex-col gap-3">
        {branches.map((entry) => (
          <BranchCard
            key={entry.branchId}
            {...entry}
            orgId={orgId}
            appId={appId}
            installId={installId}
          />
        ))}
      </div>
    </div>
  )
}
