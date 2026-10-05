import { useOrg } from '@/hooks/use-org'
import type { TInstallWorkflow } from '@/types'
import { GroupActionButton } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep/GroupApprovalActions'
import { useOpenWorkflowRunPanel } from '@/components/branches/WorkflowRunPanel'
import {
  BranchRunApproval,
  type IBranchRunApprovalItem,
} from './BranchRunApproval'
import { getGroupName } from './use-group-plan-href'

interface IBranchRunApprovalContainer {
  run: TInstallWorkflow
}

export const BranchRunApprovalContainer = ({
  run,
}: IBranchRunApprovalContainer) => {
  const { org } = useOrg()
  const orgId = org?.id ?? ''
  const openWorkflowRunPanel = useOpenWorkflowRunPanel()

  if (run.status?.status === 'cancelled') {
    return null
  }

  const items: IBranchRunApprovalItem[] = (run.steps ?? [])
    .filter(
      (step) =>
        step.execution_type === 'approval' &&
        step.status?.status === 'approval-awaiting' &&
        !step.approval?.response &&
        !!step.approval?.id
    )
    .map((step) => {
      const groupName = getGroupName(step.name)
      return {
        key: step.id ?? step.approval!.id!,
        groupName,
        onReview: () => openWorkflowRunPanel(run.id ?? '', step.id),
        actions: (
          <GroupActionButton
            action="approve"
            target={{
              orgId,
              workflowId: step.install_workflow_id ?? run.id ?? '',
              stepId: step.id ?? '',
              approvalId: step.approval!.id!,
              groupName,
            }}
          />
        ),
      }
    })

  return <BranchRunApproval items={items} />
}
