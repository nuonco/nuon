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

export const WithCompositeError = () => (
  <RunnerJobHeader
    job={{
      ...mockJob,
      status: 'not-attempted',
      composite_error: {
        type: 'runner.job_lifecycle_failure',
        severity: 'error',
        message: 'No active runner was available for this job',
        sections: [
          {
            heading: 'What happened',
            body: 'The job could not start because its assigned runner had no active process.',
          },
          {
            heading: 'How to fix',
            body: 'Check that the runner is online and healthy, then retry the operation.',
          },
        ],
      },
    }}
  />
)
