import type { TRunnerJob } from '@/types'
import { mutateData } from '@/utils'

export interface ICancelRunnerJob {
  runnerJobId: string
  orgId: string
}

export async function cancelRunnerJob({
  orgId,
  runnerJobId,
}: ICancelRunnerJob) {
  return mutateData<TRunnerJob>({
    data: {},
    errorMessage: 'Unable to cancel runner job.',
    orgId,
    path: `runner-jobs/${runnerJobId}/cancel`,
  })
}