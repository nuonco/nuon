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
  const { workflow, workflowSteps } = useWorkflow()

  const metadata = workflow?.status?.metadata
  const eagerStepsLoaded = !!metadata?.eager_steps_loaded
  const allStepsLoaded = !!metadata?.all_steps_loaded

  return (
    <WorkflowSteps
      approvalPrompt={approvalPrompt}
      planOnly={planOnly}
      workflowSteps={workflowSteps}
      eagerStepsLoaded={eagerStepsLoaded}
      allStepsLoaded={allStepsLoaded}
    />
  )
}
