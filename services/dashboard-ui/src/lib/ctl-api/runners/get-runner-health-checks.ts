import type { TRunnerHealthCheck } from '@/types'
import { queryData } from '@/utils'
import type { IGetRunner } from '../shared-interfaces'

export interface IGetRunnerHealthChecks extends IGetRunner {}

export async function getRunnerHealthChecks({
  orgId,
  runnerId,
}: IGetRunnerHealthChecks) {
  return queryData<Array<TRunnerHealthCheck>>({
    errorMessage: 'Unable to retrieve runner recent health checks.',
    orgId,
    path: `runners/${runnerId}/recent-health-checks`,
  })
}