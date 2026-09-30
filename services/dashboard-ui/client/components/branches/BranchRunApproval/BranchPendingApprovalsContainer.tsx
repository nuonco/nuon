import { useNavigate } from 'react-router'
import { useOrg } from '@/hooks/use-org'
import type { TInstallWorkflow } from '@/types'
import { GroupActionButton } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep/GroupApprovalActions'
import {
  BranchRunApproval,
  type IBranchRunApprovalItem,
} from './BranchRunApproval'
import { getGroupName, useGroupPlanHref } from './use-group-plan-href'

interface IBranchPendingApprovalsContainer {
  run?: TInstallWorkflow
  className?: string
}

export const BranchPendingApprovalsContainer = ({
  run,
  className,
}: IBranchPendingApprovalsContainer) => {
  const { org } = useOrg()
  const navigate = useNavigate()
  const groupPlanHref = useGroupPlanHref()
  const orgId = org?.id ?? ''

  if (!run) return null

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
        onReview: () => navigate(groupPlanHref(step.name)),
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

  return <BranchRunApproval items={items} className={className} />
}
