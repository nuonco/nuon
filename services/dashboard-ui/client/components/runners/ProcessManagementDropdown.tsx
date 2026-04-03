import { useNavigate } from 'react-router'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { UpdateRunnerButton } from '@/components/runners/management/UpdateRunner'
import { ShutdownRunnerControl } from '@/components/runners/management/ShutdownRunnerControl'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import type { TRunnerProcess, TRunnerSettings } from '@/types'

export const ProcessManagementDropdown = ({
  process,
  settings,
}: {
  process: TRunnerProcess
  settings?: TRunnerSettings
}) => {
  const { org } = useOrg()
  const { runner } = useRunner()
  const navigate = useNavigate()

  if (!runner) return null

  return (
    <Dropdown
      id={`process-${process.id}-mgmt`}
      buttonText={<Icon variant="DotsThreeVertical" />}
      alignment="right"
      variant="ghost"
      hideIcon
    >
      <Menu>
        <Text>Controls</Text>
        {process.status === 'active' && settings ? (
          <UpdateRunnerButton settings={settings} isMenuButton />
        ) : null}

        {process.status === 'active' ? (
          <ShutdownRunnerControl isMenuButton runnerId={runner.id} processId={process.id} />
        ) : null}

        {process.log_stream_id ? (
          <Button
            isMenuButton
            onClick={() => navigate(`/${org.id}/runner/processes/${process.id}/logs`)}
          >
            View system logs
            <Icon variant="TerminalWindowIcon" />
          </Button>
        ) : null}
      </Menu>
    </Dropdown>
  )
}
