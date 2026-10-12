import { useInstall } from '@/hooks/use-install'
import { useWorkflow } from '@/hooks/use-workflow'
import { WorkflowHeader } from './WorkflowHeader'

export const WorkflowHeaderContainer = ({
  backLink = true,
}: {
  backLink?: boolean
}) => {
  const { install } = useInstall()
  const { workflow } = useWorkflow()

  if (!workflow) return null

  return (
    <WorkflowHeader workflow={workflow} install={install} backLink={backLink} />
  )
}
