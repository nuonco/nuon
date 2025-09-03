import type { TAppConfig } from '@/types'
import { queryData } from '@/utils'
import type { IGetApp } from '../shared-interfaces'

export interface IGetAppConfigs extends IGetApp {}

export async function getAppConfigs({ appId, orgId }: IGetAppConfigs) {
  return queryData<Array<TAppConfig>>({
    errorMessage: 'Unable to retrieve app configs',
    orgId,
    path: `apps/${appId}/configs`,
  })
}