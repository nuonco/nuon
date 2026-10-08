import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { CompositeError } from '@/components/common/CompositeError'
import { Link } from '@/components/common/Link'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TCompositeError } from '@/types'
import { OverviewLoadingTrack } from './OverviewLoadingTrack'
import type { TOverviewStage, TPreviewProgress } from './overview-loading'
import {
  RolloutGroupsCard,
  type IGroupPlanApproval,
} from './RolloutGroupsCard'
import type { TTrackGroup } from './RolloutTrack'
import { RunSourceCard, type IRunSourceCard } from './RunSourceCard'

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
  installWorkflowHref?: string
  failedBuilds?: TFailedBuildLink[]
  approvals?: IGroupPlanApproval[]
  runHeaderAction?: ReactNode
}

export const BranchOverview = ({
  hasPlan,
  showInstalls,
  showRolloutLink = true,
  previewMode,
  isLoading,
  rollout,
  changes,
  groups,
  rolloutHref,
  groupHref,
  loadingStages,
  compositeError,
  installWorkflowHref,
  failedBuilds,
  approvals,
  runHeaderAction,
}: IBranchOverview) => (
  <div className="flex flex-col gap-10 p-4 md:p-6">
    {loadingStages?.length ? (
      <OverviewLoadingTrack stages={loadingStages} />
    ) : isLoading ? (
      <Loading />
    ) : null}

    {compositeError ? (
      <InstallFailureNotice error={compositeError} href={installWorkflowHref} />
    ) : null}

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
        This preview will plan the selected install. It will not roll out to the branch.
      </Banner>
    ) : null}

    {rollout ? (
      <section className="grid items-start gap-6 lg:grid-cols-2">
        <RunSourceCard
          source={rollout.source}
          title={rollout.title}
          sha={rollout.sha}
          shaUrl={rollout.shaUrl}
          author={rollout.author}
          status={rollout.status}
          commit={rollout.commit}
          previewMode={rollout.previewMode}
          baseline={rollout.baseline}
          headerAction={runHeaderAction}
        />
        <div className="min-w-0">{changes}</div>
      </section>
    ) : !isLoading ? (
      <Text variant="subtext" theme="neutral">
        {hasPlan
          ? 'No runs yet. Push a commit or start a run.'
          : 'This branch has no install groups yet. Every install updates at once.'}
      </Text>
    ) : null}

    {(showInstalls ?? hasPlan) && !isLoading ? (
      <section className="flex flex-col gap-3">
        <SectionHeader
          title="Installs"
          description={rollout?.activity}
          actions={
            showRolloutLink ? (
              <Link href={rolloutHref}>View rollout</Link>
            ) : undefined
          }
        />
        {groups.length ? (
          <RolloutGroupsCard
            groups={groups}
            groupHref={groupHref}
            approvals={approvals}
          />
        ) : (
          <Text variant="subtext" theme="neutral">
            No install groups in this run yet.
          </Text>
        )}
      </section>
    ) : null}
  </div>
)
