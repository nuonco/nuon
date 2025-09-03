import type { TLogStream } from '@/types'
import { queryData } from '@/utils'
import type { IGetRunner } from '../shared-interfaces'

export interface IGetLogStream extends Omit<IGetRunner, 'runnerId'> {
  logStreamId: string
}

export async function getLogStream({ logStreamId, orgId }: IGetLogStream) {
  return queryData<TLogStream>({
    errorMessage: 'Unable to retrieve log stream.',
    orgId,
    path: `log-streams/${logStreamId}`,
  })
}