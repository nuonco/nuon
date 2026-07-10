import { BranchRunSummary } from './BranchRunSummary'

export default {
  title: 'Branches/BranchRunSummary',
}

const mockBranchRun = {
  id: 'abrn1234567890abcdef12345',
  vcs_connection_commit: {
    sha: 'abc123def456789012345678',
    message: 'feat: update deployment config for staging\n\nAdded new environment variables',
    author_name: 'Jane Developer',
    author_email: 'jane@example.com',
  },
}

const mockBuilds = [
  {
    id: 'bld1234567890abcdef123456',
    component_name: 'web-frontend',
    component_id: 'cmp1234567890abcdef123456',
    status_v2: { status: 'active' },
  },
  {
    id: 'bld2345678901abcdef234567',
    component_name: 'api-server',
    component_id: 'cmp2345678901abcdef234567',
    status_v2: { status: 'active' },
  },
]

const mockInstallUpdates = [
  {
    id: 'iacv123456789012345678901',
    install_id: 'inst12345678901234567890',
    workflow_id: 'wf12345678901234567890ab',
    status: { status: 'success' },
  },
  {
    id: 'iacv234567890123456789012',
    install_id: 'inst23456789012345678901',
    workflow_id: 'wf23456789012345678901ab',
    status: { status: 'active' },
  },
]

export const Success = () => (
  <div className="max-w-3xl p-4">
    <BranchRunSummary
      branchRun={mockBranchRun as any}
      builds={mockBuilds as any}
      installUpdates={mockInstallUpdates as any}
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
      orgId="org123"
      appId="app123"
      branchId="branch123"
      runStatus="cancelled"
    />
  </div>
)
