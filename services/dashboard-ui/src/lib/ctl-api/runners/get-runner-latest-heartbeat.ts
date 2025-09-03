import type { TRunnerHeartbeat } from '@/types'
import { queryData } from '@/utils'
import type { IGetRunner } from '../shared-interfaces'

export interface IGetRunnerLatestHeartbeat extends IGetRunner {}

export async function getRunnerLatestHeartbeat({
  orgId,
  runnerId,
}: IGetRunnerLatestHeartbeat) {
  return queryData<TRunnerHeartbeat>({
    errorMessage: 'Unable to retrieve latest runner heartbeat.',
    orgId,
    path: `runners/${runnerId}/latest-heart-beat`,
  })
}