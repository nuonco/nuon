export default {
  title: 'Workflows/WorkflowTimeline',
}

import type { ReactNode } from 'react'
import { WorkflowApprovalsContext } from '@/providers/workflow-approvals-provider'
import { WorkflowTimeline } from './WorkflowTimeline'
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

const unnamedWorkflow: TWorkflow = {
  ...mockWorkflow,
  id: 'wfq7fplr1up5atx5zpxotbabm',
  name: undefined,
  type: 'provision',
  finished: true,
  status: { status: 'success' },
} as TWorkflow

const branchRunWorkflow: TWorkflow = {
  ...mockWorkflow,
  id: 'inwykcx31yhqs8sb3w9ynmuig3',
  name: 'Run',
  type: 'app_branch_run',
  owner_type: 'app_branches',
  finished: true,
  status: { status: 'success' },
  created_by: { email: 'nat@nuon.co' },
  app_branch_runs: [{ head_sha: '83061cbabc123' }],
} as unknown as TWorkflow

const manualBranchRunWorkflow: TWorkflow = {
  ...branchRunWorkflow,
  id: 'inwmanual1234567890abcdef',
  app_branch_runs: [{ head_sha: '83061cbabc123', event_type: 'manual' }],
} as unknown as TWorkflow

const previewBranchRunWorkflow: TWorkflow = {
  ...branchRunWorkflow,
  id: 'inw5lhpdxb26o9qdrgp3zpq0zq',
  plan_only: true,
} as TWorkflow

const manualPreviewBranchRunWorkflow: TWorkflow = {
  ...previewBranchRunWorkflow,
  id: 'inwmanualpreview1234567890',
  app_branch_runs: [
    {
      head_sha: '83061cbabc123',
      event_type: 'manual',
      preview: { mode: 'plan-only' },
      plan_only: true,
    },
  ],
} as unknown as TWorkflow

const serviceAccountBranchRunWorkflow: TWorkflow = {
  ...previewBranchRunWorkflow,
  id: 'inwm6oxsvulflygj3z77xo55l',
  created_by: {
    email:
      'orgl9cvkaqh1g8yv2jqdb19247-oidc-accbx92unf2s9wi0ve72bwlfzi@serviceaccount.nuon.co',
    account_type: 'service',
  },
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

export const NoName = () => (
  <ApprovalsProvider>
    <WorkflowTimeline
      workflows={[unnamedWorkflow]}
      pagination={{ hasNext: false, offset: 0, limit: 10 }}
      orgId="org-123"
      installId="inst-456"
    />
  </ApprovalsProvider>
)

export const BranchRuns = () => (
  <ApprovalsProvider>
    <WorkflowTimeline
      workflows={[
        serviceAccountBranchRunWorkflow,
        previewBranchRunWorkflow,
        branchRunWorkflow,
      ]}
      pagination={{ hasNext: false, offset: 0, limit: 10 }}
      orgId="org-123"
      getWorkflowHref={(wf) =>
        `/org-123/apps/app-1/branches/branch-1/runs/${wf.id}`
      }
    />
  </ApprovalsProvider>
)

export const BranchRunPreview = () => (
  <ApprovalsProvider>
    <WorkflowTimeline
      workflows={[previewBranchRunWorkflow]}
      pagination={{ hasNext: false, offset: 0, limit: 10 }}
      orgId="org-123"
      getWorkflowHref={(wf) =>
        `/org-123/apps/app-1/branches/branch-1/runs/${wf.id}`
      }
    />
  </ApprovalsProvider>
)

export const BranchRunManual = () => (
  <ApprovalsProvider>
    <WorkflowTimeline
      workflows={[manualBranchRunWorkflow]}
      pagination={{ hasNext: false, offset: 0, limit: 10 }}
      orgId="org-123"
      getWorkflowHref={(wf) =>
        `/org-123/apps/app-1/branches/branch-1/runs/${wf.id}`
      }
    />
  </ApprovalsProvider>
)

export const BranchRunManualPreview = () => (
  <ApprovalsProvider>
    <WorkflowTimeline
      workflows={[manualPreviewBranchRunWorkflow]}
      pagination={{ hasNext: false, offset: 0, limit: 10 }}
      orgId="org-123"
      getWorkflowHref={(wf) =>
        `/org-123/apps/app-1/branches/branch-1/runs/${wf.id}`
      }
    />
  </ApprovalsProvider>
)

export const BranchRunServiceAccount = () => (
  <ApprovalsProvider>
    <WorkflowTimeline
      workflows={[serviceAccountBranchRunWorkflow]}
      pagination={{ hasNext: false, offset: 0, limit: 10 }}
      orgId="org-123"
      getWorkflowHref={(wf) =>
        `/org-123/apps/app-1/branches/branch-1/runs/${wf.id}`
      }
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

export const Loading = () => (
  <ApprovalsProvider>
    <WorkflowTimeline
      workflows={[]}
      pagination={{ hasNext: false, offset: 0, limit: 10 }}
      orgId="org-123"
      installId="inst-456"
      isLoading
    />
  </ApprovalsProvider>
)
