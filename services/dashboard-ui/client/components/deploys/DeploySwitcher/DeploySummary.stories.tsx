export default {
  title: 'Deploys/DeploySwitcher/DeploySummary',
}

import { DeploySummary } from './DeploySummary'

const mockDeploy = {
  id: 'dep_abc123xyz456',
  created_at: '2024-01-15T10:30:00Z',
  created_by: { email: 'alice@example.com' },
  status_v2: { status: 'installed' },
} as any

export const Default = () => <DeploySummary deploy={mockDeploy} />

export const Latest = () => <DeploySummary deploy={mockDeploy} isLatest />

export const Deploying = () => (
  <DeploySummary
    deploy={{ ...mockDeploy, status_v2: { status: 'deploying' } }}
  />
)
