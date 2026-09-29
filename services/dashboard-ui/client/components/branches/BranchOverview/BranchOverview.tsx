import type { ReactNode } from 'react'
import { CompositeError } from '@/components/common/CompositeError'
import { Link } from '@/components/common/Link'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TCompositeError } from '@/types'
import { OverviewLoadingTrack } from './OverviewLoadingTrack'
import type { TOverviewStage } from './overview-loading'
import { RolloutTiles } from './RolloutTiles'
import type { TTrackGroup } from './RolloutTrack'
import { RunSourceCard, type IRunSourceCard } from './RunSourceCard'

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
  isLoading?: boolean
  rollout?: TOverviewRollout
  changes?: ReactNode
  groups: TTrackGroup[]
  rolloutHref: string
  onSelectGroup: (groupId: string) => void
  loadingStages?: TOverviewStage[]
  compositeError?: TCompositeError
  failedBuilds?: TFailedBuildLink[]
}

export const BranchOverview = ({
  hasPlan,
  isLoading,
  rollout,
  changes,
  groups,
  rolloutHref,
  onSelectGroup,
  loadingStages,
  compositeError,
  failedBuilds,
}: IBranchOverview) => (
  <div className="flex flex-col gap-10 p-4 md:p-6">
    {loadingStages?.length ? (
      <OverviewLoadingTrack stages={loadingStages} />
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

    {rollout ? (
      <section className="grid items-start gap-6 lg:grid-cols-[20rem_minmax(0,1fr)]">
        <RunSourceCard
          source={rollout.source}
          title={rollout.title}
          sha={rollout.sha}
          shaUrl={rollout.shaUrl}
          author={rollout.author}
          status={rollout.status}
          commit={rollout.commit}
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

    {hasPlan && !isLoading ? (
      <section className="flex flex-col gap-3">
        <SectionHeader
          title="Installs"
          description={rollout?.activity}
          actions={<Link href={rolloutHref}>View rollout</Link>}
        />
        <RolloutTiles groups={groups} onSelectGroup={onSelectGroup} />
      </section>
    ) : null}
  </div>
)
