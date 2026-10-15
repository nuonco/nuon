import { GroupActionButton } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep/GroupApprovalActions'
import {
  PlanInstallChanges,
  type TPlanInstallFacts,
} from '@/components/branches/PlanInstallChanges'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { useOrg } from '@/hooks/use-org'
import type { TInstallWorkflowStep } from '@/types'
import { BranchRunApproval } from './BranchRunApproval'
import { getGroupName } from './use-group-plan-href'

const awaitingApproval = (step: TInstallWorkflowStep, workflowId?: string) =>
  step.execution_type === 'approval' &&
  step.status?.status === 'approval-awaiting' &&
  !step.approval?.response &&
  !!step.approval?.id &&
  !!step.id &&
  !!(workflowId || step.install_workflow_id)

export const PlanDiffPanel = ({
  step,
  installFacts,
  workflowId,
  ...props
}: IPanel & {
  step: TInstallWorkflowStep
  installFacts?: TPlanInstallFacts
  workflowId?: string
}) => {
  const { org } = useOrg()
  const groupName = getGroupName(step.name)
  const resolvedWorkflowId = workflowId || step.install_workflow_id
  const showApproval = awaitingApproval(step, resolvedWorkflowId)

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
                  <GroupActionButton
                    action="approve"
                    target={{
                      orgId: org?.id ?? '',
                      workflowId: resolvedWorkflowId!,
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
