import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { useRunner } from '@/hooks/use-runner'
import { UpdateRunnerButton } from './UpdateRunner'
import { ShutdownRunnerControl } from './ShutdownRunnerControl'
import { ShutdownInstanceButton } from './ShutdownInstance'
import { DeprovisionRunnerButton } from './DeprovisionRunner'
import { PruneRunnerTokensButton } from './PruneRunnerTokens'
import type { TRunnerSettings } from '@/types'

export const ManagementDropdown = ({
  isInstallRunner = false,
  settings,
}: {
  isInstallRunner?: boolean
  settings: TRunnerSettings
}) => {
  const { runner, isManaged } = useRunner()
  if (!runner) return null
  return (
    <Dropdown
      id={`runner-${runner.id}-mgmt`}
      buttonText={
        <>
          <Icon variant="SlidersHorizontalIcon" /> {isInstallRunner ? 'Manage install runner' : 'Manage build runner'}
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

        <ShutdownRunnerControl isMenuButton runnerId={runner.id} />

        {isInstallRunner && isManaged ? (
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
