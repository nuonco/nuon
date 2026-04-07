export default {
  title: 'Sandbox/SandboxHeader',
}

import { SandboxRunContext } from '@/providers/sandbox-run-provider'
import { SandboxHeader } from './SandboxHeader'
import type { TSandboxRun, TInstall, TWorkflow } from '@/types'

const mockSandboxRun: TSandboxRun = {
  id: 'sr-123',
  created_at: '2024-01-15T10:30:00Z',
  updated_at: '2024-01-15T10:35:00Z',
  status: 'active',
  status_v2: {
    status: 'active',
    status_human_description: 'Sandbox is active and running',
  },
  run_type: 'provision',
  install_workflow_id: 'wf-123',
  runner_jobs: [],
} as unknown as TSandboxRun

const mockInstall: TInstall = {
  id: 'install-1',
  name: 'prod-acme',
  org_id: 'org-1',
  app_id: 'app-1',
  app_config_id: 'config-1',
  cloud_platform: 'aws',
  app: { name: 'My App' },
} as unknown as TInstall

const mockWorkflow: TWorkflow = {
  id: 'wf-123',
  finished: false,
} as unknown as TWorkflow

export const Default = () => (
  <SandboxRunContext.Provider value={{ sandboxRun: mockSandboxRun }}>
    <SandboxHeader
      workflow={mockWorkflow}
      stepId="step-1"
      sandboxRun={mockSandboxRun}
      install={mockInstall}
      orgId="org-1"
    />
  </SandboxRunContext.Provider>
)
