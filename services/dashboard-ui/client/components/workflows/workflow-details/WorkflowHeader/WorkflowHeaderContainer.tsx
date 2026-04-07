import { useInstall } from '@/hooks/use-install'
import { useWorkflow } from '@/hooks/use-workflow'
import { WorkflowHeader } from './WorkflowHeader'

export const WorkflowHeaderContainer = () => {
  const { install } = useInstall()
  const { workflow } = useWorkflow()

  if (!workflow) return null

  return <WorkflowHeader workflow={workflow} install={install} />
}
