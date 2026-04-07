export default {
  title: 'Runners/RunnerRecentActivity',
}

import { RunnerRecentActivity } from './RunnerRecentActivity'

const mockJobs = [
  { id: 'job-1', type: 'deploy', status: 'succeeded', created_at: '2024-01-15T10:00:00Z', group: 'deploy' },
  { id: 'job-2', type: 'build', status: 'running', created_at: '2024-01-15T09:00:00Z', group: 'build' },
] as any[]

export const Default = () => (
  <RunnerRecentActivity jobs={mockJobs} isLoading={false} hasNext={false} offset={0} />
)

export const Loading = () => (
  <RunnerRecentActivity jobs={[]} isLoading={true} hasNext={false} offset={0} />
)

export const Empty = () => (
  <RunnerRecentActivity jobs={[]} isLoading={false} hasNext={false} offset={0} />
)
