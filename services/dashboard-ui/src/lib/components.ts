import type { TBuild, TComponent, TComponentConfig } from '@/types'
import { mutateData, queryData } from '@/utils'

export interface IGetComponent {
  componentId: string
  orgId: string
}

export async function getComponent({ componentId, orgId }: IGetComponent) {
  return queryData<TComponent>({
    errorMessage: 'Unable to retrieve component.',
    orgId,
    path: `components/${componentId}`,
  })
}

export interface IGetComponentConfig extends IGetComponent {
  componentConfigId?: string
}

export async function getComponentConfig({
  componentId,
  componentConfigId,
  orgId,
}: IGetComponentConfig) {
  const configs = await queryData<Array<TComponentConfig>>({
    errorMessage: 'Unable to retrieve component config.',
    orgId,
    path: `components/${componentId}/configs`,
  })
  return componentConfigId
    ? configs.find((cfg) => cfg.id === componentConfigId)
    : configs[0]
}

export interface IGetComponentBuilds extends IGetComponent {}

export async function getComponentBuilds({
  componentId,
  orgId,
}: IGetComponent) {
  return queryData<Array<TBuild>>({
    errorMessage: 'Unable to retrieve component builds.',
    orgId,
    path: `builds?component_id=${componentId}`,
  })
}

export interface IGetComponentBuild extends Omit<IGetComponent, 'componentId'> {
  buildId: string
}

export async function getComponentBuild({
  buildId,
  orgId,
}: IGetComponentBuild) {
  return queryData<TBuild>({
    errorMessage: 'Unable to retrieve component build.',
    orgId,
    path: `components/builds/${buildId}`,
  })
}

export interface IGetLatestComponentBuild extends IGetComponent {}

export async function getLatestComponentBuild({
  componentId,
  orgId,
}: IGetLatestComponentBuild) {
  return queryData<TBuild>({
    errorMessage: 'Unable to retrieve component build.',
    orgId,
    path: `components/${componentId}/builds/latest`,
  })
}

interface ICreateComponentBuild {
  componentId: string
  orgId: string
}

export async function createComponentBuild({
  componentId,
  orgId,
}: ICreateComponentBuild) {
  return mutateData<TBuild>({
    data: { use_latest: true },
    errorMessage: 'Unable to kick off component build',
    orgId,
    path: `components/${componentId}/builds`,
  })
}
