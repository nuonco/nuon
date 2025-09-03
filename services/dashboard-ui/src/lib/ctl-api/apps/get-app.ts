import type { TApp } from '@/types'
import { queryData } from '@/utils'
import type { IGetApp } from '../shared-interfaces'

export async function getApp({ appId, orgId }: IGetApp) {
  return queryData<TApp>({
    errorMessage: 'Unable to retrieve app.',
    orgId,
    path: `apps/${appId}`,
  })
}