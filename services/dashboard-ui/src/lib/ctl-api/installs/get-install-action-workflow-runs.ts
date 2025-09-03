import type { TInstallActionWorkflowRun } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IGetInstallActionWorkflowRuns extends IGetInstall {}

export async function getInstallActionWorkflowRuns({
  installId,
  orgId,
}: IGetInstallActionWorkflowRuns) {
  return queryData<Array<TInstallActionWorkflowRun>>({
    errorMessage: 'Unable to retrieve install action workflow runs.',
    orgId,
    path: `installs/${installId}/action-workflows/runs`,
  })
}