import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getRunnerLatestHeartbeat, getRunnerSettings } from '@/lib'
import type { TRunnerHeartbeat, TRunnerGroupSettings } from '@/types'
import { LoadRunnerHeartbeat } from './LoadRunnerHeartbeat'

interface LoadRunnerHeartbeatContainerProps {
  runnerId: string
}

export const LoadRunnerHeartbeatContainer = ({ runnerId }: LoadRunnerHeartbeatContainerProps) => {
  const { org } = useOrg()
  const orgId = org.id

  const { data, error: queryError, isLoading } = useQuery<{ build?: TRunnerHeartbeat; install?: TRunnerHeartbeat }>({
    queryKey: ['runner-heartbeat', orgId, runnerId],
    queryFn: () => getRunnerLatestHeartbeat({ orgId, runnerId }),
    enabled: !!orgId && !!runnerId,
  })

  const { data: settings, isLoading: isSettingsLoading } = useQuery<TRunnerGroupSettings>({
    queryKey: ['runner-settings', orgId, runnerId],
    queryFn: () => getRunnerSettings({ orgId, runnerId }),
    enabled: !!orgId && !!runnerId,
  })

  const heartbeat = data?.build || data?.install

  return (
    <LoadRunnerHeartbeat
      heartbeat={heartbeat}
      error={queryError ? 'Unable to load runner heartbeat' : null}
      isLoading={isLoading}
      platform={settings?.platform || settings?.metadata?.['runner.platform']}
      isPlatformLoading={isSettingsLoading}
    />
  )
}
