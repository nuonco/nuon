import type { TAppSandboxConfig } from '@/types'
import { queryData } from '@/utils'
import type { IGetApp } from '../shared-interfaces'

export interface IGetAppLatestSandboxConfig extends IGetApp {}

export async function getAppLatestSandboxConfig({
  appId,
  orgId,
}: IGetAppLatestSandboxConfig) {
  return queryData<TAppSandboxConfig>({
    errorMessage: 'Unable to retrieve latest sandbox config',
    orgId,
    path: `apps/${appId}/sandbox-latest-config`,
  })
}