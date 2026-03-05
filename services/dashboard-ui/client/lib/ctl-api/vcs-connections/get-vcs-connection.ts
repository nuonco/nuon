import { api } from '@/lib/api'
import type { TVCSConnection } from '@/types'

export interface IGetVCSConnection {
  orgId: string
  connectionId: string
}

export async function getVCSConnection({
  orgId,
  connectionId,
}: IGetVCSConnection) {
  return api<TVCSConnection>({
    orgId,
    path: `vcs/connections/${connectionId}`,
  })
}
