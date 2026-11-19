import { GroupApprovalActions } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep/GroupApprovalActions'
import {
  PlanInstallChanges,
  type TPlanInstallFacts,
} from '@/components/branches/PlanInstallChanges'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import type { TInstallWorkflowStep } from '@/types'
import { BranchRunApproval } from '../BranchRunApproval'
import { getGroupName } from '../use-group-plan-href'

const awaitingApproval = (step: TInstallWorkflowStep, workflowId?: string) =>
  step.execution_type === 'approval' &&
  step.status?.status === 'approval-awaiting' &&
  !step.approval?.response &&
  !!step.approval?.id &&
  !!step.id &&
  !!workflowId

export interface IPlanDiffPanel extends IPanel {
  step: TInstallWorkflowStep
  orgId: string
  workflowId?: string
  installFacts?: TPlanInstallFacts
}

export const PlanDiffPanel = ({
  step,
  orgId,
  workflowId,
  installFacts,
  ...props
}: IPlanDiffPanel) => {
  const groupName = getGroupName(step.name)
  const showApproval = awaitingApproval(step, workflowId)

  return (
    <Panel {...props} size="3/4" heading={`Plan for ${groupName}`}>
      <div className="flex flex-col gap-6">
        {showApproval ? (
          <BranchRunApproval
            items={[
              {
                key: step.id!,
                groupName,
                actions: (
                  <GroupApprovalActions
                    target={{
                      orgId,
                      workflowId: workflowId!,
                      stepId: step.id!,
                      approvalId: step.approval!.id!,
                      groupName,
                    }}
                  />
                ),
              },
            ]}
          />
        ) : null}
        <PlanInstallChanges step={step} installFacts={installFacts} />
      </div>
    </Panel>
  )
}
