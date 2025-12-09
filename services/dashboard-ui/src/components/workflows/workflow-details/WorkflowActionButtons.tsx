'use client'

import { ApproveAllButton } from '@/components/approvals/ApproveAll'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { useWorkflow } from '@/hooks/use-workflow'
import { useQueryParams } from '@/hooks/use-query-params'
import { CancelWorkflowButton } from '../CancelWorkflow'
import { useWorkflowActions } from '@/hooks/use-workflow-actions'

export const WorkflowActionButtons = () => {
  const { workflow, hasApprovals } = useWorkflow()
  
  const temporalLinkParams = useQueryParams({
    query: `\`WorkflowId\` STARTS_WITH "${workflow?.owner_id}-execute-workflow-${workflow?.id}"`,
  })
  
  const { 
    canShowApproveAll, 
    canShowCancel, 
    canShowTemporalLink 
  } = useWorkflowActions(workflow, hasApprovals)

  return (
    <div className="flex items-center gap-4">
      {canShowApproveAll && (
        <ApproveAllButton workflow={workflow} />
      )}
      
      {canShowCancel && (
        <CancelWorkflowButton workflow={workflow} />
      )}
      
      {canShowTemporalLink && (
        <Button
          href={`/admin/temporal/namespaces/installs/workflows${temporalLinkParams}`}
          target="_blank"
        >
          View in Temporal <Icon variant="ArrowSquareOutIcon" />
        </Button>
      )}
    </div>
  )
}