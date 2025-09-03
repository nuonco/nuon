import type { TInstallWorkflow } from '@/types'
import { queryData } from '@/utils'

export interface IGetInstallWorkflow {
  installWorkflowId: string
  orgId: string
}

export async function getInstallWorkflow({
  installWorkflowId,
  orgId,
}: IGetInstallWorkflow) {
  return queryData<TInstallWorkflow>({
    errorMessage: 'Unable to get install workflow.',
    path: `install-workflows/${installWorkflowId}`,
    orgId,
  })
}