import { BranchRunSummary } from './BranchRunSummary'

export default {
  title: 'Branches/BranchRunSummary',
}

const mockBranchRun = {
  id: 'abrn1234567890abcdef12345',
  pr_number: 42,
  base_branch: 'main',
  event_type: 'pull_request',
  head_sha: 'abc123def456789012345678',
  vcs_connection_commit: {
    sha: 'abc123def456789012345678',
    message:
      'feat: update deployment config for staging\n\nAdded new environment variables',
    author_name: 'Jane Developer',
    author_email: 'jane@example.com',
  },
}

const mockBuilds = [
  {
    id: 'bld1234567890abcdef123456',
    component_name: 'web-frontend',
    component_id: 'cmp1234567890abcdef123456',
    component_config_connection: { type: 'helm_chart' },
    status_v2: { status: 'active' },
  },
  {
    id: 'bld2345678901abcdef234567',
    component_name: 'api-server',
    component_id: 'cmp2345678901abcdef234567',
    component_config_connection: { type: 'docker_build' },
    status_v2: { status: 'active' },
  },
]

const mockInstallGroupRuns = [
  {
    id: 'igr1234567890abcdef123456',
    install_group_name: 'staging',
    install_group_id: 'ig-staging',
    status: { status: 'success' },
    total_installs: 2,
    completed_installs: 2,
    failed_installs: 0,
    installs: [
      {
        install_id: 'inst12345678901234567890',
        status: 'success',
        workflow_id: 'wf-1',
      },
      {
        install_id: 'inst23456789012345678901',
        status: 'success',
        workflow_id: 'wf-2',
      },
    ],
  },
  {
    id: 'igr2345678901abcdef234567',
    install_group_name: 'production',
    install_group_id: 'ig-prod',
    status: { status: 'in-progress' },
    total_installs: 3,
    completed_installs: 1,
    failed_installs: 0,
    installs: [
      {
        install_id: 'inst34567890123456789012',
        status: 'success',
        workflow_id: 'wf-3',
      },
      {
        install_id: 'inst45678901234567890123',
        status: 'in-progress',
        workflow_id: 'wf-4',
      },
      {
        install_id: 'inst56789012345678901234',
        status: 'pending',
        workflow_id: 'wf-5',
      },
    ],
  },
]

export const Success = () => (
  <div className="max-w-3xl p-4">
    <BranchRunSummary
      branchRun={mockBranchRun as any}
      builds={mockBuilds as any}
      installUpdates={[]}
      installGroupRuns={mockInstallGroupRuns as any}
      orgId="org123"
      appId="app123"
      branchId="branch123"
      runStatus="success"
    />
  </div>
)

export const Failed = () => (
  <div className="max-w-3xl p-4">
    <BranchRunSummary
      branchRun={mockBranchRun as any}
      builds={mockBuilds as any}
      installUpdates={[]}
      installGroupRuns={[]}
      orgId="org123"
      appId="app123"
      branchId="branch123"
      runStatus="failed"
    />
  </div>
)

export const Empty = () => (
  <div className="max-w-3xl p-4">
    <BranchRunSummary
      branchRun={undefined}
      builds={[]}
      installUpdates={[]}
      installGroupRuns={[]}
      orgId="org123"
      appId="app123"
      branchId="branch123"
      runStatus="cancelled"
    />
  </div>
)
