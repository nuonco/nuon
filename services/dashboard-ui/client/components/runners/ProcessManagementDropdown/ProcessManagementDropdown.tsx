import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { ShutdownRunnerControl } from '@/components/runners/management/ShutdownRunnerControl'
import type { TRunnerProcess } from '@/types'

interface IProcessManagementDropdown {
  process: TRunnerProcess
  runnerId: string
  systemLogsHref?: string
}

export const ProcessManagementDropdown = ({
  process,
  runnerId,
  systemLogsHref,
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

        {process.log_stream_id && systemLogsHref ? (
          <Link href={systemLogsHref}>
            View system logs
            <Icon variant="TerminalWindowIcon" />
          </Link>
        ) : null}
      </Menu>
    </Dropdown>
  )
}
