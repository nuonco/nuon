import type { TAppInputConfig } from '@/types'
import { queryData } from '@/utils'
import type { IGetApp } from '../shared-interfaces'

export interface IGetAppLatestInputConfig extends IGetApp {}

export async function getAppLatestInputConfig({
  appId,
  orgId,
}: IGetAppLatestInputConfig) {
  return queryData<TAppInputConfig>({
    errorMessage: 'Unable to retrieve latest input config',
    orgId,
    path: `apps/${appId}/input-latest-config`,
  })
}