import { useWorkflow } from '@/hooks/use-workflow'
import { WorkflowSteps, WorkflowStepsSkeleton } from './WorkflowSteps'

export { WorkflowStepsSkeleton }

interface IWorkflowStepsContainer {
  approvalPrompt?: boolean
  planOnly?: boolean
}

export const WorkflowStepsContainer = ({
  approvalPrompt = false,
  planOnly = false,
}: IWorkflowStepsContainer) => {
  const { workflowSteps } = useWorkflow()

  return (
    <WorkflowSteps
      approvalPrompt={approvalPrompt}
      planOnly={planOnly}
      workflowSteps={workflowSteps}
    />
  )
}
