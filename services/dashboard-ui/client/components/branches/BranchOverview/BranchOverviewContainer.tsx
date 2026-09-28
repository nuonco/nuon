import { useNavigate } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Link } from '@/components/common/Link'
import { PageTitle } from '@/components/navigation/PageTitle'
import { BranchRunChanges } from '@/components/branches/BranchRunChanges'
import { getBranchRunBuilds } from '@/lib'
import { BranchOverview } from './BranchOverview'
import { useRolloutGroups } from './use-rollout-groups'

export const BranchOverviewContainer = () => {
  const navigate = useNavigate()
  const {
    app,
    branch,
    orgId,
    appId,
    branchId,
    basePath,
    repoSlug,
    branchRunId,
    rollout,
    groups,
    hasPlan,
    isLoading,
  } = useRolloutGroups()

  const { data: builds } = useQuery({
    queryKey: ['branch-run-builds', orgId, appId, branchId, branchRunId],
    queryFn: () =>
      getBranchRunBuilds({
        orgId: orgId!,
        appId: appId!,
        branchId,
        runId: branchRunId!,
      }),
    enabled: !!orgId && !!appId && !!branchRunId,
    placeholderData: keepPreviousData,
  })
  const build = builds?.find((item) => item.component_id && item.id)
  const buildsHref = build
    ? `${basePath}/components/${build.component_id}/builds/${build.id}`
    : undefined

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
              title="What's changed"
              headerAction={
                buildsHref ? <Link href={buildsHref}>View builds</Link> : null
              }
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
