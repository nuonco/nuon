export default {
  title: 'Workflows/WorkflowTimeline',
}

import type { ReactNode } from 'react'
import { WorkflowApprovalsContext } from '@/providers/workflow-approvals-provider'
import { WorkflowTimeline, WorkflowTimelineSkeleton } from './WorkflowTimeline'
import type { TWorkflow } from '@/types'

const ApprovalsProvider = ({ children }: { children: ReactNode }) => (
  <WorkflowApprovalsContext.Provider
    value={{ approvals: [], isLoading: false, refresh: () => {} }}
  >
    {children}
  </WorkflowApprovalsContext.Provider>
)

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
  <ApprovalsProvider>
    <WorkflowTimeline
      workflows={[mockWorkflow, completedWorkflow]}
      pagination={{ hasNext: false, offset: 0, limit: 10 }}
      orgId="org-123"
      installId="inst-456"
    />
  </ApprovalsProvider>
)

export const Empty = () => (
  <ApprovalsProvider>
    <WorkflowTimeline
      workflows={[]}
      pagination={{ hasNext: false, offset: 0, limit: 10 }}
      orgId="org-123"
      installId="inst-456"
    />
  </ApprovalsProvider>
)

export const Loading = () => <WorkflowTimelineSkeleton />
