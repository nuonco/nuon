export default {
  title: 'Workflows/WorkflowTimeline',
}

import { WorkflowTimeline, WorkflowTimelineSkeleton } from './WorkflowTimeline'
import type { TWorkflow } from '@/types'

const mockWorkflow: TWorkflow = {
  id: 'wf-123',
  name: 'Deploy app',
  type: 'deploy_components',
  plan_only: false,
  finished: false,
  approval_option: 'prompt',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:05:00Z',
  execution_time: 300000000000,
  status: { status: 'in-progress' },
  created_by: { email: 'user@example.com' },
  metadata: {},
} as TWorkflow

const completedWorkflow: TWorkflow = {
  ...mockWorkflow,
  id: 'wf-456',
  name: 'Provision runner',
  type: 'provision',
  finished: true,
  status: { status: 'success' },
} as TWorkflow

export const Default = () => (
  <WorkflowTimeline
    workflows={[mockWorkflow, completedWorkflow]}
    pagination={{ hasNext: false, offset: 0, limit: 10 }}
    orgId="org-123"
    installId="inst-456"
  />
)

export const Empty = () => (
  <WorkflowTimeline
    workflows={[]}
    pagination={{ hasNext: false, offset: 0, limit: 10 }}
    orgId="org-123"
    installId="inst-456"
  />
)

export const Loading = () => <WorkflowTimelineSkeleton />
