import { useSearchParams } from 'react-router'
import { Link } from '@/components/common/Link'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { PageTitle } from '@/components/navigation/PageTitle'
import { RolloutTrack } from './RolloutTrack'
import { useRolloutGroups } from './use-rollout-groups'

export const BranchRolloutContainer = () => {
  const [searchParams] = useSearchParams()
  const groupId = searchParams.get('group') ?? undefined
  const { app, branch, basePath, rollout, groups, hasPlan, isLoading } =
    useRolloutGroups()

  return (
    <div className="flex flex-col gap-3 p-4 md:p-6">
      <PageTitle segments={['Rollout', branch?.name, app?.name]} />
      <SectionHeader
        title="Rollout"
        description={rollout?.activity}
        actions={rollout ? <Link href={rollout.href}>View run</Link> : null}
      />
      {isLoading ? (
        <Loading />
      ) : !hasPlan ? (
        <Text variant="subtext" theme="neutral">
          This branch has no install groups yet. Every install updates at once.{' '}
          <Link href={`${basePath}/settings`}>Create a deployment plan</Link>
        </Text>
      ) : (
        <RolloutTrack
          key={`${rollout?.id}-${groupId}`}
          groups={groups}
          initialGroupId={groupId}
          emptyMessage="No runs yet. Push a commit or start a run."
        />
      )}
    </div>
  )
}
