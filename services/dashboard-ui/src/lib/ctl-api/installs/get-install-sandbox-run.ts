import type { TSandboxRun } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IGetInstallSandboxRun extends IGetInstall {
  installSandboxRunId: string
}

export async function getInstallSandboxRun({
  installSandboxRunId,
  installId,
  orgId,
}: IGetInstallSandboxRun) {
  return queryData<TSandboxRun>({
    errorMessage: 'Unable to retrieve install sandbox run.',
    orgId,
    path: `installs/sandbox-runs/${installSandboxRunId}`,
    abortTimeout: 10000,
  })
}