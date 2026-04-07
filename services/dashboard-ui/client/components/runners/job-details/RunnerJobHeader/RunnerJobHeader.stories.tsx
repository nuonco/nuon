export default {
  title: 'Runners/RunnerJobHeader',
}

import { RunnerJobHeader } from './RunnerJobHeader'

const mockJob = {
  id: 'job-1',
  status: 'in-progress',
  type: 'deploy',
  group: 'default',
  created_at: new Date().toISOString(),
} as any

export const Default = () => (
  <RunnerJobHeader job={mockJob} />
)
