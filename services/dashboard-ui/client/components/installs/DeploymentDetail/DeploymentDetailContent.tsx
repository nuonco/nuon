import { useLocation, useParams } from 'react-router'
import { useMemo } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import { useRunTemplateBuilds } from '@/components/branches/BranchOverview/use-run-template-builds'
import { useBranchConfigSections } from '@/components/branches/BranchRunChangesSummary'
import {
  ConfigChangesLoading,
  ConfigChangesViewer,
  useConfigVersionLabel,
} from '@/components/branches/ConfigChanges'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Link } from '@/components/common/Link'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Text } from '@/components/common/Text'
import { WorkflowChangesSummaryContainer } from '@/components/workflows/WorkflowChangesSummary'
import { WorkflowAlertBanners } from '@/components/workflows/WorkflowDetails'
import { WorkflowInputChanges } from '@/components/workflows/WorkflowInputChanges'
import { WorkflowSteps } from '@/components/workflows/WorkflowSteps'
import { PolicyViolations } from '@/components/workflows/step-details/PolicyViolations'
import { WorkflowActionButtons } from '@/components/workflows/workflow-details/WorkflowActionButtons'
import { useInstallLink, useInstallNested } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import { useRespondedApprovals } from '@/hooks/use-responded-approvals'
import { useWorkflow } from '@/hooks/use-workflow'
import { getBranchWorkflowRun } from '@/lib'
import { isWorkflowPath, workflowPageSuffix } from '@/lib/workflow-href'
import { AppProvider } from '@/providers/app-provider'
import type { TInstallDeploymentRecord, TWorkflow } from '@/types'
import { getPolicyViolationCounts } from '@/utils/workflow-utils'
import { humanize } from '@/utils/string-utils'
import { DeploymentConfigChanges } from './DeploymentConfigChanges'
import {
  deploymentChangeDescription,
  deploymentOutcomes,
  deploymentSteps,
  deploymentTemplateScope,
  isAwaitingDeploymentApproval,
} from './deployment-progress'

const DeploymentTemplateDiff = ({
  appId,
  branchId,
  branchRunId,
  deployment,
  workflow,
}: {
  appId: string
  branchId: string
  branchRunId: string
  deployment?: TInstallDeploymentRecord
  workflow?: TWorkflow
}) => {
  const { org } = useOrg()
  const runEnabled = !!org?.id
  const {
    data: branchRun,
    isPending: runPending,
    isPlaceholderData: runPlaceholder,
    isError: runError,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['branch-workflow-run', org?.id, appId, branchId, branchRunId],
    queryFn: () =>
      getBranchWorkflowRun({
        orgId: org!.id,
        appId,
        branchId,
        runId: branchRunId,
      }),
    enabled: runEnabled,
  })
  const scope = useMemo(
    () => deploymentTemplateScope(deployment, workflow),
    [deployment, workflow]
  )
  const { metaBuilds, templateBuilds, buildsPending } = useRunTemplateBuilds({
    orgId: org?.id,
    appId,
    branchId,
    branchRunId,
    workflowSteps: branchRun?.steps ?? [],
  })
  const scopedMetaBuilds = useMemo(
    () =>
      metaBuilds.filter((build) =>
        build.component_type === 'sandbox' || build.component_id === 'sandbox'
          ? scope.sections.includes('Sandbox')
          : scope.components.includes(
              build.component_name || build.component_id || ''
            )
      ),
    [metaBuilds, scope]
  )
  const scopedTemplateBuilds = useMemo(
    () =>
      templateBuilds.filter((build) =>
        build.kind === 'sandbox'
          ? scope.sections.includes('Sandbox')
          : scope.components.includes(build.name)
      ),
    [templateBuilds, scope]
  )
  const config = useBranchConfigSections({
    branchId,
    appBranchRunId: branchRunId,
    builds: scopedMetaBuilds,
    scope,
  })
  const versionLabel = useConfigVersionLabel({
    orgId: org?.id,
    appId,
    headConfigId: config.headConfigId,
    baseConfigId: config.baseConfigId,
  })
  const commit = config.comparison?.head_run?.vcs_connection_commit

  if (config.isError)
    return (
      <EmptyState
        emptyTitle="Template updates failed to load"
        emptyMessage="Unable to load this deployment's changes."
      />
    )
  if (
    config.isLoading ||
    (runEnabled && !runError && (runPending || runPlaceholder)) ||
    buildsPending
  )
    return <ConfigChangesLoading />
  return (
    <ConfigChangesViewer
      sections={config.sections}
      builds={scopedTemplateBuilds}
      versionLabel={versionLabel}
      previousSha={config.previousSha}
      sha={config.sha}
      message={commit?.message}
      author={commit?.author_name}
      createdAt={config.comparison?.head_run?.created_at}
    />
  )
}

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
  const branch = deployment?.app_branch
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
  if (!branch?.id || !branch.run_id || !appId)
    return <DeploymentConfigChanges deployment={deployment} />
  return (
    <AppProvider appId={appId}>
      <DeploymentTemplateDiff
        appId={appId}
        branchId={branch.id}
        branchRunId={branch.run_id}
        deployment={deployment}
        workflow={workflow}
      />
    </AppProvider>
  )
}

export const DeploymentAlerts = () => {
  const { workflow, failedSteps } = useWorkflow()
  const { hasResponded } = useRespondedApprovals()
  const steps = deploymentSteps(workflow)
  const approvals = steps.filter(
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
      {steps
        .filter((step) => {
          const { hasViolations, skipFindings } = getPolicyViolationCounts(step)
          return (
            (hasViolations || skipFindings.length > 0) &&
            !failedSteps.some((failed) => failed.id === step.id)
          )
        })
        .map((step) => (
          <div key={step.id} className="flex shrink-0 flex-col gap-2">
            <Text variant="subtext" weight="strong">
              {step.name}
            </Text>
            <PolicyViolations step={step} />
          </div>
        ))}
    </>
  )
}

export const DeploymentWorkflowContent = ({
  installId: installIdProp,
}: {
  installId?: string
} = {}) => {
  const { workflow } = useWorkflow()
  const { org } = useOrg()
  const params = useParams()
  const { pathname } = useLocation()
  const nested = useInstallNested()
  const installLink = useInstallLink()
  const installId =
    installIdProp ||
    params.installId ||
    workflow.metadata?.install_id ||
    workflow.owner_id
  const name = workflow.name || humanize(workflow.type) || 'Workflow'
  const href =
    org?.id && installId && workflow.id
      ? installLink({
          orgId: org.id,
          appId: params.appId,
          installId,
          suffix: workflowPageSuffix(workflow.id, nested),
        })
      : ''
  const onWorkflowPage = isWorkflowPath(pathname, workflow.id)
  const hasInputChanges = !!workflow.metadata?.changed_input_values

  return (
    <div className="flex flex-col gap-4">
      {!onWorkflowPage && href ? (
        <Link href={href} textVariant="base">
          {name}
        </Link>
      ) : null}
      {hasInputChanges ? (
        <>
          <SectionHeader
            title="Input changes"
            description="Values changed by this update."
          />
          <WorkflowInputChanges workflow={workflow} empty />
        </>
      ) : null}
      <SectionHeader
        title="Workflow steps"
        description={
          workflow.status?.status === 'error'
            ? undefined
            : workflow.status?.status_human_description
        }
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
