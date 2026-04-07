export default {
  title: 'Sandbox/SandboxRunsTimeline',
}

import { SandboxRunsTimeline } from './SandboxRunsTimeline'
import type { TSandboxRun } from '@/types'

const mockRun: TSandboxRun = {
  id: 'sr-123',
  created_at: '2024-01-15T10:30:00Z',
  updated_at: '2024-01-15T10:35:00Z',
  status: 'active',
  status_v2: { status: 'active' },
  run_type: 'provision',
  created_by: { email: 'user@example.com' },
} as TSandboxRun

const mockDriftedRun: TSandboxRun = {
  ...mockRun,
  id: 'sr-456',
  created_at: '2024-01-14T10:30:00Z',
  updated_at: '2024-01-14T10:35:00Z',
  status: 'drifted',
  status_v2: { status: 'drifted' },
  run_type: 'drift_scan',
} as unknown as TSandboxRun

export const Default = () => (
  <SandboxRunsTimeline
    runs={[mockRun, mockDriftedRun]}
    pagination={{ hasNext: false, offset: 0, limit: 10 }}
    orgId="org-1"
    installId="install-1"
  />
)

export const Empty = () => (
  <SandboxRunsTimeline
    runs={[]}
    pagination={{ hasNext: false, offset: 0, limit: 10 }}
    orgId="org-1"
    installId="install-1"
  />
)

export const WithPagination = () => (
  <SandboxRunsTimeline
    runs={[mockRun]}
    pagination={{ hasNext: true, offset: 0, limit: 10 }}
    orgId="org-1"
    installId="install-1"
  />
)
