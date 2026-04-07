export default {
  title: 'Builds/BuildTimeline',
}

import { BuildTimeline } from './BuildTimeline'
import type { TBuild } from '@/types'

const mockBuild: TBuild = {
  id: 'bld-123',
  created_at: '2024-01-15T10:30:00Z',
  updated_at: '2024-01-15T10:35:00Z',
  status: 'success',
  status_v2: { status: 'success' },
  created_by: { email: 'user@example.com' },
} as TBuild

const mockDriftedBuild: TBuild = {
  ...mockBuild,
  id: 'bld-456',
  created_at: '2024-01-14T09:00:00Z',
  updated_at: '2024-01-14T09:10:00Z',
  status: 'drifted',
  status_v2: { status: 'drifted' },
} as TBuild

const mockDuplicateBuild: TBuild = {
  ...mockBuild,
  id: 'bld-789',
  created_at: '2024-01-13T14:20:00Z',
  updated_at: '2024-01-13T14:25:00Z',
  status_v2: { status: 'success', metadata: { duplicate_build: true } },
} as TBuild

const mockBuildWithVcs: TBuild = {
  ...mockBuild,
  id: 'bld-abc',
  created_at: '2024-01-12T08:45:00Z',
  updated_at: '2024-01-12T08:50:00Z',
  vcs_connection_commit: {
    sha: 'abc123def456',
    message: 'fix: resolve deployment issue with config',
  },
} as TBuild

export const Default = () => (
  <BuildTimeline
    builds={[mockBuild, mockDriftedBuild, mockDuplicateBuild, mockBuildWithVcs]}
    pagination={{ hasNext: false, offset: 0, limit: 10 }}
    orgId="org-1"
    appId="app-1"
    componentId="comp-1"
    componentName="API server"
  />
)

export const Empty = () => (
  <BuildTimeline
    builds={[]}
    pagination={{ hasNext: false, offset: 0, limit: 10 }}
    orgId="org-1"
    appId="app-1"
    componentId="comp-1"
    componentName="API server"
  />
)

export const WithPagination = () => (
  <BuildTimeline
    builds={[mockBuild]}
    pagination={{ hasNext: true, offset: 0, limit: 10 }}
    orgId="org-1"
    appId="app-1"
    componentId="comp-1"
    componentName="API server"
  />
)
