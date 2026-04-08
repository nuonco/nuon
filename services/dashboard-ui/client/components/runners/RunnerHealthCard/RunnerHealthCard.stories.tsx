export default {
  title: 'Runners/RunnerHealthCard',
}

import { RunnerHealthCard, RunnerHealthCardSkeleton, RunnerHealthEmptyCard } from './RunnerHealthCard'
import type { TRunnerHealthCheck } from '@/types'

const now = Date.now()

const mockHealthchecks: TRunnerHealthCheck[] = Array.from({ length: 30 }, (_, i) => ({
  id: `hc_${i}`,
  status_code: i === 5 ? 1 : i === 20 ? 900 : 0,
  minute_bucket: new Date(now - (30 - i) * 60000).toISOString(),
})) as TRunnerHealthCheck[]

export const Default = () => (
  <RunnerHealthCard healthchecks={mockHealthchecks} />
)

export const AllHealthy = () => (
  <RunnerHealthCard
    healthchecks={mockHealthchecks.map((hc) => ({ ...hc, status_code: 0 }))}
  />
)

export const Empty = () => <RunnerHealthEmptyCard />

export const Loading = () => <RunnerHealthCardSkeleton />
