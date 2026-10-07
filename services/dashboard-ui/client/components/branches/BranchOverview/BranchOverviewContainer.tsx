import { useEffect, useMemo, useRef } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { ProviderError } from '@/components/layout/ProviderError'
import { PageTitle } from '@/components/navigation/PageTitle'
import { BranchRunChangesSummary } from '@/components/branches/BranchRunChangesSummary'
import { stepStatusCategory } from '@/components/branches/shared/step-status'
import { useSurfaces } from '@/hooks/use-surfaces'
import { previewModeDisplayLabel } from '@/components/branches/shared/preview-mode'
import { getBranchRunBuilds, getBranchRunComparison } from '@/lib'
import { commitUrl } from './run-source'
import type { TAPIError } from '@/types'
import { BranchOverview, type TFailedBuildLink } from './BranchOverview'
import { changedBuildRows, type TBuildMeta } from './changed-builds'
import {
  buildOverviewLoadingStages,
  installFailureHref,
  overviewCompositeError,
} from './overview-loading'
import { RunBuildsPanel } from './RunBuildsPanel'
import { useGroupPlanApprovals } from '@/components/branches/BranchRunApproval/use-group-plan-approvals'
import { useRolloutGroups } from './use-rollout-groups'

const isBuildStep = (name?: string) =>
  !!name && /build/i.test(name) && !/(?:fetch|sync) app config/i.test(name)

const NO_META_BUILDS: TBuildMeta[] = []

export const BranchOverviewContainer = () => {
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
    pinnedWorkflowId,
    rolloutHref,
    groupHref,
    rolloutError,
    branchRunId,
    branchRun,
    rollout,
    workflowSteps,
    showLoadingTrack,
    groups,
    hasPlan,
    showInstalls,
    previewMode,
    repoSlug,
    isLoading,
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

  const { data: comparison } = useQuery({
    queryKey: [
      'branch-run-comparison',
      orgId,
      appId,
      branchId,
      branchRunId,
      'config',
    ],
    queryFn: () =>
      getBranchRunComparison({
        orgId: orgId!,
        appId: appId!,
        branchId,
        runId: branchRunId!,
        includeDiff: ['config'],
      }),
    enabled: !!previewMode && !!orgId && !!appId && !!branchRunId,
    placeholderData: keepPreviousData,
    retry: 1,
  })
  const baseSha = comparison?.base_run?.vcs_connection_commit?.sha
  const baseline = previewMode
    ? comparison
      ? {
          sha: baseSha,
          shaUrl: commitUrl(repoSlug, baseSha),
          runHref: comparison.base_run?.workflow_id
            ? `/${orgId}/apps/${appId}/branches/${branchId}/runs/${comparison.base_run.workflow_id}`
            : undefined,
        }
      : undefined
    : undefined

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
    updatePanel(openBuildsPanelId, <RunBuildsPanel rows={changedBuilds} />)
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
        previewMode,
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

  if (pinnedWorkflowId && rolloutError && !rollout) {
    return (
      <>
        <PageTitle segments={['Run', app?.name]} />
        <ProviderError error={rolloutError as TAPIError} />
      </>
    )
  }

  return (
    <>
      <PageTitle
        segments={
          pinnedWorkflowId
            ? [rollout?.title ?? 'Run', app?.name]
            : [branch?.name, app?.name]
        }
      />
      <BranchOverview
        hasPlan={hasPlan}
        showInstalls={showInstalls}
        showRolloutLink={!previewMode}
        previewMode={previewMode}
        isLoading={isLoading}
        rollout={
          rollout
            ? {
                ...rollout,
                previewMode: previewMode
                  ? previewModeDisplayLabel(previewMode)
                  : undefined,
                baseline,
              }
            : undefined
        }
        changes={
          branchRunId ? (
            <BranchRunChangesSummary
              branchId={branchId}
              appBranchRunId={branchRunId}
              builds={metaBuilds}
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
        rolloutHref={rolloutHref}
        groupHref={groupHref}
        approvals={approvals}
      />
    </>
  )
}
