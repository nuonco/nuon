import { mutateData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IDeployComponentBuild extends IGetInstall {
  buildId: string
}

export async function deployComponentBuild({
  buildId,
  installId,
  orgId,
}: IDeployComponentBuild) {
  return mutateData({
    errorMessage: 'Unable to deploy component to install.',
    data: { build_id: buildId },
    orgId,
    path: `installs/${installId}/deploys`,
  })
}