import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { ApproveAllButton } from '@/components/approvals/ApproveAll'
import type { TWorkflow } from '@/types'
import { CancelWorkflowButton } from '../../CancelWorkflow'
import { WorkflowChangesTrigger } from '../../WorkflowChangesSummary'

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
      {workflow?.id && (
        <AdminDashboardLink
          path={`/workflows/${workflow.id}`}
          label="Admin panel"
        />
      )}

      <WorkflowChangesTrigger workflow={workflow} />

      {canShowApproveAll && <ApproveAllButton workflow={workflow} />}

      {canShowCancel && <CancelWorkflowButton workflow={workflow} />}
    </div>
  )
}
