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
        path_filter: 'apps/api/**',
      },
    }}
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
