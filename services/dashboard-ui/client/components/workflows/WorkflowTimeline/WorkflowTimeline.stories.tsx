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
  created_by: { email: 'developer@example.com' },
  app_branch_runs: [
    {
      head_sha: '83061cbabc123',
      metadata: { trigger: 'push' },
      vcs_connection_commit: {
        sha: '83061cbabc123',
        message: 'feat: add regional deployment controls',
        author_name: 'Example Developer',
        created_at: '2024-01-01T00:00:00Z',
      },
    },
  ],
} as unknown as TWorkflow

const manualBranchRunWorkflow: TWorkflow = {
  ...branchRunWorkflow,
  id: 'inwmanual1234567890abcdef',
  name: 'Manual run',
  created_by: { email: 'operator@example.com' },
  app_branch_runs: [
    {
      event_type: 'manual',
      metadata: { trigger: 'manual' },
      vcs_connection_commit: {
        sha: '5aeca0565df689db877c5aad7c0535e81264801d',
        message: 'chore: update branch configuration',
        author_name: 'Example Developer',
        created_at: '2024-01-01T00:00:00Z',
      },
    },
  ],
} as unknown as TWorkflow

const previewBranchRunWorkflow: TWorkflow = {
  ...branchRunWorkflow,
  id: 'inw5lhpdxb26o9qdrgp3zpq0zq',
  plan_only: true,
  app_branch_runs: [
    {
      plan_only: true,
      metadata: { trigger: 'pull_request' },
      preview: { mode: 'plan-only', source: 'commit' },
      vcs_connection_commit: {
        sha: '6987a43568abc8222c96d100284e50c045964258',
        message: 'feat: preview sandbox networking changes',
        author_name: 'Example Developer',
        created_at: '2024-01-01T00:00:00Z',
      },
    },
  ],
} as unknown as TWorkflow

const manualPreviewBranchRunWorkflow: TWorkflow = {
  ...previewBranchRunWorkflow,
  id: 'inwmanualpreview1234567890',
  created_by: { email: 'reviewer@example.com' },
  app_branch_runs: [
    {
      event_type: 'manual',
      metadata: { trigger: 'manual' },
      preview: { mode: 'plan-only', source: 'branch' },
      plan_only: true,
      vcs_connection_commit: {
        sha: 'b17c91f0643cadfe2834bc2409e975cc530e3611',
        message: 'fix: validate preview install inputs',
        author_name: 'Example Reviewer',
        created_at: '2024-01-01T00:00:00Z',
      },
    },
  ],
} as unknown as TWorkflow

const tagBranchRunWorkflow: TWorkflow = {
  ...branchRunWorkflow,
  id: 'inwtag1234567890abcdefghij',
  app_branch_runs: [
    {
      metadata: { trigger: 'tag', tag: 'v2.4.0' },
      vcs_connection_commit: {
        sha: '25abff839c9c640f49ea5dfa323f05de8980d5b1',
        message: 'release: v2.4.0',
        author_name: 'Example Developer',
        created_at: '2024-01-01T00:00:00Z',
      },
    },
  ],
} as unknown as TWorkflow

const labelBranchRunWorkflow: TWorkflow = {
  ...branchRunWorkflow,
  id: 'inwlabel1234567890abcdefgh',
  app_branch_runs: [
    {
      pr_number: 142,
      metadata: {
        trigger: 'github_label',
        pr_number: 142,
        github_label: 'deploy-preview',
      },
      vcs_connection_commit: {
        sha: 'd9c00a9a5344333c11658e6cef5d2a0a92d202b',
        message: 'feat: add checkout flow',
        author_name: 'Example Contributor',
        created_at: '2024-01-01T00:00:00Z',
      },
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

export const BranchRunTriggers = () => (
  <div className="max-w-4xl">
    <ApprovalsProvider>
      <WorkflowTimeline
        workflows={[
          manualBranchRunWorkflow,
          manualPreviewBranchRunWorkflow,
          previewBranchRunWorkflow,
          tagBranchRunWorkflow,
          branchRunWorkflow,
          labelBranchRunWorkflow,
        ]}
        pagination={{ hasNext: false, offset: 0, limit: 10 }}
        orgId="org-123"
        getWorkflowHref={(wf) =>
          `/org-123/apps/app-1/branches/branch-1/runs/${wf.id}`
        }
      />
    </ApprovalsProvider>
  </div>
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

export const FilteredEmpty = () => (
  <ApprovalsProvider>
    <WorkflowTimeline
      workflows={[]}
      pagination={{ hasNext: false, offset: 0, limit: 10 }}
      orgId="org-123"
      installId="inst-456"
      isFiltered
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
