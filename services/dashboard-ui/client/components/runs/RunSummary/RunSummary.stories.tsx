export default {
  title: 'Runs/RunSummary',
}

import { Text } from '@/components/common/Text'
import type { TRunnerJob } from '@/types'
import { RunSummary } from './RunSummary'

const created = '2026-08-24T10:12:00Z'
const planned = '2026-08-24T10:13:40Z'
const applied = '2026-08-24T10:19:42Z'

const jobs = [
  {
    id: 'rj01hzk8t3fqp2r9x4m7wcn5vb',
    operation: 'create-plan',
    status: 'finished',
    status_v2: { status: 'success' },
    started_at: created,
    finished_at: planned,
    execution_count: 1,
  },
  {
    id: 'rj01hzk8t3fqp2r9x4m7wcn5vc',
    operation: 'apply-plan',
    status: 'finished',
    status_v2: { status: 'success' },
    started_at: planned,
    finished_at: applied,
    execution_count: 1,
  },
] as unknown as TRunnerJob[]

const failedJobs = [
  jobs[0],
  {
    ...jobs[1],
    status: 'failed',
    status_v2: { status: 'error' },
    status_description: 'AccessDenied: iam:PassRole is not allowed on the execution role.',
  },
] as unknown as TRunnerJob[]

const timings = [
  { label: 'Created', time: created },
  { label: 'Planned', time: planned },
  { label: 'Applied', time: applied },
]

export const Default = () => (
  <RunSummary
    status={{ status: 'success', status_human_description: 'Deploy applied.' }}
    timings={timings}
    duration={{ beginTime: created, endTime: applied }}
    jobs={jobs}
    jobHref={(job) => `/org-001/installs/inst-001/runner/jobs/${job.id}`}
    triggeredBy={<Text variant="subtext">jane@example.com</Text>}
  />
)

export const Failed = () => (
  <RunSummary
    status={{
      status: 'error',
      status_human_description: 'The apply step failed.',
    }}
    timings={timings}
    duration={{ beginTime: created, endTime: applied }}
    jobs={failedJobs}
    jobHref={(job) => `/org-001/installs/inst-001/runner/jobs/${job.id}`}
  />
)

export const Queued = () => (
  <RunSummary
    status={{ status: 'queued' }}
    timings={[
      { label: 'Created', time: created },
      { label: 'Planned' },
      { label: 'Applied' },
    ]}
    duration={{ beginTime: created }}
    jobs={[]}
  />
)

export const Loading = () => (
  <RunSummary
    isLoading
    timings={[{ label: 'Created' }, { label: 'Planned' }, { label: 'Applied' }]}
    duration={{}}
  />
)
