'use client'

import { ComponentsGraph } from '@/components/apps/ConfigGraph'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { DeployAllComponentsButton } from '@/components/install-components/management/DeployAllComponents'
import { DriftScanAllComponentsButton } from '@/components/install-components/management/DriftScanAllComponents'
import { TeardownAllComponentsButton } from '@/components/install-components/management/TeardownAllComponents'
import { useInstall } from '@/hooks/use-install'

export const ManageAllDropdown = () => {
  const { install } = useInstall()
  return (
    <Dropdown
      id="install-components-mgmt"
      variant="ghost"
      buttonText={
        <>
          <Icon variant="SlidersHorizontalIcon" /> Manage
        </>
      }
      alignment="right"
    >
      <Menu>
        <ComponentsGraph
          appId={install?.app_id}
          configId={install?.app_config_id}
        />
        <DriftScanAllComponentsButton isMenuButton />
        <DeployAllComponentsButton isMenuButton />
        <hr />
        <TeardownAllComponentsButton isMenuButton />
      </Menu>
    </Dropdown>
  )
}
