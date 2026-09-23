import type { ReactNode } from 'react'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TAppBranchConfig, TAppBranchRun } from '@/types'
import { TrackedRunCard } from './TrackedRunCard'

export interface IInstallAppBranch {
  appliedConfigId?: string
  appliedRun?: TAppBranchRun
  appliedRunHref?: string
  branchConfig?: TAppBranchConfig
  branchHref?: string
  branchName?: string
  history?: ReactNode
  isLoading?: boolean
  latestRun?: TAppBranchRun
  latestRunHref?: string
}

export const InstallAppBranch = ({
  appliedConfigId,
  appliedRun,
  appliedRunHref,
  branchConfig,
  branchHref,
  branchName,
  history,
  isLoading,
  latestRun,
  latestRunHref,
}: IInstallAppBranch) => {
  if (isLoading && !branchName) {
    return <Skeleton height="180px" width="100%" />
  }

  if (!branchName) {
    return (
      <EmptyState
        variant="history"
        emptyTitle="No app branch connected"
        emptyMessage="Connect this install to an app branch to track its applied configuration."
      />
    )
  }

  const vcs =
    branchConfig?.connected_github_vcs_config ??
    branchConfig?.public_git_vcs_config
  const repoHref = vcs?.repo
    ? (vcs.repo.startsWith('http')
        ? vcs.repo
        : `https://github.com/${vcs.repo}`
      ).replace(/\.git\/?$/, '')
    : undefined
  const isCurrent = !!latestRun?.id && latestRun.id === appliedRun?.id

  return (
    <div className="flex flex-col gap-4">
      <Card className="!p-4 !gap-4">
        <div className="flex items-center justify-between gap-3">
          <Text variant="body" weight="strong">
            Tracking
          </Text>
          <Status status={isCurrent ? 'active' : 'pending'} variant="badge">
            {isCurrent ? 'Current' : 'Update available'}
          </Status>
        </div>
        <div className="flex flex-wrap gap-x-8 gap-y-4">
          <LabeledValue label="App branch">
            {branchHref ? (
              <Link href={branchHref} textVariant="subtext">
                {branchName}
              </Link>
            ) : (
              <Text variant="subtext">{branchName}</Text>
            )}
          </LabeledValue>
          {appliedConfigId ? (
            <LabeledValue label="Applied config">
              <ID>{appliedConfigId}</ID>
            </LabeledValue>
          ) : null}
          {vcs?.repo ? (
            <LabeledValue label="Repository">
              {repoHref ? (
                <Link href={repoHref} isExternal textVariant="subtext">
                  {vcs.repo}
                </Link>
              ) : (
                <Text variant="subtext">{vcs.repo}</Text>
              )}
            </LabeledValue>
          ) : null}
          {vcs?.branch ? (
            <LabeledValue label="Git branch">
              <Text variant="subtext" family="mono">
                {vcs.branch}
              </Text>
            </LabeledValue>
          ) : null}
          {vcs?.directory && vcs.directory !== '.' ? (
            <LabeledValue label="Directory">
              <Text variant="subtext" family="mono">
                {vcs.directory}
              </Text>
            </LabeledValue>
          ) : null}
        </div>
      </Card>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <TrackedRunCard
          label="Expected / latest run"
          run={latestRun}
          href={latestRunHref}
          emptyMessage="This branch has not run yet."
        />
        <TrackedRunCard
          label="Currently applied"
          run={appliedRun}
          href={appliedRunHref}
          emptyMessage="No branch run has been applied to this install yet."
        />
      </div>

      {history ? (
        <div className="flex flex-col gap-3">
          <Text variant="body" weight="strong">
            Applied config history
          </Text>
          {history}
        </div>
      ) : null}
    </div>
  )
}
