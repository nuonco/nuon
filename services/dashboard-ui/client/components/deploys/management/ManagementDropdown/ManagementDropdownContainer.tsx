import { useDeploy } from '@/hooks/use-deploy'
import type { TComponent, TWorkflow } from '@/types'
import { ManagementDropdown } from './ManagementDropdown'

export const ManagementDropdownContainer = ({
  component,
  workflow,
}: {
  component: TComponent
  workflow: TWorkflow
}) => {
  const { deploy } = useDeploy()

  return (
    <ManagementDropdown
      component={component}
      workflow={workflow}
      deploy={deploy}
    />
  )
}
