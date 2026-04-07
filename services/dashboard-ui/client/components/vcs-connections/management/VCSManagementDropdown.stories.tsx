export default {
  title: 'VCS Connections/VCSManagementDropdown',
}

import { VCSManagementDropdown } from './VCSManagementDropdown'
import type { TVCSConnection } from '@/types'

const mockConnection: TVCSConnection = {
  id: 'vcs-conn-1',
  github_account_name: 'nuonco',
  github_account_id: '12345678',
  github_install_id: '87654321',
  created_at: new Date(Date.now() - 86400000).toISOString(),
} as TVCSConnection

export const Default = () => <VCSManagementDropdown vcs_connection={mockConnection} />
