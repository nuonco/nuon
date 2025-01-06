import type {
  TActionWorkflow,
  TApp,
  TAppConfig,
  TAppInputConfig,
  TAppRunnerConfig,
  TAppSandboxConfig,
  TComponent,
  TInstall,
} from '@/types'
import { queryData } from '@/utils'

export interface IGetApps {
  orgId: string
}

export async function getApps({ orgId }: IGetApps) {
  return queryData<Array<TApp>>({
    errorMessage: 'Unable to retrieve your apps.',
    orgId,
    path: 'apps',
  })
}

export interface IGetApp extends IGetApps {
  appId: string
}

export async function getApp({ appId, orgId }: IGetApp) {
  return queryData<TApp>({
    errorMessage: 'Unable to retrieve app.',
    orgId,
    path: `apps/${appId}`,
  })
}

export interface IGetAppComponents extends IGetApp {}

export async function getAppComponents({ appId, orgId }: IGetAppComponents) {
  return queryData<Array<TComponent>>({
    errorMessage: 'Unable to retrieve app components',
    orgId,
    path: `apps/${appId}/components`,
  })
}

export interface IGetAppConfigs extends IGetApp {}

export async function getAppConfigs({ appId, orgId }: IGetAppConfigs) {
  return queryData<Array<TAppConfig>>({
    errorMessage: 'Unable to retrieve app configs',
    orgId,
    path: `apps/${appId}/configs`,
  })
}

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

export interface IGetAppInstalls extends IGetApp {}

export async function getAppInstalls({ appId, orgId }: IGetAppInstalls) {
  return queryData<Array<TInstall>>({
    errorMessage: 'Unable to retrieve app installs',
    orgId,
    path: `apps/${appId}/installs`,
  })
}

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

export interface IGetAppActionWorkflows extends IGetApp {}

export async function getAppActionWorkflows({
  appId,
  orgId,
}: IGetAppActionWorkflows) {
  return queryData<Array<TActionWorkflow>>({
    errorMessage: 'Unable to retrieve app action workflows',
    orgId,
    path: `apps/${appId}/action-workflows`,
  })
}

export interface IGetAppActionWorkflow extends IGetApps {
  actionWorkflowId: string
}

export async function getAppActionWorkflow({
  actionWorkflowId,
  orgId,
}: IGetAppActionWorkflow) {
  return queryData<TActionWorkflow>({
    errorMessage: 'Unable to retrieve action workflow',
    orgId,
    path: `action-workflows/${actionWorkflowId}`,
  })
}
