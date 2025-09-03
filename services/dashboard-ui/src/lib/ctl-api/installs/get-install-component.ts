import type { TInstallComponent } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IGetInstallComponent extends IGetInstall {
  componentId: string
}

export async function getInstallComponent({
  componentId,
  installId,
  orgId,
}: IGetInstallComponent) {
  return queryData<TInstallComponent>({
    errorMessage: 'Unable to retrieve the components for this install.',
    orgId,
    path: `installs/${installId}/components/${componentId}`,
  })
}