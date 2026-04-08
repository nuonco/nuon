export default {
  title: 'VCS Connections/VCSAccountLink',
}

import { VCSAccountLink } from './VCSAccountLink'
import type { TVCSConnection } from '@/types'

const mockConnection: TVCSConnection = {
  id: 'vcs-1',
  github_account_name: 'nuonco',
  github_account_id: '12345678',
  github_install_id: '87654321',
} as TVCSConnection

export const Default = () => <VCSAccountLink vcs_connection={mockConnection} />

export const NoAccountName = () => (
  <VCSAccountLink
    vcs_connection={{ ...mockConnection, github_account_name: undefined } as TVCSConnection}
  />
)
