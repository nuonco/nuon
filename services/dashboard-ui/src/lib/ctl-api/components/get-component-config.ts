import type { TComponentConfig } from '@/types'
import { queryData } from '@/utils'
import type { IGetComponent } from '../shared-interfaces'

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