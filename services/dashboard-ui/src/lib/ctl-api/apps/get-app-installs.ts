import type { TInstall } from '@/types'
import { queryData } from '@/utils'
import type { IGetApp } from '../shared-interfaces'

export interface IGetAppInstalls extends IGetApp {}

export async function getAppInstalls({ appId, orgId }: IGetAppInstalls) {
  return queryData<Array<TInstall>>({
    errorMessage: 'Unable to retrieve app installs',
    orgId,
    path: `apps/${appId}/installs`,
  })
}