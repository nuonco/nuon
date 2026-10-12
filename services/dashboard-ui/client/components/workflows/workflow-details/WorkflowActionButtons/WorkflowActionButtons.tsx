import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { ApproveAllButton } from '@/components/approvals/ApproveAll'
import type { TWorkflow } from '@/types'
import { CancelWorkflowButton } from '../../CancelWorkflow'
import { WorkflowChangesTrigger } from '../../WorkflowChangesSummary'

export interface IWorkflowActionButtons {
  workflow: TWorkflow
  canShowApproveAll: boolean
  canShowCancel: boolean
  showChangesTrigger?: boolean
}

export const WorkflowActionButtons = ({
  workflow,
  canShowApproveAll,
  canShowCancel,
  showChangesTrigger = true,
}: IWorkflowActionButtons) => {
  return (
    <div className="flex items-center gap-4">
      {workflow?.id && (
        <AdminDashboardLink
          path={`/workflows/${workflow.id}`}
          label="Admin panel"
        />
      )}

      {showChangesTrigger ? (
        <WorkflowChangesTrigger workflow={workflow} />
      ) : null}

      {canShowApproveAll && <ApproveAllButton workflow={workflow} />}

      {canShowCancel && <CancelWorkflowButton workflow={workflow} />}
    </div>
  )
}
