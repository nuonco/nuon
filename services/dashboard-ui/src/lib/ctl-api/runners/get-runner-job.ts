import type { TRunnerJob } from '@/types'
import { queryData } from '@/utils'
import type { IGetRunner } from '../shared-interfaces'

export interface IGetRunnerJob extends Omit<IGetRunner, 'runnerId'> {
  runnerJobId: string
}

export async function getRunnerJob({ orgId, runnerJobId }: IGetRunnerJob) {
  return queryData<TRunnerJob>({
    errorMessage: 'Unable to retrieve runner job.',
    orgId,
    path: `runner-jobs/${runnerJobId}`,
  })
}