import type { TActionWorkflow } from '@/types'
import { queryData } from '@/utils'
import type { IGetApps } from '../shared-interfaces'

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