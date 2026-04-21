import type { TVCSConnection, TVCSConnectionReposResponse } from '@/types'
import { GitHubAccountSection } from './GitHubAccountSection'
import { RepositoriesSection } from './RepositoriesSection'

interface IConnectionDetail {
  vcs_connection: TVCSConnection
  repos?: TVCSConnectionReposResponse
  reposError?: any
  isLoadingRepos?: boolean
}

export const ConnectionDetail = ({
  vcs_connection,
  repos,
  reposError,
  isLoadingRepos = false,
}: IConnectionDetail) => (
  <>
    <GitHubAccountSection vcs_connection={vcs_connection} />
    <RepositoriesSection repos={repos} error={reposError} isLoading={isLoadingRepos} />
  </>
)
