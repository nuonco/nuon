import { ApproveAllButton } from '@/components/approvals/ApproveAll'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { TWorkflow } from '@/types'
import { CancelWorkflowButton } from '../../CancelWorkflow'

export interface IWorkflowActionButtons {
  workflow: TWorkflow
  temporalLinkParams: string
  canShowApproveAll: boolean
  canShowCancel: boolean
  canShowTemporalLink: boolean
}

export const WorkflowActionButtons = ({
  workflow,
  temporalLinkParams,
  canShowApproveAll,
  canShowCancel,
  canShowTemporalLink,
}: IWorkflowActionButtons) => {
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
