import { useDeepLinkedPanels } from '@/hooks/use-deep-linked-panels'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TInstallWorkflow } from '@/types'
import { GroupActionButton } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep/GroupApprovalActions'
import { WarmPlanInstallChanges } from '@/components/branches/PlanInstallChanges/WarmPlanInstallChanges'
import {
  BranchRunApproval,
  type IBranchRunApprovalItem,
} from './BranchRunApproval'
import { PlanDiffPanel } from './PlanDiffPanel'
import { getGroupName } from './use-group-plan-href'

interface IBranchPendingApprovalsContainer {
  run?: TInstallWorkflow
  className?: string
}

export const BranchPendingApprovalsContainer = ({
  run,
  className,
}: IBranchPendingApprovalsContainer) => {
  const { org } = useOrg()
  const { addPanel } = useSurfaces()
  const orgId = org?.id ?? ''
  const cancelled = !run || run.status?.status === 'cancelled'
  const approvalSteps = cancelled
    ? []
    : (run.steps ?? []).filter(
        (step) =>
          step.execution_type === 'approval' &&
          step.status?.status === 'approval-awaiting' &&
          !step.approval?.response &&
          !!step.approval?.id
      )

  useDeepLinkedPanels(
    approvalSteps.map((step) => ({
      key: `plan-diff-${step.id}`,
      build: () => <PlanDiffPanel step={step} />,
    }))
  )

  if (!run || cancelled) return null

  const items: IBranchRunApprovalItem[] = approvalSteps.map((step) => {
    const groupName = getGroupName(step.name)

    return {
      key: step.id ?? step.approval!.id!,
      groupName,
      onReview: () => {
        addPanel(<PlanDiffPanel step={step} />, `plan-diff-${step.id}`)
      },
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

  return (
    <>
      <WarmPlanInstallChanges steps={run.steps} />
      <BranchRunApproval items={items} className={className} />
    </>
  )
}
