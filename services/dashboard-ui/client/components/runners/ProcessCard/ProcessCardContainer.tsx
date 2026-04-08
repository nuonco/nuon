import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import {
  getProcessLatestHeartbeat,
  getRunnerRecentHealthChecks,
} from '@/lib'
import type { TRunnerProcess, TRunnerSettings } from '@/types'
import { isLessThan15SecondsOld } from '@/utils/time-utils'
import { ProcessManagementDropdown } from '@/components/runners/ProcessManagementDropdown'
import { ProcessCard } from './ProcessCard'

export const ProcessCardContainer = ({
  process,
  settings,
  shouldPoll = false,
  pollInterval = 10000,
}: {
  process: TRunnerProcess
  settings?: TRunnerSettings
  shouldPoll?: boolean
  pollInterval?: number
}) => {
  const { org } = useOrg()
  const { runner } = useRunner()

  const { data: heartbeat } = useQuery({
    queryKey: ['process-heartbeat', org?.id, runner?.id, process.id],
    queryFn: () =>
      getProcessLatestHeartbeat({
        orgId: org.id,
        runnerId: runner.id,
        processId: process.id,
      }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!runner?.id && !!process.id,
  })

  const { data: healthchecks } = useQuery({
    queryKey: ['process-health-checks', org?.id, runner?.id, process.id],
    queryFn: () =>
      getRunnerRecentHealthChecks({
        orgId: org.id,
        runnerId: runner.id,
        processId: process.id,
      }),
    refetchInterval: shouldPoll ? 60000 : false,
    enabled: !!org?.id && !!runner?.id && !!process.id,
  })

  const isConnected = isLessThan15SecondsOld(heartbeat?.created_at)
  const configuredVersion = settings?.container_image_tag || '-'
  const reportedVersion = heartbeat?.version || process.version || '-'

  return (
    <ProcessCard
      process={process}
      settings={settings}
      runnerId={runner?.id}
      isConnected={isConnected}
      heartbeatCreatedAt={heartbeat?.created_at}
      configuredVersion={configuredVersion}
      reportedVersion={reportedVersion}
      healthchecks={healthchecks || []}
      managementDropdown={
        <ProcessManagementDropdown
          process={process}
          settings={settings}
        />
      }
    />
  )
}
