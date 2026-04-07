import { useWorkflow } from '@/hooks/use-workflow'
import { WorkflowDetails } from './WorkflowDetails'

export const WorkflowDetailsContainer = () => {
  const { failedSteps } = useWorkflow()
  return <WorkflowDetails failedSteps={failedSteps} />
}
