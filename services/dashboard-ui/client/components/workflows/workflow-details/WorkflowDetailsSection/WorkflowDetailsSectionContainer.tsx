import { useOrg } from '@/hooks/use-org'
import { useWorkflow } from '@/hooks/use-workflow'
import { useInstall } from '@/hooks/use-install'
import { WorkflowDetailsSection } from './WorkflowDetailsSection'

export const WorkflowDetailsSectionContainer = () => {
  const { workflow } = useWorkflow()
  const { org } = useOrg()
  const { install } = useInstall()

  if (!workflow) return null

  return (
    <WorkflowDetailsSection
      workflow={workflow}
      orgId={org.id}
      install={install}
    />
  )
}
