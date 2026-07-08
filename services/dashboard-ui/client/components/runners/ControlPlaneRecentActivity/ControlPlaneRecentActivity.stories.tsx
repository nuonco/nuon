export default {
  title: 'Runners/ControlPlaneRecentActivity',
}

import { RunnerRecentActivityComponent } from '../RunnerRecentActivity'

const mockJobs = [
  { id: 'job-1', type: 'build', status: 'succeeded', created_at: '2024-01-15T10:00:00Z', group: 'build' },
  { id: 'job-2', type: 'sandbox-build', status: 'running', created_at: '2024-01-15T09:00:00Z', group: 'build' },
] as any[]

export const Default = () => (
  <RunnerRecentActivityComponent jobs={mockJobs} isLoading={false} hasNext={false} offset={0} />
)

export const Loading = () => (
  <RunnerRecentActivityComponent jobs={[]} isLoading={true} hasNext={false} offset={0} />
)

export const Empty = () => (
  <RunnerRecentActivityComponent jobs={[]} isLoading={false} hasNext={false} offset={0} />
)
