import { api } from '@/lib/api'
import type { TVCSBranch } from '@/types'

export const getConnectionBranches = (
  orgId: string,
  connectionId: string,
  owner: string,
  repo: string
) =>
  api<TVCSBranch[]>({
    path: `vcs/connections/${connectionId}/branches?owner=${owner}&repo=${repo}`,
    orgId,
  })
