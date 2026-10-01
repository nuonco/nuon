import { useEffect, useMemo, useRef } from 'react'
import { useNavigate } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { PageTitle } from '@/components/navigation/PageTitle'
import { BranchRunChangesSummary } from '@/components/branches/BranchRunChangesSummary'
import { stepStatusCategory } from '@/components/branches/shared/step-status'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getBranchRunBuilds } from '@/lib'
import { BranchOverview, type TFailedBuildLink } from './BranchOverview'
import { changedBuildRows, type TBuildMeta } from './changed-builds'
import {
  buildOverviewLoadingStages,
  installFailureHref,
  overviewCompositeError,
} from './overview-loading'
import { RunBuildsPanel } from './RunBuildsPanel'
import { useRolloutGroups } from './use-rollout-groups'

const isBuildStep = (name?: string) =>
  !!name && /build/i.test(name) && !/(?:fetch|sync) app config/i.test(name)

const NO_META_BUILDS: TBuildMeta[] = []

export const BranchOverviewContainer = () => {
  const navigate = useNavigate()
  const { addPanel, updatePanel, panels } = useSurfaces()
  const buildsPanelId = useRef<string | null>(null)
  const openBuildsPanelId =
    panels.find((panel) => panel?.id === buildsPanelId.current)?.id ?? null
  const {
    app,
    branch,
    orgId,
    appId,
    branchId,
    basePath,
    branchRunId,
    branchRun,
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
  const buildStep = workflowSteps.find((step) => isBuildStep(step.name))
  const buildMetadata = (buildStep?.status?.metadata ?? {}) as {
    builds?: TBuildMeta[]
    sandbox_build_id?: string
  }
  const metaBuilds = buildMetadata.builds ?? NO_META_BUILDS
  const changedBuilds = useMemo(
    () =>
      changedBuildRows({
        metaBuilds,
        runBuilds: (builds ?? []).map((build) => ({
          id: build.id,
          component_id: build.component_id,
          component_name: build.component_name,
          status: build.status_v2?.status || build.status,
        })),
        orgId,
        appId,
        sandboxBuildId: buildMetadata.sandbox_build_id,
      }),
    [metaBuilds, builds, orgId, appId, buildMetadata.sandbox_build_id]
  )
  const hasBuilds = metaBuilds.length > 0 || (builds?.length ?? 0) > 0

  useEffect(() => {
    if (!openBuildsPanelId) return
    updatePanel(
      openBuildsPanelId,
      <RunBuildsPanel rows={changedBuilds} />
    )
  }, [openBuildsPanelId, changedBuilds, updatePanel])

  const openBuilds = () => {
    if (openBuildsPanelId) return
    buildsPanelId.current = addPanel(<RunBuildsPanel rows={changedBuilds} />)
  }

  const loadingStages = showLoadingTrack
    ? buildOverviewLoadingStages({
        steps: workflowSteps,
        sha: rollout?.sha,
        groupStatuses: groups.map((group) => group.status),
      })
    : undefined
  const appConfigStage = loadingStages?.find(
    (stage) => stage.id === 'app-config'
  )
  const changesPending =
    appConfigStage?.status === 'pending' ||
    appConfigStage?.status === 'in-progress'
  const compositeError = overviewCompositeError(
    workflowSteps,
    branchRun?.composite_error
  )

  return (
    <>
      <PageTitle segments={[branch?.name, app?.name]} />
      <BranchOverview
        hasPlan={hasPlan}
        isLoading={isLoading}
        rollout={rollout}
        changes={
          branchRunId ? (
            <BranchRunChangesSummary
              branchId={branchId}
              appBranchRunId={branchRunId}
              title="Template and source changes"
              isPending={changesPending}
              headerAction={
                hasBuilds ? (
                  <Button size="sm" onClick={openBuilds}>
                    View builds
                  </Button>
                ) : null
              }
            />
          ) : null
        }
        groups={groups}
        loadingStages={loadingStages}
        compositeError={compositeError}
        installWorkflowHref={installFailureHref(compositeError, orgId)}
        failedBuilds={failedBuilds}
        rolloutHref={`${basePath}/rollout`}
        onSelectGroup={(groupId) =>
          navigate(`${basePath}/rollout?group=${encodeURIComponent(groupId)}`)
        }
      />
    </>
  )
}
