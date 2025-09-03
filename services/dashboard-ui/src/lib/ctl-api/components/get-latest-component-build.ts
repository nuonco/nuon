import type { TBuild } from '@/types'
import { queryData } from '@/utils'
import type { IGetComponent } from '../shared-interfaces'

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