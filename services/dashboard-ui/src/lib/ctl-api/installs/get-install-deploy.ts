import type { TInstallDeploy } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IGetInstallDeploy extends IGetInstall {
  installDeployId: string
}

export async function getInstallDeploy({
  installDeployId,
  installId,
  orgId,
}: IGetInstallDeploy) {
  return queryData<TInstallDeploy>({
    errorMessage: 'Unable to retrieve install deployment.',
    orgId,
    path: `installs/${installId}/deploys/${installDeployId}`,
  })
}