import type { Meta, StoryObj } from '@ladle/react'
import { DeploySummary } from './DeploySummary'

export default {
  title: 'Deploys/DeploySwitcher/DeploySummary',
} satisfies Meta

const mockDeploy = {
  id: 'dep_abc123xyz456',
  created_at: '2024-01-15T10:30:00Z',
  created_by: { email: 'alice@example.com' },
  status_v2: { status: 'installed' },
} as any

export const Default: StoryObj = {
  render: () => <DeploySummary deploy={mockDeploy} />,
}

export const Latest: StoryObj = {
  render: () => <DeploySummary deploy={mockDeploy} isLatest />,
}

export const Deploying: StoryObj = {
  render: () => (
    <DeploySummary
      deploy={{ ...mockDeploy, status_v2: { status: 'deploying' } }}
    />
  ),
}
