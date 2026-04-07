import { useWorkflow } from '@/hooks/use-workflow'
import { WorkflowStatusSection } from './WorkflowStatusSection'

export const WorkflowStatusSectionContainer = () => {
  const { workflow } = useWorkflow()

  if (!workflow) return null

  return <WorkflowStatusSection workflow={workflow} />
}
