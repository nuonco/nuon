import type { TRunner } from '@/types'
import { queryData } from '@/utils'
import type { IGetRunner } from '../shared-interfaces'

export async function getRunner({ orgId, runnerId }: IGetRunner) {
  return queryData<TRunner>({
    errorMessage: 'Unable to retrieve runner.',
    orgId,
    path: `runners/${runnerId}`,
  })
}