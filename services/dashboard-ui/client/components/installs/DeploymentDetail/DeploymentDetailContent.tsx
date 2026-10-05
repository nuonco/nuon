import { useMemo } from 'react'
import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import { BranchRunChanges } from '@/components/branches/BranchRunChanges/BranchRunChanges'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Text } from '@/components/common/Text'
import { WorkflowChangesSummaryContainer } from '@/components/workflows/WorkflowChangesSummary'
import { WorkflowAlertBanners } from '@/components/workflows/WorkflowDetails'
import { WorkflowSteps } from '@/components/workflows/WorkflowSteps'
import { WorkflowActionButtons } from '@/components/workflows/workflow-details/WorkflowActionButtons'
import { useRespondedApprovals } from '@/hooks/use-responded-approvals'
import { useWorkflow } from '@/hooks/use-workflow'
import { AppProvider } from '@/providers/app-provider'
import type { TInstallDeploymentRecord, TWorkflow } from '@/types'
import { DeploymentConfigChanges } from './DeploymentConfigChanges'
import {
  deploymentChangeDescription,
  deploymentOutcomes,
  deploymentSteps,
  isAwaitingDeploymentApproval,
} from './deployment-progress'

export const DeploymentTemplateContent = ({
  deployment,
  appId,
  workflow,
  isLoading = false,
}: {
  deployment?: TInstallDeploymentRecord
  appId?: string
  workflow?: TWorkflow
  isLoading?: boolean
}) => {
  const branch = deployment?.app_branch
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
  if (isLoading) return <Text loading loadingWidth={32} />
  return branch?.id && branch.run_id && appId ? (
    <AppProvider appId={appId}>
      <BranchRunChanges
        branchId={branch.id}
        appBranchRunId={branch.run_id}
        showRunComparison={false}
        title="Template updates"
        scope={scope}
      />
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
