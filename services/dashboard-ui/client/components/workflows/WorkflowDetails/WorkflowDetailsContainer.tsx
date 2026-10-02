import { useWorkflow } from '@/hooks/use-workflow'
import { WorkflowDetails } from './WorkflowDetails'

export const WorkflowDetailsContainer = ({
  showBanners = true,
  backLink = true,
}: {
  showBanners?: boolean
  backLink?: boolean
}) => {
  const { workflow, failedSteps } = useWorkflow()
  return (
    <WorkflowDetails
      workflow={workflow}
      failedSteps={failedSteps}
      showBanners={showBanners}
      backLink={backLink}
    />
  )
}
