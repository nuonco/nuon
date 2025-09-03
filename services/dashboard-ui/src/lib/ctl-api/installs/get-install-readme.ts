import type { TReadme } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IGetInstallReadme extends IGetInstall {}

export async function getInstallReadme({
  orgId,
  installId,
}: IGetInstallReadme) {
  return queryData<TReadme>({
    errorMessage: 'Unable to retrieve the install README.',
    orgId,
    path: `installs/${installId}/readme`,
    abortTimeout: 100000,
  })
}