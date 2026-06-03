import { ComponentsGraph } from '@/components/apps/ConfigGraph'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { BuildAllComponentsButton } from '../BuildAllComponents'

interface IManagementDropdown {
  appId: string
  appConfigId: string
}

export const ManagementDropdown = ({ appId, appConfigId }: IManagementDropdown) => {
  return (
    <Dropdown
      id="components-mgmt"
      buttonText={
        <>
          <Icon variant="SlidersHorizontalIcon" /> Manage
        </>
      }
      alignment="right"
    >
      <Menu>
        <ComponentsGraph appId={appId} configId={appConfigId} />
        <BuildAllComponentsButton isMenuButton />
      </Menu>
    </Dropdown>
  )
}
