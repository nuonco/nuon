'use client'

import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { UpdateRunnerButton } from './UpdateRunner'
import { ShutdownRunnerButton } from './ShutdownRunner'
import { ShutdownMngRunnerButton } from './ShutdownMngRunner'
import { ShutdownInstanceButton } from './ShutdownInstance'
import { DeprovisionRunnerButton } from './DeprovisionRunner'
import type { TRunner, TRunnerSettings } from '@/types'

export const ManagementDropdown = ({
  isInstallRunner = false,
  isManagedRunner = false,
  runner,
  settings,
}: {
  isInstallRunner?: boolean
  isManagedRunner?: boolean
  runner: TRunner
  settings: TRunnerSettings
}) => {
  return (
    <Dropdown
      id={`runner-${runner.id}-mgmt`}
      buttonText={
        <>
          <Icon variant="SlidersHorizontalIcon" /> Manage runner
        </>
      }
      alignment="right"
      variant={!isInstallRunner ? 'primary' : 'secondary'}
    >
      <Menu>
        <Text>Controls</Text>
        {settings ? (
          <UpdateRunnerButton
            runnerId={runner.id}
            settings={settings}
            isMenuButton
          />
        ) : null}

        {isInstallRunner && isManagedRunner ? (
          <ShutdownMngRunnerButton runnerId={runner.id} isMenuButton />
        ) : (
          <ShutdownRunnerButton runnerId={runner.id} isMenuButton />
        )}

        {isInstallRunner && isManagedRunner ? (
          <ShutdownInstanceButton runnerId={runner.id} isMenuButton />
        ) : null}

        {isInstallRunner && <hr />}

        {isInstallRunner && <Text>Remove</Text>}
        {isInstallRunner && <DeprovisionRunnerButton isMenuButton />}
      </Menu>
    </Dropdown>
  )
}
