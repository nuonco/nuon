import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { CompositeError } from '@/components/common/CompositeError'
import { Link } from '@/components/common/Link'
import { Loading } from '@/components/common/Loading'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import {
  CommitRange,
  InstallGroupCards,
} from '@/components/branches/InstallGroupCards'
import type { TCompositeError } from '@/types'
import { OverviewLoadingTrack } from './OverviewLoadingTrack'
import type { TOverviewStage, TPreviewProgress } from './overview-loading'
import type { IGroupPlanApproval } from './RolloutGroupsCard'
import type { TTrackGroup, TTrackInstall } from './RolloutTrack'
import {
  RunCommitSummary,
  RunSourceMark,
  type IRunSourceCard,
} from './RunSourceCard'

export const InstallFailureNotice = ({
  error,
  href,
}: {
  error: TCompositeError
  href?: string
}) => (
  <div className="flex flex-col gap-2">
    <CompositeError error={error} />
    {href ? <Link href={href}>View install workflow</Link> : null}
  </div>
)

export interface TOverviewRollout extends IRunSourceCard {
  id: string
  href: string
  activity?: string
}

export interface TFailedBuildLink {
  id: string
  name: string
  href: string
}

export interface IBranchOverview {
  hasPlan: boolean
  showInstalls?: boolean
  showRolloutLink?: boolean
  previewMode?: TPreviewProgress
  isLoading?: boolean
  rollout?: TOverviewRollout
  changes?: ReactNode
  groups: TTrackGroup[]
  rolloutHref: string
  groupHref: (groupId: string) => string
  loadingStages?: TOverviewStage[]
  compositeError?: TCompositeError
  failedBuilds?: TFailedBuildLink[]
  approvals?: IGroupPlanApproval[]
  runHeaderAction?: ReactNode
  versionLabel?: string
  previousSha?: string
  onSelectInstall?: (install: TTrackInstall) => void
}

export const BranchOverview = ({
  hasPlan,
  showInstalls,
  previewMode,
  isLoading,
  rollout,
  changes,
  groups,
  loadingStages,
  compositeError,
  failedBuilds,
  approvals,
  runHeaderAction,
  versionLabel,
  previousSha,
  onSelectInstall,
}: IBranchOverview) => {
  const commit = rollout
    ? {
        message: rollout.commit?.message,
        author: rollout.commit?.author ?? rollout.author,
        sha: rollout.commit?.sha ?? rollout.sha,
        previousSha,
        createdAt: rollout.commit?.createdAt,
      }
    : undefined

  return (
    <div className="flex flex-col gap-8 p-4 md:p-6">
      {rollout ? (
        <div className="flex flex-col gap-3">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <span className="flex flex-wrap items-center gap-3">
              <Status status={rollout.status} />
              {versionLabel ? (
                <Text variant="subtext" theme="neutral" family="mono">
                  {versionLabel}
                </Text>
              ) : null}
              {commit ? <CommitRange commit={commit} /> : null}
              {rollout.source ? (
                <RunSourceMark source={rollout.source} />
              ) : null}
              {runHeaderAction}
            </span>
            {changes}
          </div>
          {loadingStages?.length ? (
            <OverviewLoadingTrack stages={loadingStages} />
          ) : null}
        </div>
      ) : loadingStages?.length ? (
        <div className="flex items-center justify-between gap-4">
          <OverviewLoadingTrack stages={loadingStages} />
          {runHeaderAction}
        </div>
      ) : isLoading ? (
        <Loading />
      ) : null}

      {compositeError ? <CompositeError error={compositeError} /> : null}

      {failedBuilds?.length ? (
        <div className="flex flex-col gap-2">
          <Text variant="subtext" weight="strong">
            Failed builds
          </Text>
          <ul className="flex flex-col gap-1">
            {failedBuilds.map((build) => (
              <li key={build.id}>
                <Link href={build.href} textVariant="subtext">
                  <Text as="span" variant="subtext" family="mono">
                    {build.name}
                  </Text>
                </Link>
              </li>
            ))}
          </ul>
        </div>
      ) : null}

      {previewMode === 'build-only' ? (
        <Banner theme="info">This preview will not update any install.</Banner>
      ) : null}
      {previewMode === 'plan-only' ? (
        <Banner theme="info">
          This preview will plan the selected install. It will not roll out to
          the branch.
        </Banner>
      ) : null}

      {rollout ? (
        <RunCommitSummary
          source={rollout.source}
          title={rollout.title}
          sha={rollout.sha}
          shaUrl={rollout.shaUrl}
          author={rollout.author}
          status={rollout.status}
          commit={rollout.commit}
          previewMode={rollout.previewMode}
          baseline={rollout.baseline}
        />
      ) : !isLoading ? (
        <Text variant="subtext" theme="neutral">
          {hasPlan
            ? 'No runs yet. Push a commit or start a run.'
            : 'This branch has no install groups yet. Every install updates at once.'}
        </Text>
      ) : null}

      {(showInstalls ?? hasPlan) && !isLoading ? (
        groups.length ? (
          <InstallGroupCards
            groups={groups}
            commit={commit}
            approvals={approvals}
            onSelectInstall={onSelectInstall}
          />
        ) : (
          <Text variant="subtext" theme="neutral">
            No install groups in this run yet.
          </Text>
        )
      ) : null}
    </div>
  )
}
