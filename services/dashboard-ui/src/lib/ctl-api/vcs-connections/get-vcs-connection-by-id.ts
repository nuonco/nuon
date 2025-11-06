import { api } from '@/lib/api'
import type { TVCSConnection } from '@/types'

export interface IGetVCSConnectionById {
  orgId: string
  connectionId: string
}

export async function getVCSConnectionById({
  orgId,
  connectionId,
}: IGetVCSConnectionById) {
  return api<TVCSConnection>({
    orgId,
    path: `vcs/connections/${connectionId}`,
  })
}
