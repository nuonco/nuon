import type { TInstallActionWorkflow } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IGetInstallActionWorkflowLatestRuns extends IGetInstall {}

export async function getInstallActionWorkflowLatestRun({
  installId,
  orgId,
}: IGetInstallActionWorkflowLatestRuns) {
  return queryData<Array<TInstallActionWorkflow>>({
    errorMessage: 'Unable to retrieve latest install action workflow runs',
    orgId,
    path: `installs/${installId}/action-workflows/latest-runs`,
  })
}