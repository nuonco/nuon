'use client'

import { ComponentsGraph } from '@/components/apps/ConfigGraph'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { BuildAllComponentsButton } from './BuildAllComponents'
import { useApp } from '@/hooks/use-app'

export const ManagementDropdown = () => {
  const { app } = useApp()
  return (
    <Dropdown
      id="components-mgmt"
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
          appId={app?.id}
          configId={app?.app_configs?.[0]?.id}
        />
        <BuildAllComponentsButton isMenuButton />
      </Menu>
    </Dropdown>
  )
}