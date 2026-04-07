export default {
  title: 'VCS Connections/GitHubAccountSection',
}

import { GitHubAccountSection } from './GitHubAccountSection'
import type { TVCSConnection } from '@/types'

const mockConnection: TVCSConnection = {
  id: 'vcs-conn-abc123',
  github_account_name: 'nuonco',
  github_account_id: '12345678',
  github_install_id: '87654321',
  created_at: new Date(Date.now() - 86400000 * 30).toISOString(),
  updated_at: new Date(Date.now() - 3600000).toISOString(),
  vcs_connection_commit: [],
} as TVCSConnection

export const Default = () => <GitHubAccountSection vcs_connection={mockConnection} />

export const WithCommits = () => (
  <GitHubAccountSection
    vcs_connection={{
      ...mockConnection,
      vcs_connection_commit: [{ id: 'c1' }, { id: 'c2' }] as any,
    }}
  />
)
