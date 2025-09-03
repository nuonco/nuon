import type { TInstallActionWorkflow } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IGetInstallActionWorkflowRecentRuns extends IGetInstall {
  actionWorkflowId: string
  offset?: string
  limit?: string
}

export async function getInstallActionWorkflowRecentRun({
  actionWorkflowId,
  installId,
  orgId,
}: IGetInstallActionWorkflowRecentRuns) {
  return queryData<TInstallActionWorkflow>({
    errorMessage: 'Unable to retrieve install action workflow runs',
    orgId,
    path: `installs/${installId}/action-workflows/${actionWorkflowId}/recent-runs`,
  })
}