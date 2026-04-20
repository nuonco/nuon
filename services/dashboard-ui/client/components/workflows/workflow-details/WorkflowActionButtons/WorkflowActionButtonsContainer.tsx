import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { useWorkflow } from '@/hooks/use-workflow'
import { useWorkflowActions } from '@/hooks/use-workflow-actions'
import { WorkflowActionButtons } from './WorkflowActionButtons'

export const WorkflowActionButtonsContainer = () => {
  const { workflow, hasApprovals } = useWorkflow()
  const { user, isLoading } = useAuth()
  const config = useConfig()

  const {
    canShowApproveAll,
    canShowCancel,
  } = useWorkflowActions(workflow, hasApprovals)

  const isAdminVisible = !isLoading && !!user?.email?.endsWith('@nuon.co')

  return (
    <WorkflowActionButtons
      workflow={workflow}
      canShowApproveAll={canShowApproveAll}
      canShowCancel={canShowCancel}
      isAdminVisible={isAdminVisible}
      adminDashboardUrl={config.adminDashboardUrl ?? ''}
    />
  )
}
