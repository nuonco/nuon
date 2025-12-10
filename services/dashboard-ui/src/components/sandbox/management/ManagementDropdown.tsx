'use client'

import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { DriftScanSandboxButton } from './DriftScanSandbox'
import { ReprovisionSandboxButton } from './ReprovisionSandbox'
import { DeprovisionSandboxButton } from './DeprovisionSandbox'

export const ManagementDropdown = () => {
  return (
    <Dropdown
      id="sandbox-mgmt"
      buttonText={
        <>
          <Icon variant="SlidersHorizontalIcon" /> Manage sandbox
        </>
      }
      alignment="right"
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
