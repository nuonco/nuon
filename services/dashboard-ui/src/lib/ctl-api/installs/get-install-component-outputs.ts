import type { TInstallComponentOutputs } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IGetInstallComponentOutputs extends IGetInstall {
  componentId: string
}

export async function getInstallComponentOutputs({
  componentId,
  installId,
  orgId,
}: IGetInstallComponentOutputs) {
  return queryData<TInstallComponentOutputs>({
    errorMessage: 'Unable to retrieve install component outputs.',
    orgId,
    path: `installs/${installId}/components/${componentId}/outputs`,
  })
}