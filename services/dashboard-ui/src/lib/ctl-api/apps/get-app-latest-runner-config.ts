import type { TAppRunnerConfig } from '@/types'
import { queryData } from '@/utils'
import type { IGetApp } from '../shared-interfaces'

export interface IGetAppLatestRunnerConfig extends IGetApp {}

export async function getAppLatestRunnerConfig({
  appId,
  orgId,
}: IGetAppLatestRunnerConfig) {
  return queryData<TAppRunnerConfig>({
    errorMessage: 'Unable to retrieve latest runner config',
    orgId,
    path: `apps/${appId}/runner-latest-config`,
  })
}