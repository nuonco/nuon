import { useActiveWorkflows } from '@/hooks/use-active-workflows'
import { useOrg } from '@/hooks/use-org'
import { useWorkflowApprovals } from '@/hooks/use-workflow-approvals'
import { PendingApprovals } from './PendingApprovals'

export const PendingApprovalsContainer = () => {
  const { org } = useOrg()
  const { approvals } = useWorkflowApprovals()
  const { activeWorkflows } = useActiveWorkflows()

  return (
    <PendingApprovals
      orgId={org?.id}
      approvals={approvals}
      activeWorkflows={activeWorkflows}
    />
  )
}
