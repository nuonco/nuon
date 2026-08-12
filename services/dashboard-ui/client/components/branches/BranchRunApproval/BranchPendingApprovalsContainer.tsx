import { useNavigate } from 'react-router'
import { useOrg } from '@/hooks/use-org'
import type { TInstallWorkflow } from '@/types'
import { GroupActionButton } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep/GroupApprovalActions'
import { BranchRunApproval, type IBranchRunApprovalItem } from './BranchRunApproval'

interface IBranchPendingApprovalsContainer {
  run?: TInstallWorkflow
  runHref?: string
  className?: string
}

const getGroupName = (name?: string) =>
  name?.replace(/^plan install group:\s*/i, '').trim() || 'install group'

export const BranchPendingApprovalsContainer = ({
  run,
  runHref,
  className,
}: IBranchPendingApprovalsContainer) => {
  const { org } = useOrg()
  const navigate = useNavigate()
  const orgId = org?.id ?? ''

  if (!run || !runHref) return null

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
      const params = new URLSearchParams({ workflow: run.id ?? '' })
      if (step.id) params.set('step', step.id)

      return {
        key: step.id ?? step.approval!.id!,
        groupName,
        onReview: () => navigate(`${runHref}?${params.toString()}`),
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
