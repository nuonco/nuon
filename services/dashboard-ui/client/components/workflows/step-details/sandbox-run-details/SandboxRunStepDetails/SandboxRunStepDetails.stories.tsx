export default {
  title: 'Workflows/SandboxRunStepDetails',
}

import { SandboxRunStepDetails } from './SandboxRunStepDetails'
import type { TWorkflowStep, TSandboxRun } from '@/types'

const mockStep = {
  id: 'step-1',
  name: 'run sandbox',
  step_target_type: 'install_sandbox_runs',
  step_target_id: 'run-1',
  owner_id: 'inst-1',
  execution_type: 'system',
  status: { status: 'in-progress', history: [] },
} as TWorkflowStep

const mockSandboxRun = {
  id: 'run-1',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:05:00Z',
  runner_jobs: [],
} as TSandboxRun

export const Default = () => (
  <SandboxRunStepDetails
    step={mockStep}
    orgId="org-123"
    sandboxRun={mockSandboxRun}
    isLoading={false}
  />
)

export const Loading = () => (
  <SandboxRunStepDetails
    step={mockStep}
    orgId="org-123"
    isLoading={true}
  />
)
