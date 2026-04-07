import { useWorkflow } from '@/hooks/use-workflow'
import { useQueryParams } from '@/hooks/use-query-params'
import { useWorkflowActions } from '@/hooks/use-workflow-actions'
import { WorkflowActionButtons } from './WorkflowActionButtons'

export const WorkflowActionButtonsContainer = () => {
  const { workflow, hasApprovals } = useWorkflow()

  const temporalLinkParams = useQueryParams({
    query: `\`WorkflowId\` STARTS_WITH "${workflow?.owner_id}-execute-workflow-${workflow?.id}"`,
  })

  const {
    canShowApproveAll,
    canShowCancel,
    canShowTemporalLink,
  } = useWorkflowActions(workflow, hasApprovals)

  return (
    <WorkflowActionButtons
      workflow={workflow}
      temporalLinkParams={temporalLinkParams}
      canShowApproveAll={canShowApproveAll}
      canShowCancel={canShowCancel}
      canShowTemporalLink={canShowTemporalLink}
    />
  )
}
