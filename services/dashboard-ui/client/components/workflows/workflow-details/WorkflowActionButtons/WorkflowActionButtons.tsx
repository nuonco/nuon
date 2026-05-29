import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { ApproveAllButton } from '@/components/approvals/ApproveAll'
import { CreateManagedMonitorButton } from '@/components/datadog'
import type { TWorkflow } from '@/types'
import { CancelWorkflowButton } from '../../CancelWorkflow'

export interface IWorkflowActionButtons {
  workflow: TWorkflow
  canShowApproveAll: boolean
  canShowCancel: boolean
}

export const WorkflowActionButtons = ({
  workflow,
  canShowApproveAll,
  canShowCancel,
}: IWorkflowActionButtons) => {
  return (
    <div className="flex items-center gap-4">
      {canShowApproveAll && (
        <ApproveAllButton workflow={workflow} />
      )}

      {canShowCancel && (
        <CancelWorkflowButton workflow={workflow} />
      )}

      {workflow?.id && (
        <CreateManagedMonitorButton
          targetType="workflow"
          targetId={workflow.id}
          displayName={workflow.name || workflow.type}
          size="sm"
        />
      )}

      {workflow?.id && (
        <AdminDashboardLink
          path={`/workflows/${workflow.id}`}
          label="Admin panel"
        />
      )}
    </div>
  )
}
