export default {
  title: 'Workflows/RunnerStepDetails',
}

import { RunnerStepDetails } from './RunnerStepDetails'
import type { TWorkflowStep } from '@/types'

const mockStep = {
  id: 'step-1',
  name: 'provision runner',
  step_target_id: 'runner-1',
  step_target_type: 'runners',
  owner_id: 'inst-1',
  status: { status: 'in-progress', history: [] },
} as TWorkflowStep

export const Loading = () => (
  <RunnerStepDetails
    step={mockStep}
    orgId="org-123"
    isRunnerLoading={true}
    isHeartbeatLoading={true}
    isHealthCheckLoading={true}
  />
)
