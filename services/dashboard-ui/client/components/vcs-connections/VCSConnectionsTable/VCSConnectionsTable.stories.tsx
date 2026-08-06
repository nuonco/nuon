export default {
  title: 'VCSConnections/VCSConnectionsTable',
}

import { VCSConnectionsTable, type TVCSConnectionRow } from './VCSConnectionsTable'

const rows: TVCSConnectionRow[] = [
  {
    connection: { id: 'vcs-1', github_account_name: 'powertoolsdev' },
    href: '/org-1/settings/vcs/vcs-1',
    status: 'active',
    checkedAt: '2026-08-06T09:00:00Z',
  },
  {
    connection: { id: 'vcs-2', github_account_name: 'nuonco-shared' },
    href: '/org-1/settings/vcs/vcs-2',
    status: 'suspended',
    checkedAt: '2026-08-05T12:30:00Z',
  },
  {
    connection: { id: 'vcs-3', github_account_name: 'nuonco' },
    href: '/org-1/settings/vcs/vcs-3',
    isLoadingStatus: true,
  },
]

export const Default = () => <VCSConnectionsTable data={rows} />

export const Loading = () => <VCSConnectionsTable data={[]} isLoading />

export const Empty = () => <VCSConnectionsTable data={[]} />
