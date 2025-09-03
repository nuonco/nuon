import type { TSandboxRun } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IGetInstallSandboxRuns extends IGetInstall {}

export async function getInstallSandboxRuns({
  installId,
  orgId,
}: IGetInstallSandboxRuns) {
  return queryData<Array<TSandboxRun>>({
    errorMessage: 'Unable to get install sandbox runs',
    orgId,
    path: `installs/${installId}/sandbox-runs`,
  })
}