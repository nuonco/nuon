import { useMemo } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { ProviderError } from '@/components/layout/ProviderError'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useBranchConfigSections } from '@/components/branches/BranchRunChangesSummary'
import {
  configDiagnosticLines,
  isConfigValidationError,
} from '@/components/branches/BranchRunChangesSummary/config-diagnostics'
import { ConfigParseFailure } from '@/components/branches/BranchRunChangesSummary/ConfigParseFailure'
import {
  TemplateChangesButton,
  type TTemplateBuildChange,
} from '@/components/branches/ConfigChanges'
import { InstallDeploymentPanel } from '@/components/installs/InstallDeploymentPanel'
import { stepStatusCategory } from '@/components/branches/shared/step-status'
import { useSurfaces } from '@/hooks/use-surfaces'
import { previewModeDisplayLabel } from '@/components/branches/shared/preview-mode'
import { getAppConfig, getBranchRunBuilds, getSandboxBuild } from '@/lib'
import { vcsRepo } from '@/utils/vcs-urls'
import { commitUrl } from './run-source'
import type { TAPIError } from '@/types'
import { BranchOverview, type TFailedBuildLink } from './BranchOverview'
import { changedBuildRows, type TBuildMeta } from './changed-builds'
import {
  buildOverviewLoadingStages,
  preRolloutCompositeError,
} from './overview-loading'
import { useGroupPlanApprovals } from '@/components/branches/BranchRunApproval/use-group-plan-approvals'
import { useRolloutGroups } from './use-rollout-groups'

const isBuildStep = (name?: string) =>
  !!name && /build/i.test(name) && !/(?:fetch|sync) app config/i.test(name)

const NO_META_BUILDS: TBuildMeta[] = []

export const BranchOverviewContainer = () => {
  const { addPanel } = useSurfaces()
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
  const { data: sandboxBuild } = useQuery({
    queryKey: ['sandbox-build', orgId, appId, buildMetadata.sandbox_build_id],
    queryFn: () =>
      getSandboxBuild({
        orgId: orgId!,
        appId: appId!,
        buildId: buildMetadata.sandbox_build_id!,
      }),
    enabled: !!orgId && !!appId && !!buildMetadata.sandbox_build_id,
  })
  const templateBuilds = useMemo<TTemplateBuildChange[]>(
    () =>
      changedBuilds.map((row) => {
        const commit = sandboxBuild?.vcs_connection_commit
        return {
          ...row,
          kind: row.kind ?? 'component',
          commit:
            row.kind === 'sandbox' && commit?.sha
              ? {
                  sha: commit.sha,
                  message: commit.message,
                  author: commit.author_name,
                  repo: vcsRepo(app?.sandbox_config),
                }
              : undefined,
        }
      }),
    [changedBuilds, sandboxBuild, app?.sandbox_config]
  )
  const configSections = useBranchConfigSections({
    branchId,
    appBranchRunId: branchRunId,
    builds: metaBuilds,
  })
  const versionQuery = (configId?: string) => ({
    queryKey: ['app-config-version', orgId, appId, configId],
    queryFn: () =>
      getAppConfig({
        orgId: orgId!,
        appId: appId!,
        appConfigId: configId!,
      }),
    enabled: !!orgId && !!appId && !!configId,
    select: (config: { version?: number }) => config.version,
  })
  const { data: nextVersion } = useQuery(
    versionQuery(configSections.headConfigId)
  )
  const { data: previousVersion } = useQuery(
    versionQuery(configSections.baseConfigId)
  )
  const versionLabel =
    nextVersion == null
      ? undefined
      : previousVersion == null || previousVersion === nextVersion
        ? `v${nextVersion}`
        : `v${previousVersion} → v${nextVersion}`
  const comparison = configSections.comparison
  const baseSha = configSections.previousSha
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
  const failure = preRolloutCompositeError(
    workflowSteps,
    branchRun?.composite_error
  )
  const configError = isConfigValidationError(failure) ? failure : undefined
  const compositeError = configError ? undefined : failure

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
          configError ? (
            <ConfigParseFailure
              className="w-full basis-full"
              title="Changes"
              lines={configDiagnosticLines(configError)}
            />
          ) : branchRunId ? (
            <TemplateChangesButton
              sections={configSections.sections}
              builds={templateBuilds}
              versionLabel={versionLabel}
              previousSha={configSections.previousSha}
              sha={configSections.sha ?? rollout?.sha}
              message={rollout?.commit?.message}
              author={rollout?.commit?.author ?? rollout?.author}
              createdAt={rollout?.commit?.createdAt}
              source={rollout?.source}
              isPending={changesPending}
              isLoading={configSections.isLoading}
              isError={configSections.isError}
            />
          ) : null
        }
        groups={groups}
        loadingStages={loadingStages}
        compositeError={compositeError}
        failedBuilds={failedBuilds}
        rolloutHref={rolloutHref}
        groupHref={groupHref}
        approvals={approvals}
        versionLabel={versionLabel}
        previousSha={configSections.previousSha}
        onSelectInstall={(install) =>
          install.workflowId
            ? addPanel(
                <InstallDeploymentPanel
                  orgId={orgId}
                  appId={appId}
                  installId={install.id}
                  workflowId={install.workflowId}
                  title={install.name}
                  repo={repoSlug}
                />
              )
            : undefined
        }
        runHeaderAction={
          rollout?.id ? (
            <AdminDashboardLink path={`/workflows/${rollout.id}`} />
          ) : null
        }
      />
    </>
  )
}
