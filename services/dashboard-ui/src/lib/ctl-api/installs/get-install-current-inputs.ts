import type { TInstallInputs } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IGetInstallCurrentInputs extends IGetInstall {}

export async function getInstallCurrentInputs({
  installId,
  orgId,
}: IGetInstallCurrentInputs) {
  return queryData<TInstallInputs>({
    errorMessage: 'Unable to retrieve current install inputs.',
    orgId,
    path: `installs/${installId}/inputs/current`,
  })
}