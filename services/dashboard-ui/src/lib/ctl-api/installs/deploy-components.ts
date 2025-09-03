import { mutateData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IDeployComponents extends IGetInstall {}

export async function deployComponents({
  installId,
  orgId,
}: IDeployComponents) {
  return mutateData({
    data: { error_behavior: 'string' },
    errorMessage: 'Unable to deploy components to install.',
    orgId,
    path: `installs/${installId}/components/deploy-all`,
  })
}