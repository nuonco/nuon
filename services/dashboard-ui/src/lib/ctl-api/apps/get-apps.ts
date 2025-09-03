import type { TApp } from '@/types'
import { queryData } from '@/utils'
import type { IGetApps } from '../shared-interfaces'

export async function getApps({ orgId }: IGetApps) {
  return queryData<Array<TApp>>({
    errorMessage: 'Unable to retrieve your apps.',
    orgId,
    path: 'apps',
  })
}