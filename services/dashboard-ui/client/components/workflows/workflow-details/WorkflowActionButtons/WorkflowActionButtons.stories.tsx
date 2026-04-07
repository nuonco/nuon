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
    temporalLinkParams="?query=test"
    canShowApproveAll={true}
    canShowCancel={true}
    canShowTemporalLink={true}
  />
)

export const CancelOnly = () => (
  <WorkflowActionButtons
    workflow={mockWorkflow}
    temporalLinkParams=""
    canShowApproveAll={false}
    canShowCancel={true}
    canShowTemporalLink={false}
  />
)

export const Empty = () => (
  <WorkflowActionButtons
    workflow={mockWorkflow}
    temporalLinkParams=""
    canShowApproveAll={false}
    canShowCancel={false}
    canShowTemporalLink={false}
  />
)
