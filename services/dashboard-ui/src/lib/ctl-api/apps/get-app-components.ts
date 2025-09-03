import type { TComponent } from '@/types'
import { queryData } from '@/utils'
import type { IGetApp } from '../shared-interfaces'

export interface IGetAppComponents extends IGetApp {}

export async function getAppComponents({ appId, orgId }: IGetAppComponents) {
  return queryData<Array<TComponent>>({
    errorMessage: 'Unable to retrieve app components',
    orgId,
    path: `apps/${appId}/components`,
  })
}