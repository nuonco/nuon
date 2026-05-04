import { api } from '@/lib/api'

export async function removeVCSConnection({
  orgId,
  connectionId,
  deleteGithubApp = false,
}: {
  orgId: string
  connectionId: string
  deleteGithubApp?: boolean
}) {
  const query = deleteGithubApp ? '?delete_github_app=true' : ''
  return api({
    method: 'DELETE',
    orgId,
    path: `vcs/connections/${connectionId}${query}`,
  })
}
