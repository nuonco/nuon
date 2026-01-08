import { api } from '@/lib/api'

export interface Repo {
  full_name: string
  name: string
  owner: string
  private: boolean
}

export const getConnectionRepos = (orgId: string, connectionId: string) =>
  api<Repo[]>({
    path: `vcs/connections/${connectionId}/repos`,
    orgId,
  })
