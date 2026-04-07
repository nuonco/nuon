import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { UpdateRunnerButton } from '@/components/runners/management/UpdateRunner'
import { ShutdownRunnerControl } from '@/components/runners/management/ShutdownRunnerControl'
import type { TRunnerProcess, TRunnerSettings } from '@/types'

interface IProcessManagementDropdown {
  process: TRunnerProcess
  settings?: TRunnerSettings
  runnerId: string
  onViewSystemLogs?: () => void
}

export const ProcessManagementDropdown = ({
  process,
  settings,
  runnerId,
  onViewSystemLogs,
}: IProcessManagementDropdown) => {
  return (
    <Dropdown
      buttonClassName="!p-1"
      id={`process-${process.id}-mgmt`}
      buttonText={<Icon variant="DotsThreeVertical" />}
      alignment="right"
      variant="ghost"
      hideIcon
    >
      <Menu>
        <Text>Controls</Text>
        {process.composite_status?.status === 'active' && settings ? (
          <UpdateRunnerButton settings={settings} isMenuButton />
        ) : null}

        {process.composite_status?.status === 'active' ? (
          <ShutdownRunnerControl isMenuButton isManaged runnerId={runnerId} processId={process.id} />
        ) : null}

        {process.log_stream_id && onViewSystemLogs ? (
          <Button
            isMenuButton
            onClick={onViewSystemLogs}
          >
            View system logs
            <Icon variant="TerminalWindowIcon" />
          </Button>
        ) : null}
      </Menu>
    </Dropdown>
  )
}
