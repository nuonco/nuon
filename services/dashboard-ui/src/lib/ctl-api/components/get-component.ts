import type { TComponent } from '@/types'
import { queryData } from '@/utils'
import type { IGetComponent } from '../shared-interfaces'

export async function getComponent({ componentId, orgId }: IGetComponent) {
  return queryData<TComponent>({
    errorMessage: 'Unable to retrieve component.',
    orgId,
    path: `components/${componentId}`,
  })
}