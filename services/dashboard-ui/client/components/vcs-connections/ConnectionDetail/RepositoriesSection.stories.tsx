export default {
  title: 'VCS Connections/RepositoriesSection',
}

import { RepositoriesSection } from './RepositoriesSection'
import type { TVCSConnectionReposResponse } from '@/types'

const mockRepos: TVCSConnectionReposResponse = {
  total_count: 6,
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
    {
      id: 4,
      full_name: 'nuonco/private-fork',
      html_url: 'https://github.com/nuonco/private-fork',
      description: 'A private fork',
      private: true,
      fork: true,
      default_branch: 'main',
      updated_at: new Date(Date.now() - 172800000).toISOString(),
    },
    {
      id: 5,
      full_name: 'nuonco/infra-modules',
      html_url: 'https://github.com/nuonco/infra-modules',
      description: 'Shared Terraform infrastructure modules',
      private: true,
      fork: false,
      default_branch: 'main',
      updated_at: new Date(Date.now() - 43200000).toISOString(),
    },
    {
      id: 6,
      full_name: 'nuonco/docs',
      html_url: 'https://github.com/nuonco/docs',
      description: null,
      private: false,
      fork: false,
      default_branch: 'main',
      updated_at: new Date(Date.now() - 604800000).toISOString(),
    },
  ],
} as TVCSConnectionReposResponse

export const Default = () => (
  <RepositoriesSection repos={mockRepos} isLoading={false} />
)

export const Loading = () => <RepositoriesSection isLoading={true} />

export const Empty = () => (
  <RepositoriesSection
    repos={{ total_count: 0, repositories: [] } as TVCSConnectionReposResponse}
    isLoading={false}
  />
)

export const Error = () => (
  <RepositoriesSection
    isLoading={false}
    error={{ error: 'Failed to load repositories. Check your GitHub connection.' }}
  />
)

export const SingleRepo = () => (
  <RepositoriesSection
    repos={{
      total_count: 1,
      repositories: [mockRepos.repositories![0]],
    } as TVCSConnectionReposResponse}
    isLoading={false}
  />
)
