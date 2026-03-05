import { api } from '@/lib/api'
import type { TVCSConnectionReposResponse } from '@/types'

export interface IGetVCSConnectionRepos {
  orgId: string
  connectionId: string
}

export async function getVCSConnectionRepos({
  orgId,
  connectionId,
}: IGetVCSConnectionRepos) {
  return api<TVCSConnectionReposResponse>({
    orgId,
    path: `vcs/connections/${connectionId}/repos`,
  })
}
