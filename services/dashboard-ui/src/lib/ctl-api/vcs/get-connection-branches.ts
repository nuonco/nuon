import { api } from '@/lib/api'

export interface Branch {
  name: string
}

export const getConnectionBranches = (
  orgId: string,
  connectionId: string,
  owner: string,
  repo: string,
) =>
  api<Branch[]>({
    path: `vcs/connections/${connectionId}/branches?owner=${owner}&repo=${repo}`,
    orgId,
  })
