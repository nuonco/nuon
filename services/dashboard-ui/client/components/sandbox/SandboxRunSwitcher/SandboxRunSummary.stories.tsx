export default {
  title: 'Sandbox/SandboxRunSummary',
}

import { SandboxRunSummary } from './SandboxRunSummary'
import type { TSandboxRun } from '@/types'

const mockRun: TSandboxRun = {
  id: 'sanrun-abc123xyz456',
  created_at: new Date(Date.now() - 600000).toISOString(),
  status_v2: { status: 'active', status_human_description: 'Running' },
  created_by: { email: 'jane@example.com' },
} as TSandboxRun

export const Default = () => <SandboxRunSummary sandboxRun={mockRun} />

export const Latest = () => <SandboxRunSummary sandboxRun={mockRun} isLatest />

export const Error = () => (
  <SandboxRunSummary
    sandboxRun={{ ...mockRun, status_v2: { status: 'error', status_human_description: 'Failed' } }}
  />
)
