import { useNavigate } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Link } from '@/components/common/Link'
import { PageTitle } from '@/components/navigation/PageTitle'
import { BranchRunChanges } from '@/components/branches/BranchRunChanges'
import { stepStatusCategory } from '@/components/branches/shared/step-status'
import { getBranchRunBuilds } from '@/lib'
import { BranchOverview, type TFailedBuildLink } from './BranchOverview'
import {
  buildOverviewLoadingStages,
  overviewCompositeError,
} from './overview-loading'
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
    workflowSteps,
    showLoadingTrack,
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
  const failedBuilds: TFailedBuildLink[] = (builds ?? []).flatMap((build) => {
    const status = build.status_v2?.status || build.status
    if (stepStatusCategory(status) !== 'error') return []
    if (!orgId || !appId || !build.component_id || !build.id) return []
    return [
      {
        id: build.id,
        name: build.component_name || build.component_id,
        href: `/${orgId}/apps/${appId}/components/${build.component_id}/builds/${build.id}`,
      },
    ]
  })
  const build = builds?.find((item) => item.component_id && item.id)
  const buildsHref =
    failedBuilds.length === 0 && build
      ? `${basePath}/components/${build.component_id}/builds/${build.id}`
      : undefined
  const loadingStages = showLoadingTrack
    ? buildOverviewLoadingStages({ steps: workflowSteps, sha: rollout?.sha })
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
        loadingStages={loadingStages}
        compositeError={overviewCompositeError(workflowSteps)}
        failedBuilds={failedBuilds}
        rolloutHref={`${basePath}/rollout`}
        onSelectGroup={(groupId) =>
          navigate(`${basePath}/rollout?group=${encodeURIComponent(groupId)}`)
        }
      />
    </>
  )
}
