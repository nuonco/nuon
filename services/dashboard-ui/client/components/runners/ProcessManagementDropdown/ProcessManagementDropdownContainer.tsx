import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import type { TRunnerProcess } from '@/types'
import { ProcessManagementDropdown } from './ProcessManagementDropdown'

export const ProcessManagementDropdownContainer = ({
  process,
  runnerBasePath,
}: {
  process: TRunnerProcess
  runnerBasePath?: string
}) => {
  const { org } = useOrg()
  const { runner } = useRunner()

  if (!runner) return null

  const basePath = runnerBasePath ?? `/${org?.id}/runner`

  return (
    <ProcessManagementDropdown
      process={process}
      runnerId={runner.id}
      systemLogsHref={
        process.log_stream_id
          ? `${basePath}/processes/${process.id}/logs`
          : undefined
      }
    />
  )
}
