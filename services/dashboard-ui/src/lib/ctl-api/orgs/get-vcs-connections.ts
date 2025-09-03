import type { TVCSConnection } from '@/types'
import { queryData } from '@/utils'
import type { IGetOrg } from '../shared-interfaces'

export interface IGetVCSConnections extends IGetOrg {}

export async function getVCSConnections({ orgId }: IGetVCSConnections) {
  return queryData<Array<TVCSConnection>>({
    errorMessage: 'Unable to retrieve connected version control systems',
    orgId,
    path: `vcs/connections`,
  })
}