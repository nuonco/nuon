export default {
  title: 'VCS Connections/ConnectionDetails',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ConnectionDetailsModal } from './ConnectionDetails'

const mockConnection = {
  id: 'vcs-abc123',
  github_account_name: 'acme-corp',
  github_account_id: '12345',
  github_install_id: 'inst-1',
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  vcs_connection_commit: [],
} as any

export const Default = () => (
  <ModalStory>
    <ConnectionDetailsModal
      vcs_connection={mockConnection}
      status={{ status: 'active', checked_at: new Date().toISOString() }}
      isLoadingStatus={false}
      repos={{ repositories: [], total_count: 0 } as any}
      isLoadingRepos={false}
    />
  </ModalStory>
)

export const LoadingStatus = () => (
  <ModalStory>
    <ConnectionDetailsModal
      vcs_connection={mockConnection}
      isLoadingStatus={true}
      isLoadingRepos={true}
    />
  </ModalStory>
)

export const WithRepos = () => (
  <ModalStory>
    <ConnectionDetailsModal
      vcs_connection={mockConnection}
      status={{ status: 'active', checked_at: new Date().toISOString() }}
      isLoadingStatus={false}
      repos={{
        total_count: 2,
        repositories: [
          {
            id: 1,
            full_name: 'acme-corp/web-app',
            html_url: 'https://github.com/acme-corp/web-app',
            description: 'Main web application',
            private: true,
            fork: false,
            default_branch: 'main',
            updated_at: new Date().toISOString(),
          },
          {
            id: 2,
            full_name: 'acme-corp/infra',
            html_url: 'https://github.com/acme-corp/infra',
            description: 'Infrastructure as code',
            private: false,
            fork: false,
            default_branch: 'main',
            updated_at: new Date().toISOString(),
          },
        ],
      } as any}
      isLoadingRepos={false}
    />
  </ModalStory>
)
