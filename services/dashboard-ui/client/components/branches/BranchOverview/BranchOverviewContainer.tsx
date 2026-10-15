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
  useConfigVersionLabel,
} from '@/components/branches/ConfigChanges'
import { InstallDeploymentPanel } from '@/components/installs/InstallDeploymentPanel'
import { stepStatusCategory } from '@/components/branches/shared/step-status'
import { useSurfaces } from '@/hooks/use-surfaces'
import { previewModeDisplayLabel } from '@/components/branches/shared/preview-mode'
import { commitUrl } from './run-source'
import type { TAPIError } from '@/types'
import { BranchOverview, type TFailedBuildLink } from './BranchOverview'
import {
  buildOverviewLoadingStages,
  groupCompositeError,
  preRolloutCompositeError,
} from './overview-loading'
import { useGroupPlanApprovals } from '@/components/branches/BranchRunApproval/use-group-plan-approvals'
import { useRolloutGroups } from './use-rollout-groups'
import { useRunTemplateBuilds } from './use-run-template-builds'

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

  const { builds, metaBuilds, templateBuilds } = useRunTemplateBuilds({
    orgId,
    appId,
    branchId,
    branchRunId,
    workflowSteps,
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
  const configSections = useBranchConfigSections({
    branchId,
    appBranchRunId: branchRunId,
    builds: metaBuilds,
  })
  const versionLabel = useConfigVersionLabel({
    orgId,
    appId,
    headConfigId: configSections.headConfigId,
    baseConfigId: configSections.baseConfigId,
  })
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
  const groupsWithErrors = groups.map((group) => {
    const error = groupCompositeError(
      workflowSteps,
      group,
      branchRun?.composite_error
    )
    return error ? { ...group, error } : group
  })

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
              className="w-full"
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
        groups={groupsWithErrors}
        loadingStages={loadingStages}
        compositeError={failure}
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
