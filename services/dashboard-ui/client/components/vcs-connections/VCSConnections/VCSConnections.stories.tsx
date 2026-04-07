export default {
  title: 'VCS Connections/VCSConnections',
}

import { VCSConnections as VCSConnectionsComponent } from './VCSConnections'

const mockConnections = [
  {
    id: 'vcs-1',
    github_account_name: 'acme-corp',
    github_account_id: '12345',
    github_install_id: 'inst-1',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'vcs-2',
    github_account_name: 'other-org',
    github_account_id: '67890',
    github_install_id: 'inst-2',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
] as any

export const Default = () => (
  <VCSConnectionsComponent
    vcsConnections={mockConnections}
    statusMap={{
      'vcs-1': { theme: 'success', isLoading: false },
      'vcs-2': { theme: 'success', isLoading: false },
    }}
  />
)

export const Loading = () => (
  <VCSConnectionsComponent
    vcsConnections={mockConnections}
    statusMap={{
      'vcs-1': { theme: undefined, isLoading: true },
      'vcs-2': { theme: undefined, isLoading: true },
    }}
  />
)
