export default {
  title: 'Workflows/StepDetailPanel',
}

import { StepDetailPanel } from './StepDetailPanel'
import type { TWorkflowStep } from '@/types'

const noop = () => {}

const mockStep: TWorkflowStep = {
  id: 'step-1',
  name: 'deploy component',
  execution_type: 'system',
  step_target_type: 'install_deploys',
  step_target_id: 'deploy-1',
  install_workflow_id: 'wf-1',
  owner_id: 'inst-1',
  finished: false,
  started_at: '2024-01-01T00:00:00Z',
  execution_time: 0,
  status: { status: 'in-progress', history: [] },
  created_by: { email: 'user@example.com' },
} as TWorkflowStep

export const Default = () => (
  <StepDetailPanel
    step={mockStep}
    panelId="panel-1"
    panelKey="step-1"
    isVisible={true}
    onClose={noop}
  >
    <div>Step content goes here</div>
  </StepDetailPanel>
)
