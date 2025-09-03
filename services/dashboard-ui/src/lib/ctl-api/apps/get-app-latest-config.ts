import type { TAppConfig } from '@/types'
import { queryData } from '@/utils'
import type { IGetApp } from '../shared-interfaces'

export interface IGetAppLatestConfig extends IGetApp {}

export async function getAppLatestConfig({
  appId,
  orgId,
}: IGetAppLatestConfig) {
  return queryData<TAppConfig>({
    errorMessage: 'Unable to retrieve latest app config',
    orgId,
    path: `apps/${appId}/latest-config`,
  })
}