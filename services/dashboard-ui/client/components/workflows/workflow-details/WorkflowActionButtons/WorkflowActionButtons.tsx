import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { ApproveAllButton } from '@/components/approvals/ApproveAll'
import type { TWorkflow } from '@/types'
import { CancelWorkflowButton } from '../../CancelWorkflow'

export interface IWorkflowActionButtons {
  workflow: TWorkflow
  canShowApproveAll: boolean
  canShowCancel: boolean
  isAdminVisible: boolean
  adminDashboardUrl?: string
}

export const WorkflowActionButtons = ({
  workflow,
  canShowApproveAll,
  canShowCancel,
  isAdminVisible,
  adminDashboardUrl,
}: IWorkflowActionButtons) => {
  return (
    <div className="flex items-center gap-4">
      {canShowApproveAll && (
        <ApproveAllButton workflow={workflow} />
      )}

      {canShowCancel && (
        <CancelWorkflowButton workflow={workflow} />
      )}

      {isAdminVisible && adminDashboardUrl && workflow?.id && (
        <Button href={`${adminDashboardUrl}/workflows/${workflow.id}`} target="_blank">

          View in admin panel <Icon variant="ArrowSquareOutIcon" />
        </Button>
      )}
    </div>
  )
}
