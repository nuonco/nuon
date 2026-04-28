import { useWorkflow } from '@/hooks/use-workflow'
import { useWorkflowActions } from '@/hooks/use-workflow-actions'
import { WorkflowActionButtons } from './WorkflowActionButtons'

export const WorkflowActionButtonsContainer = () => {
  const { workflow, hasApprovals } = useWorkflow()

  const {
    canShowApproveAll,
    canShowCancel,
  } = useWorkflowActions(workflow, hasApprovals)

  return (
    <WorkflowActionButtons
      workflow={workflow}
      canShowApproveAll={canShowApproveAll}
      canShowCancel={canShowCancel}
    />
  )
}
