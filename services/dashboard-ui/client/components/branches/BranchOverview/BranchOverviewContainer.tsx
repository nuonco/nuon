import { useMemo } from 'react'
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
  TemplateChangesContent,
  useConfigVersionLabel,
} from '@/components/branches/ConfigChanges'
import { BuildPanel } from '@/components/builds/BuildPanel'
import { CancelWorkflowButton } from '@/components/workflows/CancelWorkflow'
import {
  InstallDeploymentPanel,
  PendingInstallPanel,
} from '@/components/installs/InstallDeploymentPanel'
import { stepStatusCategory } from '@/components/branches/shared/step-status'
import { useSurfaces } from '@/hooks/use-surfaces'
import { previewModeDisplayLabel } from '@/components/branches/shared/preview-mode'
import { commitUrl } from './run-source'
import type { TAPIError } from '@/types'
import { BranchOverview, type TFailedBuildLink } from './BranchOverview'
import { BranchOverviewV2 } from './BranchOverviewV2'
import {
  buildOverviewLoadingStages,
  groupCompositeError,
  preRolloutCompositeError,
} from './overview-loading'
import { useGroupPlanApprovals } from '@/components/branches/BranchRunApproval/use-group-plan-approvals'
import { WarmPlanInstallChanges } from '@/components/branches/PlanInstallChanges/WarmPlanInstallChanges'
import { useGroupApprovedVersions } from './use-group-approved-versions'
import { useRolloutGroups } from './use-rollout-groups'
import { mergedChangedBuildRows } from './changed-builds'
import {
  isTerminalRunStatus,
  useRunTemplateBuilds,
} from './use-run-template-builds'

export const BranchOverviewContainer = ({
  variant = 'current',
}: {
  variant?: 'current' | 'v2'
}) => {
  const Overview = variant === 'v2' ? BranchOverviewV2 : BranchOverview
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
    rolloutWorkflow,
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
  const approvedVersions = useGroupApprovedVersions({
    orgId,
    appId,
    branchId,
    repoSlug,
    enabled: variant === 'v2',
  })

  const {
    builds,
    buildsPending,
    metaBuilds,
    runBuilds,
    sandboxBuildId,
    templateBuilds,
  } = useRunTemplateBuilds({
    orgId,
    appId,
    branchId,
    branchRunId,
    workflowSteps,
    isRunTerminal: isTerminalRunStatus(rolloutWorkflow?.status?.status),
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
  const buildRows = useMemo(
    () =>
      mergedChangedBuildRows({
        metaBuilds,
        runBuilds,
        orgId,
        appId,
        sandboxBuildId,
        components: configSections.sections.find(
          (section) => section.sectionKey === 'components'
        )?.entities,
      }),
    [
      metaBuilds,
      runBuilds,
      orgId,
      appId,
      sandboxBuildId,
      configSections.sections,
    ]
  )
  const configPending =
    !!orgId &&
    !!appId &&
    !!branchId &&
    !!branchRunId &&
    configSections.comparison === undefined &&
    !configSections.isError
  const buildRowsLoading =
    buildRows.length === 0 && (buildsPending || configPending)
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
  const changesProps = {
    sections: configSections.sections,
    builds: templateBuilds,
    versionLabel,
    previousSha: configSections.previousSha,
    sha: configSections.sha ?? rollout?.sha,
    message: rollout?.commit?.message,
    author: rollout?.commit?.author ?? rollout?.author,
    createdAt: rollout?.commit?.createdAt,
    source: rollout?.source,
    isPending: changesPending,
    isLoading: configSections.isLoading || buildsPending,
    isError: configSections.isError,
  }
  const configParseFailure = configError ? (
    <ConfigParseFailure
      className="w-full"
      lines={configDiagnosticLines(configError)}
    />
  ) : null
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
      <WarmPlanInstallChanges steps={workflowSteps} />
      <PageTitle
        segments={
          pinnedWorkflowId
            ? [rollout?.title ?? 'Run', app?.name]
            : [branch?.name, app?.name]
        }
      />
      <Overview
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
          configParseFailure ??
          (branchRunId ? <TemplateChangesButton {...changesProps} /> : null)
        }
        configChanges={
          configParseFailure ??
          (branchRunId ? <TemplateChangesContent {...changesProps} /> : null)
        }
        configChangesKey={[
          configError ? 'parse-error' : '',
          configSections.sections.length,
          templateBuilds?.length ?? 0,
          versionLabel,
          changesProps.sha,
          changesPending,
          configSections.isLoading || buildsPending,
          configSections.isError,
        ].join(':')}
        groups={groupsWithErrors}
        loadingStages={loadingStages}
        buildRows={buildRows}
        buildRowsLoading={buildRowsLoading}
        compositeError={failure}
        failedBuilds={failedBuilds}
        rolloutHref={rolloutHref}
        groupHref={groupHref}
        approvals={approvals}
        versionLabel={versionLabel}
        previousSha={configSections.previousSha}
        approvedVersions={approvedVersions.versions}
        approvedVersionsLoading={approvedVersions.isLoading}
        onSelectInstall={(install) => {
          if (!install.id || install.id.startsWith('queued-')) return
          addPanel(
            install.workflowId ? (
              <InstallDeploymentPanel
                orgId={orgId}
                appId={appId}
                installId={install.id}
                workflowId={install.workflowId}
                title={install.name}
                repo={repoSlug}
                initialTab="workflow"
              />
            ) : (
              <PendingInstallPanel title={install.name} status={install.status} />
            )
          )
        }}
        onSelectBuild={(build) =>
          build.componentId
            ? addPanel(
                <BuildPanel
                  componentId={build.componentId}
                  buildId={build.id}
                  title={build.name}
                />
              )
            : undefined
        }
        cancelRollout={
          rolloutWorkflow?.id ? (
            <CancelWorkflowButton workflow={rolloutWorkflow} size="sm">
              Cancel rollout
            </CancelWorkflowButton>
          ) : null
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
