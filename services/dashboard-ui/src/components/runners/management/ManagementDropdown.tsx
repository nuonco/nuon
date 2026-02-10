'use client'

import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { useRunner } from '@/hooks/use-runner'
import { UpdateRunnerButton } from './UpdateRunner'
import { ShutdownRunnerButton } from './ShutdownRunner'
import { ShutdownMngRunnerButton } from './ShutdownMngRunner'
import { ShutdownInstanceButton } from './ShutdownInstance'
import { DeprovisionRunnerButton } from './DeprovisionRunner'
import { PruneRunnerTokensButton } from './PruneRunnerTokens'
import type { TRunnerSettings } from '@/types'

export const ManagementDropdown = ({
  isInstallRunner = false,
  isManagedRunner = false,
  settings,
}: {
  isInstallRunner?: boolean
  isManagedRunner?: boolean
  settings: TRunnerSettings
}) => {
  const { runner } = useRunner()
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
          <UpdateRunnerButton settings={settings} isMenuButton />
        ) : null}

        {isInstallRunner && isManagedRunner ? (
          <ShutdownMngRunnerButton isMenuButton />
        ) : (
          <ShutdownRunnerButton isMenuButton />
        )}

        {isInstallRunner && isManagedRunner ? (
          <ShutdownInstanceButton isMenuButton />
        ) : null}

        {isInstallRunner ? <PruneRunnerTokensButton isMenuButton /> : null}

        {isInstallRunner && <hr />}

        {isInstallRunner && <Text>Remove</Text>}
        {isInstallRunner && <DeprovisionRunnerButton isMenuButton />}
      </Menu>
    </Dropdown>
  )
}
