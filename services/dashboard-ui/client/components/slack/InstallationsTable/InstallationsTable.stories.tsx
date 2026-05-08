import { InstallationsTable } from './InstallationsTable'

export default { title: 'Slack/InstallationsTable' }

const mock = [
  {
    id: 'sli-001',
    team_id: 'T0123456789',
    team_name: 'nuonco',
    status: 'active' as const,
    created_at: '2026-04-30T15:00:00Z',
  },
  {
    id: 'sli-002',
    team_id: 'T9876543210',
    team_name: 'acme-corp',
    status: 'uninstalled' as const,
    created_at: '2026-03-20T15:00:00Z',
  },
]

const mockLinks = [
  {
    id: 'sol-001',
    team_id: 'T0123456789',
    org_id: 'org-demo',
    status: 'verified' as const,
    created_at: '2026-04-30T15:00:00Z',
  },
  {
    id: 'sol-002',
    team_id: 'T9876543210',
    org_id: 'org-demo',
    status: 'verified' as const,
    created_at: '2026-03-20T15:00:00Z',
  },
]

export const Default = () => (
  <InstallationsTable data={mock} links={mockLinks} isLoading={false} />
)
export const Loading = () => (
  <InstallationsTable data={[]} links={[]} isLoading={true} />
)
export const Empty = () => (
  <InstallationsTable data={[]} links={[]} isLoading={false} />
)
export const NoLinkForInstallation = () => (
  <InstallationsTable data={mock} links={[]} isLoading={false} />
)
