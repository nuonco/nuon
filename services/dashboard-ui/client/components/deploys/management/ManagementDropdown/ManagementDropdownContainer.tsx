import { useDeploy } from '@/hooks/use-deploy'
import type { TComponent, TWorkflow } from '@/types'
import { ManagementDropdown } from './ManagementDropdown'

export const ManagementDropdownContainer = ({
  component,
  currentBuildId,
  workflow,
}: {
  component: TComponent
  currentBuildId?: string
  workflow: TWorkflow
}) => {
  const { deploy } = useDeploy()

  return (
    <ManagementDropdown
      component={component}
      currentBuildId={currentBuildId}
      workflow={workflow}
      deploy={deploy}
    />
  )
}
