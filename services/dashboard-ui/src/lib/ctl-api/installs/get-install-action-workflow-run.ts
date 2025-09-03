import type { TInstallActionWorkflowRun } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IGetInstallActionWorkflowRun extends IGetInstall {
  actionWorkflowRunId: string
}

export async function getInstallActionWorkflowRun({
  actionWorkflowRunId,
  installId,
  orgId,
}: IGetInstallActionWorkflowRun) {
  return queryData<TInstallActionWorkflowRun>({
    errorMessage: 'Unable to retrieve install action workflow run.',
    orgId,
    path: `installs/${installId}/action-workflows/runs/${actionWorkflowRunId}`,
  })
}