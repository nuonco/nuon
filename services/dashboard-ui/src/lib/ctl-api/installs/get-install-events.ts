import type { TInstallEvent } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IGetInstallEvents extends IGetInstall {}

export async function getInstallEvents({
  installId,
  orgId,
}: IGetInstallEvents) {
  return queryData<Array<TInstallEvent>>({
    errorMessage: 'Unable to retrieve install events.',
    orgId,
    path: `installs/${installId}/events`,
  })
}