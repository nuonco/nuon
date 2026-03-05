import { api } from '@/lib/api'
import type { TVCSConnectionStatus } from '@/types'

export interface ICheckVCSConnectionStatus {
  orgId: string
  connectionId: string
}

export async function checkVCSConnectionStatus({
  orgId,
  connectionId,
}: ICheckVCSConnectionStatus) {
  return api<TVCSConnectionStatus>({
    orgId,
    path: `vcs/connections/${connectionId}/check-status`,
  })
}
