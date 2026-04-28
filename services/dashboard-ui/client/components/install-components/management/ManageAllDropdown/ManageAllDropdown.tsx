import { ComponentsGraph } from '@/components/apps/ConfigGraph'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { DeployAllComponentsButton } from '@/components/install-components/management/DeployAllComponents'
import { DriftScanAllComponentsButton } from '@/components/install-components/management/DriftScanAllComponents'
import { TeardownAllComponentsButton } from '@/components/install-components/management/TeardownAllComponents'

interface IManageAllDropdown {
  appId: string
  appConfigId: string
}

export const ManageAllDropdown = ({ appId, appConfigId }: IManageAllDropdown) => {
  return (
    <Dropdown
      id="install-components-mgmt"
      variant="secondary"
      buttonText={
        <>
          Component controls
        </>
      }
      alignment="right"
    >
      <Menu>
        <ComponentsGraph appId={appId} configId={appConfigId} />
        <DriftScanAllComponentsButton isMenuButton />
        <DeployAllComponentsButton isMenuButton />
        <hr />
        <TeardownAllComponentsButton isMenuButton />
      </Menu>
    </Dropdown>
  )
}
