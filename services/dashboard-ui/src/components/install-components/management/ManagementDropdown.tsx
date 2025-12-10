'use client'

import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { DeployComponentButton } from '@/components/install-components/management/DeployComponent'
import { DriftScanComponentButton } from '@/components/install-components/management/DriftScanComponent'
import { TeardownComponentButton } from '@/components/install-components/management/TeardownComponent'
import type { TComponent } from '@/types'

export const ManagementDropdown = ({
  component,
  currentBuildId,
}: {
  component: TComponent
  currentBuildId?: string
}) => {
  return (
    <Dropdown
      id={`component-${component.id}-mgmt`}
      variant="primary"
      buttonText={
        <>
          <Icon variant="SlidersHorizontalIcon" /> Manage
        </>
      }
      alignment="right"
    >
      <Menu>
        <Text>Controls</Text>
        <DriftScanComponentButton
          component={component}
          currentBuildId={currentBuildId}
          isMenuButton
        />
        <DeployComponentButton
          component={component}
          currentBuildId={currentBuildId}
          isMenuButton
        />
        <hr />
        <Text>Remove</Text>
        <TeardownComponentButton component={component} isMenuButton />
      </Menu>
    </Dropdown>
  )
}
