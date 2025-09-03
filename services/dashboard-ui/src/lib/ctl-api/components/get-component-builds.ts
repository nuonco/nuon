import type { TBuild } from '@/types'
import { queryData } from '@/utils'
import type { IGetComponent } from '../shared-interfaces'

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