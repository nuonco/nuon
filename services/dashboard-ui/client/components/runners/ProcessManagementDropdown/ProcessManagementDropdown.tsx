import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { ShutdownRunnerControl } from '@/components/runners/management/ShutdownRunnerControl'
import type { TRunnerProcess } from '@/types'

interface IProcessManagementDropdown {
  process: TRunnerProcess
  runnerId: string
  onViewSystemLogs?: () => void
}

export const ProcessManagementDropdown = ({
  process,
  runnerId,
  onViewSystemLogs,
}: IProcessManagementDropdown) => {
  return (
    <Dropdown
      id={`process-${process.id}-mgmt`}
      icon={<Icon variant="DotsThreeVerticalIcon" />}
      alignment="right"
      buttonText=""
      buttonClassName="!p-1"
      variant="ghost"
    >
      <Menu>
        <Text>Controls</Text>
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
