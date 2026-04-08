import { useOrg } from '@/hooks/use-org'
import { ActiveWorkflows } from './ActiveWorkflows'
import type { TInstall, TWorkflow } from '@/types'

interface IActiveWorkflowsContainer {
  workflows: TWorkflow[]
  install?: TInstall
  hasDivider?: boolean
}

export const ActiveWorkflowsContainer = ({
  workflows,
  install,
  hasDivider,
}: IActiveWorkflowsContainer) => {
  const { org } = useOrg()

  return (
    <ActiveWorkflows
      orgId={org.id}
      workflows={workflows}
      install={install}
      hasDivider={hasDivider}
    />
  )
}
