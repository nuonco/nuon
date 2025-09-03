import type { TBuild } from '@/types'
import { queryData } from '@/utils'
import type { IGetComponent } from '../shared-interfaces'

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