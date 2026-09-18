export default {
  title: 'Runs/RunFailureBanner',
}

import type { TRunnerJob } from '@/types'
import { RunFailureBanner } from './RunFailureBanner'

const failedJobs = [
  {
    id: 'rj01hzk8t3fqp2r9x4m7wcn5vc',
    operation: 'apply-plan',
    status: 'failed',
    status_v2: { status: 'error' },
    status_description:
      'AccessDenied: iam:PassRole is not allowed on the execution role.',
  },
] as unknown as TRunnerJob[]

export const Default = () => (
  <RunFailureBanner
    status={{
      status: 'error',
      status_human_description: 'The apply step failed.',
    }}
    jobs={failedJobs}
  />
)

export const ReasonFromJob = () => (
  <RunFailureBanner status={{ status: 'error' }} jobs={failedJobs} />
)

export const Succeeded = () => (
  <RunFailureBanner
    status={{ status: 'success', status_human_description: 'Build finished.' }}
  />
)
