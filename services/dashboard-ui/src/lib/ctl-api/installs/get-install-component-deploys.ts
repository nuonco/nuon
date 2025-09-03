import type { TInstallDeploy } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstallComponent } from './get-install-component'

export interface IGetInstallComponentDeploys extends IGetInstallComponent {}

export async function getInstallComponentDeploys({
  componentId,
  installId,
  orgId,
}: IGetInstallComponentDeploys) {
  return queryData<Array<TInstallDeploy>>({
    errorMessage: 'Unable to retrieve deployments for this install component.',
    orgId,
    path: `installs/${installId}/components/${componentId}/deploys`,
  })
}