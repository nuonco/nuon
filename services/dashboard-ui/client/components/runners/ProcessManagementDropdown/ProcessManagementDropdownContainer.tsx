import { useNavigate } from 'react-router'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import type { TRunnerProcess, TRunnerSettings } from '@/types'
import { ProcessManagementDropdown } from './ProcessManagementDropdown'

export const ProcessManagementDropdownContainer = ({
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
    <ProcessManagementDropdown
      process={process}
      settings={settings}
      runnerId={runner.id}
      onViewSystemLogs={
        process.log_stream_id
          ? () => navigate(`/${org.id}/runner/processes/${process.id}/logs`)
          : undefined
      }
    />
  )
}
