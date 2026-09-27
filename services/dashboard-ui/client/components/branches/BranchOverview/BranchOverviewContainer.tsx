import { useNavigate } from 'react-router'
import { PageTitle } from '@/components/navigation/PageTitle'
import { BranchRunChanges } from '@/components/branches/BranchRunChanges'
import { BranchOverview } from './BranchOverview'
import { useRolloutGroups } from './use-rollout-groups'

export const BranchOverviewContainer = () => {
  const navigate = useNavigate()
  const {
    app,
    branch,
    branchId,
    basePath,
    repoSlug,
    branchRunId,
    rollout,
    groups,
    hasPlan,
    isLoading,
  } = useRolloutGroups()

  return (
    <>
      <PageTitle segments={[branch?.name, app?.name]} />
      <BranchOverview
        hasPlan={hasPlan}
        isLoading={isLoading}
        rollout={rollout}
        changes={
          branchRunId ? (
            <BranchRunChanges
              branchId={branchId}
              appBranchRunId={branchRunId}
              repoSlug={repoSlug}
              showRunComparison={false}
            />
          ) : null
        }
        groups={groups}
        rolloutHref={`${basePath}/rollout`}
        onSelectGroup={(groupId) =>
          navigate(`${basePath}/rollout?group=${encodeURIComponent(groupId)}`)
        }
      />
    </>
  )
}
