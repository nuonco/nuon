export default {
  title: 'Install Components/InstallComponentBuildCard',
}

import { InstallComponentBuildCard, InstallComponentBuildCardSkeleton } from './InstallComponentBuildCard'

const mockBuild = {
  id: 'bld-abc123',
  component_id: 'comp-1',
  created_at: '2024-01-15T10:00:00Z',
  updated_at: '2024-01-15T10:05:00Z',
  status_v2: { status: 'active' },
} as any

export const Default = () => (
  <InstallComponentBuildCard
    build={mockBuild}
    orgId="org-1"
    installAppId="app-1"
    installAppConfigId="config-1"
  />
)

export const Loading = () => <InstallComponentBuildCardSkeleton />
