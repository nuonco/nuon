import { queryData } from '@/utils'

export interface IGetRunnerJobPlan {
  orgId: string
  runnerJobId: string
}

export async function getRunnerJobPlan({
  orgId,
  runnerJobId,
}: IGetRunnerJobPlan) {
  return queryData<any>({
    errorMessage: 'Unable to retrieve runner job plant',
    orgId,
    path: `runner-jobs/${runnerJobId}/plan`,
  })
}