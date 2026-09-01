export default {
  title: 'Branches/BranchSourceCard',
}

import { BranchSourceCard } from './BranchSourceCard'

export const Connected = () => (
  <BranchSourceCard
    config={{
      connected_github_vcs_config: {
        repo: 'acme/platform-configs',
        branch: 'main',
        directory: 'apps/api',
      },
    }}
    onEdit={() => {}}
  />
)

export const PublicRepo = () => (
  <BranchSourceCard
    config={{
      public_git_vcs_config: {
        repo: 'https://github.com/nuonco/example-app-configs',
        branch: 'main',
        directory: 'httpbin',
      },
    }}
  />
)

export const NoSource = () => <BranchSourceCard config={undefined} />

export const WithLatestRun = () => (
  <BranchSourceCard
    config={{
      public_git_vcs_config: {
        repo: 'nuonco/example-app-configs',
        branch: 'main',
        directory: 'httpbin',
      },
    }}
    latestRun={{
      status: 'success',
      href: '#',
      message: 'feat: add resources section to customer portal readme (#273)',
      author: 'Nat Hamilton',
      avatarUrl: 'https://github.com/nat.png',
      sha: '85d067ecafe1234',
      createdAt: '2026-08-12T09:00:00Z',
    }}
  />
)
