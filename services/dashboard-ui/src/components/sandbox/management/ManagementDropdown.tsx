'use client'

import { Dropdown, type IDropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { DriftScanSandboxButton } from './DriftScanSandbox'
import { ReprovisionSandboxButton } from './ReprovisionSandbox'
import { DeprovisionSandboxButton } from './DeprovisionSandbox'

export const ManagementDropdown = ({
  alignment = 'right',
  ...props
}: Omit<IDropdown, 'id' | 'buttonText' | 'children'>) => {
  return (
    <Dropdown
      id="sandbox-mgmt"
      buttonText={
        <>
          <Icon variant="SlidersHorizontalIcon" /> Manage sandbox
        </>
      }
      alignment={alignment}
      {...props}
    >
      <Menu>
        <div className="px-2 pt-2 pb-1">
          <Text variant="subtext" theme="neutral">
            Controls
          </Text>
        </div>

        <DriftScanSandboxButton isMenuButton />
        <ReprovisionSandboxButton isMenuButton />

        <hr />
        <div className="px-2 pt-2 pb-1">
          <Text variant="subtext" theme="neutral">
            Remove
          </Text>
        </div>

        <DeprovisionSandboxButton isMenuButton />
      </Menu>
    </Dropdown>
  )
}
