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
}: IActiveWorkflowsContainer) => {
  return (
    <ActiveWorkflows
      workflows={workflows}
      install={install}
    />
  )
}
