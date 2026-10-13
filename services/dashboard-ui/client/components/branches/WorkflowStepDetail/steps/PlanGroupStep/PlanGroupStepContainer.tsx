import { useApp } from '@/hooks/use-app'
import type { TInstallWorkflowStep } from '@/types'
import { PlanGroupStep } from './PlanGroupStep'
import { GroupApprovalActions } from './GroupApprovalActions'
import { usePlanGroupInstalls } from './use-plan-group-installs'

interface IPlanGroupStepContainer {
  step: TInstallWorkflowStep
  metadata: Record<string, any>
  workflowStatus?: string
  hideHeading?: boolean
  diffOnly?: boolean
  onSelectInstall?: (installId: string) => void
  installFacts?: Record<
    string,
    {
      labels?: Record<string, string>
      region?: string
      status?: string
      detail?: string
      appliedConfigId?: string
    }
  >
}

export const PlanGroupStepContainer = ({
  step,
  metadata,
  workflowStatus,
  hideHeading,
  diffOnly,
  onSelectInstall,
  installFacts,
}: IPlanGroupStepContainer) => {
  const { labelColors } = useApp()
  const { orgId, approvalId, groupName, installs } = usePlanGroupInstalls(
    step,
    metadata
  )

  const hasApproval = step.execution_type === 'approval' && !!approvalId
  const hasResponse = !!step.approval?.response
  const isAwaiting = step.status?.status === 'approval-awaiting'
  const isCancelled =
    workflowStatus === 'cancelled' || step.status?.status === 'cancelled'
  const showApproveBar =
    !diffOnly && hasApproval && isAwaiting && !hasResponse && !isCancelled

  return (
    <PlanGroupStep
      installs={installs}
      groupName={groupName}
      labelColors={labelColors}
      orgId={orgId}
      hasResponse={diffOnly ? false : hasResponse}
      responseType={step.approval?.response?.type}
      showApproveBar={showApproveBar}
      isInProgress={step.status?.status === 'in-progress'}
      hideHeading={hideHeading}
      onSelectInstall={onSelectInstall}
      installFacts={installFacts}
      actions={
        showApproveBar ? (
          <GroupApprovalActions
            target={{
              orgId,
              workflowId: step.install_workflow_id,
              stepId: step.id,
              approvalId: approvalId!,
              groupName: groupName || 'install group',
            }}
          />
        ) : undefined
      }
    />
  )
}
