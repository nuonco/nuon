export default {
  title: 'Workflows/WorkflowActionButtons',
}

import { WorkflowActionButtons } from './WorkflowActionButtons'
import type { TWorkflow } from '@/types'

const mockWorkflow = {
  id: 'wf-123',
  owner_id: 'inst-456',
  type: 'deploy_components',
  status: { status: 'in-progress' },
  finished: false,
  approval_option: 'prompt',
} as TWorkflow

export const AllButtons = () => (
  <WorkflowActionButtons
    workflow={mockWorkflow}
    canShowApproveAll={true}
    canShowCancel={true}
  />
)

export const CancelOnly = () => (
  <WorkflowActionButtons
    workflow={mockWorkflow}
    canShowApproveAll={false}
    canShowCancel={true}
  />
)

export const Empty = () => (
  <WorkflowActionButtons
    workflow={mockWorkflow}
    canShowApproveAll={false}
    canShowCancel={false}
  />
)
