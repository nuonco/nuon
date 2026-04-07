import { Dropdown, type IDropdown } from '@/components/common/Dropdown'
import { DeployMenuContainer as DeployMenu } from './DeployMenuContainer'

interface IDeploySwitcher
  extends Omit<IDropdown, 'children' | 'id' | 'buttonText'> {
  componentId: string
  deployId: string
}

export const DeploySwitcher = ({
  alignment = 'right',
  componentId,
  deployId,
  ...props
}: IDeploySwitcher) => {
  return (
    <Dropdown
      id="deploy-switcher"
      alignment={alignment}
      buttonText="Latest deploys"
      {...props}
    >
      <DeployMenu activeDeployId={deployId} componentId={componentId} />
    </Dropdown>
  )
}
