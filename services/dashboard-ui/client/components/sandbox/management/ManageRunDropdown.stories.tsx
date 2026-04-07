export default {
  title: 'Sandbox/ManageRunDropdown',
}

import { SandboxRunContext } from '@/providers/sandbox-run-provider'
import { ManageRunDropdown } from './ManageRunDropdown'
import type { TSandboxRun } from '@/types'

const mockSandboxRun: TSandboxRun = {
  id: 'run-001',
  install_id: 'inst-001',
  org_id: 'org-001',
  status: 'active',
  status_description: 'Running',
  status_v2: { status: 'active', status_human_description: 'Running' },
  runner_jobs: [
    { id: 'job-001' } as any,
    { id: 'job-002' } as any,
  ],
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const mockSandboxRunNoJobs: TSandboxRun = {
  ...mockSandboxRun,
  runner_jobs: [],
  status_v2: { status: 'error', status_human_description: 'Failed' },
}

const mockWorkflow = {
  id: 'wf-001',
  finished: false,
} as any

export const WithRunnerJobs = () => (
  <SandboxRunContext.Provider value={{ sandboxRun: mockSandboxRun }}>
    <ManageRunDropdown workflow={mockWorkflow} />
  </SandboxRunContext.Provider>
)

export const WithoutRunnerJobs = () => (
  <SandboxRunContext.Provider value={{ sandboxRun: mockSandboxRunNoJobs }}>
    <ManageRunDropdown />
  </SandboxRunContext.Provider>
)

export const WithCancelOption = () => (
  <SandboxRunContext.Provider
    value={{
      sandboxRun: {
        ...mockSandboxRun,
        status_v2: { status: 'pending', status_human_description: 'Waiting' },
      },
    }}
  >
    <ManageRunDropdown workflow={mockWorkflow} />
  </SandboxRunContext.Provider>
)
