import { useDeploy } from '@/hooks/use-deploy'
import { useInstall } from '@/hooks/use-install'
import { DeployHeader } from './DeployHeader'
import type { TComponent, TWorkflow } from '@/types'

interface IDeployHeaderContainer {
  component: TComponent
  workflow: TWorkflow
  stepId: string
}

export const DeployHeaderContainer = ({ component, workflow, stepId }: IDeployHeaderContainer) => {
  const { deploy } = useDeploy()
  const { install } = useInstall()

  return (
    <DeployHeader
      component={component}
      workflow={workflow}
      stepId={stepId}
      deploy={deploy}
      install={install}
    />
  )
}
