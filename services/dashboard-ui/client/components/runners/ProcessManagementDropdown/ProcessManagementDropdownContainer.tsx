import { useNavigate } from 'react-router'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import type { TRunnerProcess } from '@/types'
import { ProcessManagementDropdown } from './ProcessManagementDropdown'

export const ProcessManagementDropdownContainer = ({
  process,
}: {
  process: TRunnerProcess
}) => {
  const { org } = useOrg()
  const { runner } = useRunner()
  const navigate = useNavigate()

  if (!runner) return null

  return (
    <ProcessManagementDropdown
      process={process}
      runnerId={runner.id}
      onViewSystemLogs={
        process.log_stream_id
          ? () => navigate(`/${org.id}/runner/processes/${process.id}/logs`)
          : undefined
      }
    />
  )
}
