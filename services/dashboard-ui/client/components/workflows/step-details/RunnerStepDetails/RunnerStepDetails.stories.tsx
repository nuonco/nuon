export default {
  title: 'Workflows/RunnerStepDetails',
}

import { RunnerStepDetails } from './RunnerStepDetails'
import type { TWorkflowStep, TRunnerProcess } from '@/types'

const mockStep = {
  id: 'step-1',
  name: 'provision runner',
  step_target_id: 'runner-1',
  step_target_type: 'runners',
  owner_id: 'inst-1',
  status: { status: 'in-progress', history: [] },
} as TWorkflowStep

const mockProcess = {
  id: 'proc-123',
  type: 'runner',
  composite_status: { status: 'active' },
  version: '1.2.3',
  labels: ['primary'],
  started_at: '2024-01-15T08:00:00Z',
  warnings: [],
} as unknown as TRunnerProcess

export const Loading = () => (
  <RunnerStepDetails
    step={mockStep}
    orgId="org-123"
    processes={[]}
    processesLoading={true}
  />
)

export const Empty = () => (
  <RunnerStepDetails
    step={mockStep}
    orgId="org-123"
    processes={[]}
    processesLoading={false}
  />
)

export const SingleProcess = () => (
  <RunnerStepDetails
    step={mockStep}
    orgId="org-123"
    processes={[mockProcess]}
    processesLoading={false}
  />
)

export const TwoProcesses = () => (
  <RunnerStepDetails
    step={mockStep}
    orgId="org-123"
    processes={[
      mockProcess,
      { ...mockProcess, id: 'proc-456' } as TRunnerProcess,
    ]}
    processesLoading={false}
  />
)
