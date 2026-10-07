import { useMemo } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import {
  changedBuildRows,
  type TBuildMeta,
} from '@/components/branches/BranchOverview/changed-builds'
import { BuildChangeCards } from '@/components/branches/BranchOverview/RunBuildsPanel'
import { BranchRunChanges } from '@/components/branches/BranchRunChanges/BranchRunChanges'
import { BranchRunChangesSummary } from '@/components/branches/BranchRunChangesSummary'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Text } from '@/components/common/Text'
import { WorkflowChangesSummaryContainer } from '@/components/workflows/WorkflowChangesSummary'
import { WorkflowAlertBanners } from '@/components/workflows/WorkflowDetails'
import { WorkflowSteps } from '@/components/workflows/WorkflowSteps'
import { WorkflowActionButtons } from '@/components/workflows/workflow-details/WorkflowActionButtons'
import { useOrg } from '@/hooks/use-org'
import { useRespondedApprovals } from '@/hooks/use-responded-approvals'
import { useWorkflow } from '@/hooks/use-workflow'
import { getBranchRunBuilds, getBranchWorkflowRun } from '@/lib'
import { AppProvider } from '@/providers/app-provider'
import type { TInstallDeploymentRecord, TWorkflow } from '@/types'
import { DeploymentConfigChanges } from './DeploymentConfigChanges'
import {
  deploymentChangeDescription,
  deploymentOutcomes,
  deploymentSteps,
  isAwaitingDeploymentApproval,
} from './deployment-progress'

const isBuildStep = (name?: string) =>
  !!name && /build/i.test(name) && !/(?:fetch|sync) app config/i.test(name)

const NO_BUILDS: TBuildMeta[] = []

export const DeploymentTemplateContent = ({
  deployment,
  appId,
  workflow,
  isLoading = false,
  isError = false,
  onRetry,
}: {
  deployment?: TInstallDeploymentRecord
  appId?: string
  workflow?: TWorkflow
  isLoading?: boolean
  isError?: boolean
  onRetry?: () => void
}) => {
  const { org } = useOrg()
  const branch = deployment?.app_branch
  const ready = !!branch?.id && !!branch.run_id && !!appId && !!org?.id
  const { data: branchRun } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: [
      'branch-workflow-run',
      org?.id,
      appId,
      branch?.id,
      branch?.run_id,
    ],
    queryFn: () =>
      getBranchWorkflowRun({
        orgId: org!.id,
        appId: appId!,
        branchId: branch!.id,
        runId: branch!.run_id!,
      }),
    enabled: ready && !isLoading && !isError,
  })
  const { data: runBuilds } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['branch-run-builds', org?.id, appId, branch?.id, branch?.run_id],
    queryFn: () =>
      getBranchRunBuilds({
        orgId: org!.id,
        appId: appId!,
        branchId: branch!.id,
        runId: branch!.run_id!,
      }),
    enabled: ready && !isLoading && !isError,
  })
  const scope = useMemo(() => {
    const resources = deployment?.affected_resources
    const components = [
      ...new Set([
        ...(resources?.components ?? []),
        ...deploymentOutcomes(deployment, workflow)
          .filter((outcome) => outcome.category === 'Components')
          .map((outcome) => outcome.name),
      ]),
    ]
    return {
      sections: [
        ...(resources?.stack ? ['Stack', 'Permissions', 'Policies'] : []),
        ...(resources?.sandbox ? ['Sandbox'] : []),
        ...(components.length ? ['Components'] : []),
        ...(deployment?.type === 'install_config_update'
          ? ['Install inputs', 'Secrets']
          : []),
      ],
      components,
    }
  }, [deployment, workflow])
  const buildStep = branchRun?.steps?.find((step) => isBuildStep(step.name))
  const buildMetadata = (buildStep?.status?.metadata ?? {}) as {
    builds?: TBuildMeta[]
    sandbox_build_id?: string
  }
  const metaBuilds = useMemo(
    () =>
      (buildMetadata.builds ?? NO_BUILDS).filter((build) =>
        build.component_type === 'sandbox' || build.component_id === 'sandbox'
          ? scope.sections.includes('Sandbox')
          : scope.components.includes(
              build.component_name || build.component_id || ''
            )
      ),
    [buildMetadata.builds, scope]
  )
  const changedBuilds = useMemo(
    () =>
      changedBuildRows({
        metaBuilds,
        runBuilds: (runBuilds ?? [])
          .filter((build) =>
            scope.components.includes(
              build.component_name || build.component_id || ''
            )
          )
          .map((build) => ({
            id: build.id,
            component_id: build.component_id,
            component_name: build.component_name,
            status: build.status_v2?.status || build.status,
          })),
        orgId: org?.id,
        appId,
        sandboxBuildId: buildMetadata.sandbox_build_id,
      }),
    [
      metaBuilds,
      runBuilds,
      scope,
      org?.id,
      appId,
      buildMetadata.sandbox_build_id,
    ]
  )
  if (isLoading) return <Text loading loadingWidth={32} />
  if (isError)
    return (
      <EmptyState
        emptyTitle="Template updates failed to load"
        emptyMessage="Unable to load this deployment's changes."
        action={
          onRetry ? (
            <Button variant="secondary" onClick={onRetry}>
              Try again
            </Button>
          ) : undefined
        }
      />
    )
  return branch?.id && branch.run_id && appId ? (
    <AppProvider appId={appId}>
      <div className="flex flex-col gap-4">
        <BranchRunChangesSummary
          branchId={branch.id}
          appBranchRunId={branch.run_id}
          builds={metaBuilds}
          title="Template updates"
          scope={scope}
        />
        <BuildChangeCards rows={changedBuilds} />
        <BranchRunChanges
          branchId={branch.id}
          appBranchRunId={branch.run_id}
          showRunComparison={false}
          title="Source files"
          scope={scope}
        />
      </div>
    </AppProvider>
  ) : (
    <DeploymentConfigChanges deployment={deployment} />
  )
}

export const DeploymentAlerts = () => {
  const { workflow, failedSteps } = useWorkflow()
  const { hasResponded } = useRespondedApprovals()
  const approvals = deploymentSteps(workflow).filter(
    (step) =>
      isAwaitingDeploymentApproval(step) &&
      step.approval?.type &&
      !['approve-all', 'noop'].includes(step.approval.type) &&
      !hasResponded(step.id)
  )
  return (
    <>
      <WorkflowAlertBanners workflow={workflow} failedSteps={failedSteps} />
      {approvals.map((step) => (
        <ApprovalBanner key={step.id} step={step} />
      ))}
    </>
  )
}

export const DeploymentWorkflowContent = () => {
  const { workflow } = useWorkflow()
  return (
    <div className="flex flex-col gap-4">
      <SectionHeader
        title="Workflow steps"
        description={workflow.status?.status_human_description}
        actions={<WorkflowActionButtons />}
      />
      <WorkflowSteps
        approvalPrompt={workflow.approval_option === 'prompt'}
        planOnly={workflow.plan_only}
      />
    </div>
  )
}

export const DeploymentChangesContent = ({
  deployment,
}: {
  deployment?: TInstallDeploymentRecord
}) => {
  const { workflow } = useWorkflow()
  return (
    <div className="flex flex-col gap-4">
      <SectionHeader
        title="Change summary"
        description={deploymentChangeDescription(
          workflow,
          deploymentOutcomes(deployment, workflow)
        )}
      />
      <WorkflowChangesSummaryContainer />
    </div>
  )
}
