import { api } from '@/lib/api'
import type { TVCSConnectionCommit } from '@/types'

export interface IGetVCSConnectionCommits {
  orgId: string
  connectionId: string
  limit?: number
  branch?: string
}

export async function getVCSConnectionCommits({
  orgId,
  connectionId,
  limit = 50,
  branch,
}: IGetVCSConnectionCommits) {
  const params = new URLSearchParams()
  params.set('limit', String(limit))
  if (branch) {
    params.set('branch', branch)
  }
  return api<TVCSConnectionCommit[]>({
    orgId,
    path: `vcs/connections/${connectionId}/commits?${params.toString()}`,
  })
}
