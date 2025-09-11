import type { TVCSConnection } from '@/types'
import { queryData } from '@/utils'

export interface IGetVCSConnections {
  orgId: string
}

export async function getVCSConnections({ orgId }: IGetVCSConnections) {
  return queryData<Array<TVCSConnection>>({
    errorMessage: 'Unable to retrieve connected version control systems',
    orgId,
    path: `vcs/connections`,
  })
}
