import { useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router'
import { CommitLink } from '@/components/common/GitReferenceLink'
import { Link } from '@/components/common/Link'
import { Loading } from '@/components/common/Loading'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { PageTitle } from '@/components/navigation/PageTitle'
import { type TOverviewRollout } from './BranchOverview'
import { useGroupPlanApprovals } from '@/components/branches/BranchRunApproval/use-group-plan-approvals'
import { RolloutGroupsCard } from './RolloutGroupsCard'
import { useRolloutGroups } from './use-rollout-groups'

const RolloutRunCard = ({ rollout }: { rollout: TOverviewRollout }) => {
  const message = rollout.commit?.message?.split('\n')[0]
  const sha = rollout.commit?.sha ?? rollout.sha
  const shaUrl = rollout.commit?.shaUrl ?? rollout.shaUrl
  const author = rollout.commit?.author ?? rollout.author
  return (
    <div className="flex flex-col gap-2 rounded-xl border bg-white px-4 py-3 shadow-sm dark:bg-dark-grey-900">
      <Text variant="body" weight="strong" className="break-words">
        {message || rollout.title}
      </Text>
      {sha || author ? (
        <span className="flex flex-wrap items-center gap-x-2 gap-y-1">
          {sha ? <CommitLink sha={sha} href={shaUrl} /> : null}
          {author ? (
            <Text variant="subtext" theme="neutral">
              {author}
            </Text>
          ) : null}
        </span>
      ) : null}
    </div>
  )
}

export const BranchRolloutContainer = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const legacyGroupId = searchParams.get('group')
  const legacyInstallId = searchParams.get('install')
  const {
    app,
    branch,
    basePath,
    rolloutHref,
    groupHref,
    rollout,
    groups,
    hasPlan,
    isLoading,
    workflowSteps,
  } = useRolloutGroups()
  const approvals = useGroupPlanApprovals(
    rollout?.id
      ? {
          id: rollout.id,
          status: { status: rollout.status },
          steps: workflowSteps,
        }
      : undefined,
    groups
  )

  useEffect(() => {
    if (!legacyGroupId) return
    const next = new URLSearchParams()
    if (legacyInstallId) next.set('install', legacyInstallId)
    const query = next.toString()
    navigate(
      `${rolloutHref}/groups/${encodeURIComponent(legacyGroupId)}${query ? `?${query}` : ''}`,
      { replace: true }
    )
  }, [legacyGroupId, legacyInstallId, navigate, rolloutHref])

  return (
    <div className="flex flex-col gap-3 p-4 md:p-6">
      <PageTitle segments={['Rollout', branch?.name, app?.name]} />
      <SectionHeader
        title="Rollout"
        description={rollout?.activity}
        actions={
          rollout ? (
            <span className="flex items-center gap-3">
              <Status status={rollout.status} />
              <Link href={rollout.href}>View run</Link>
            </span>
          ) : undefined
        }
      />
      {rollout ? <RolloutRunCard rollout={rollout} /> : null}
      {isLoading ? (
        <Loading />
      ) : !hasPlan ? (
        <Text variant="subtext" theme="neutral">
          This branch has no install groups yet. Every install updates at once.{' '}
          <Link href={`${basePath}/settings`}>Create a deployment plan</Link>
        </Text>
      ) : groups.length === 0 ? (
        <Text variant="subtext" theme="neutral">
          No runs yet. Push a commit or start a run.
        </Text>
      ) : (
        <RolloutGroupsCard
          groups={groups}
          groupHref={groupHref}
          approvals={approvals}
        />
      )}
    </div>
  )
}
