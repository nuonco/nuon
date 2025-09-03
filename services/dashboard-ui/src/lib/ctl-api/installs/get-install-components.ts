import type { TInstallComponent } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IGetInstallComponents extends IGetInstall {}

export async function getInstallComponents({
  installId,
  orgId,
}: IGetInstallComponents) {
  return queryData<Array<TInstallComponent>>({
    errorMessage: 'Unable to retrieve the components for this install.',
    orgId,
    path: `installs/${installId}/components`,
  })
}