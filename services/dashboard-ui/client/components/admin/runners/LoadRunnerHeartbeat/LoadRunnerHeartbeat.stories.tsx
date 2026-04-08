export default {
  title: 'Admin/LoadRunnerHeartbeat',
}

import { LoadRunnerHeartbeat } from './LoadRunnerHeartbeat'

export const Default = () => (
  <LoadRunnerHeartbeat
    heartbeat={{ version: 'v1.2.3', alive_time: 3600000000000, created_at: new Date().toISOString() } as any}
    error={null}
    isLoading={false}
    platform="linux/amd64"
    isPlatformLoading={false}
  />
)

export const Loading = () => (
  <LoadRunnerHeartbeat
    heartbeat={undefined}
    error={null}
    isLoading={true}
    isPlatformLoading={false}
  />
)

export const NoHeartbeat = () => (
  <LoadRunnerHeartbeat
    heartbeat={undefined}
    error={null}
    isLoading={false}
    isPlatformLoading={false}
  />
)
