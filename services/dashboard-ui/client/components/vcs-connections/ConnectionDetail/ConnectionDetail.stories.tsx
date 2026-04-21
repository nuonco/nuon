export default {
  title: 'VCS Connections/ConnectionDetail',
}

import { ConnectionDetail } from './ConnectionDetail'
import type { TVCSConnection, TVCSConnectionReposResponse } from '@/types'

const mockConnection: TVCSConnection = {
  id: 'vcs-conn-abc123',
  github_account_name: 'nuonco',
  github_account_id: '12345678',
  github_install_id: '87654321',
  created_at: new Date(Date.now() - 86400000 * 30).toISOString(),
  updated_at: new Date(Date.now() - 3600000).toISOString(),
  vcs_connection_commit: [],
} as TVCSConnection

const mockRepos: TVCSConnectionReposResponse = {
  total_count: 3,
  repositories: [
    {
      id: 1,
      full_name: 'nuonco/example-app',
      html_url: 'https://github.com/nuonco/example-app',
      description: 'Example application configurations',
      private: false,
      fork: false,
      default_branch: 'main',
      updated_at: new Date(Date.now() - 86400000).toISOString(),
    },
    {
      id: 2,
      full_name: 'nuonco/private-app',
      html_url: 'https://github.com/nuonco/private-app',
      description: null,
      private: true,
      fork: false,
      default_branch: 'main',
      updated_at: new Date(Date.now() - 3600000).toISOString(),
    },
    {
      id: 3,
      full_name: 'nuonco/forked-app',
      html_url: 'https://github.com/nuonco/forked-app',
      description: 'A forked repository',
      private: false,
      fork: true,
      default_branch: 'develop',
      updated_at: new Date(Date.now() - 7200000).toISOString(),
    },
  ],
} as TVCSConnectionReposResponse

export const Default = () => (
  <ConnectionDetail vcs_connection={mockConnection} repos={mockRepos} />
)

export const LoadingRepos = () => (
  <ConnectionDetail vcs_connection={mockConnection} isLoadingRepos />
)

export const ReposError = () => (
  <ConnectionDetail
    vcs_connection={mockConnection}
    reposError={{ error: 'Failed to load repositories. Check your GitHub connection.' }}
  />
)

export const NoRepos = () => (
  <ConnectionDetail
    vcs_connection={mockConnection}
    repos={{ total_count: 0, repositories: [] } as TVCSConnectionReposResponse}
  />
)
