import type { TInstallWorkflow } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IGetInstallWorkflows extends IGetInstall {}

export async function getInstallWorkflows({
  installId,
  orgId,
}: IGetInstallWorkflows) {
  return queryData<Array<TInstallWorkflow>>({
    errorMessage: 'Unable to get install workflows.',
    path: `installs/${installId}/workflows`,
    orgId,
  })
}