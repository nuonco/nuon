export default {
  title: 'Runners/RunnerDetailsCard',
}

import { RunnerDetailsCard, RunnerDetailsCardSkeleton } from './RunnerDetailsCard'
import type { TRunner, TRunnerGroup } from '@/types'

const mockRunner = {
  id: 'rnr_abc123',
  status: 'active',
} as TRunner

const mockRunnerGroup = {
  platform: 'EKS',
} as unknown as TRunnerGroup

const mockHeartbeat = {
  version: 'v1.2.3',
  created_at: new Date().toISOString(),
  started_at: new Date(Date.now() - 3600000).toISOString(),
}

export const Default = () => (
  <RunnerDetailsCard
    runner={mockRunner}
    runnerGroup={mockRunnerGroup}
    heartbeat={mockHeartbeat}
  />
)

export const Disconnected = () => (
  <RunnerDetailsCard
    runner={mockRunner}
    runnerGroup={mockRunnerGroup}
    heartbeat={{
      ...mockHeartbeat,
      created_at: new Date(Date.now() - 60000).toISOString(),
    }}
  />
)

export const NoHeartbeat = () => (
  <RunnerDetailsCard
    runner={mockRunner}
    runnerGroup={mockRunnerGroup}
    heartbeat={undefined}
  />
)

export const Loading = () => <RunnerDetailsCardSkeleton />
