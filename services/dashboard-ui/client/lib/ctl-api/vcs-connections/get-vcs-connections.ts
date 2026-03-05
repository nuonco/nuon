import { api } from '@/lib/api'
import type { TVCSConnection } from '@/types'

export interface IGetVCSConnections {
  orgId: string
}

export async function getVCSConnections({ orgId }: IGetVCSConnections) {
  return api<TVCSConnection[]>({
    orgId,
    path: `vcs/connections`,
  })
}
