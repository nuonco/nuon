import type { TActionWorkflow } from '@/types'
import { queryData } from '@/utils'
import type { IGetApp } from '../shared-interfaces'

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