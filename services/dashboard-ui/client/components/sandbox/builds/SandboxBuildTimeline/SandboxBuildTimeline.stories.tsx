export default {
  title: 'Sandbox/SandboxBuildTimeline',
}

import { SandboxBuildTimeline } from './SandboxBuildTimeline'
import type { TAppSandboxBuild } from '@/types'

const mockBuild: TAppSandboxBuild = {
  id: 'bld-123',
  created_at: '2024-01-15T10:30:00Z',
  updated_at: '2024-01-15T10:35:00Z',
  status: 'active',
  status_v2: { status: 'active', metadata: {} },
  created_by: { email: 'user@example.com' },
  vcs_connection_commit: {
    sha: 'abc123def456',
    message: 'feat: add new terraform module',
  },
} as unknown as TAppSandboxBuild

const mockDriftedBuild: TAppSandboxBuild = {
  ...mockBuild,
  id: 'bld-456',
  status: 'drifted',
  status_v2: { status: 'drifted', metadata: {} },
} as unknown as TAppSandboxBuild

const mockDuplicateBuild: TAppSandboxBuild = {
  ...mockBuild,
  id: 'bld-789',
  status_v2: { status: 'active', metadata: { duplicate_build: true } },
} as unknown as TAppSandboxBuild

export const Default = () => (
  <SandboxBuildTimeline
    builds={[mockBuild, mockDriftedBuild, mockDuplicateBuild]}
    pagination={{ hasNext: false, offset: 0, limit: 10 }}
    orgId="org-1"
    appId="app-1"
    isEmpty={false}
  />
)

export const Empty = () => (
  <SandboxBuildTimeline
    builds={[]}
    pagination={{ hasNext: false, offset: 0, limit: 10 }}
    orgId="org-1"
    appId="app-1"
    isEmpty={true}
  />
)

export const WithPagination = () => (
  <SandboxBuildTimeline
    builds={[mockBuild]}
    pagination={{ hasNext: true, offset: 0, limit: 10 }}
    orgId="org-1"
    appId="app-1"
    isEmpty={false}
  />
)
