import type { TRunnerGroup } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IGetInstallRunnerGroup extends IGetInstall {}

export async function getInstallRunnerGroup({
  installId,
  orgId,
}: IGetInstallRunnerGroup) {
  return queryData<TRunnerGroup>({
    errorMessage: 'Unable to retrieve install runner group.',
    orgId,
    path: `installs/${installId}/runner-group`,
  })
}