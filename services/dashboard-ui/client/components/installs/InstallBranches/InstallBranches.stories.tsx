import { InstallBranches } from './InstallBranches'

export default {
  title: 'Installs/InstallBranches',
}

const mockBranches: any[] = [
  {
    branchId: 'branch-001',
    branchName: 'main',
    active: true,
    activatedAt: '2026-07-01T10:00:00Z',
    latestRun: {
      id: 'wf-001',
      status: { status: 'success', status_human_description: 'Completed' },
      finished_at: '2026-07-08T15:30:00Z',
      app_branch_runs: [
        {
          id: 'abrn-001',
          vcs_connection_commit: {
            sha: 'abc123def456789012345678',
            message: 'feat: update deployment config',
            author_name: 'Jane Developer',
          },
        },
      ],
    },
    branchRun: {
      id: 'abrn-001',
      vcs_connection_commit: {
        sha: 'abc123def456789012345678',
        message: 'feat: update deployment config',
        author_name: 'Jane Developer',
      },
    },
    builds: [
      { id: 'bld-001', component_name: 'web-frontend', status_v2: { status: 'active' } },
      { id: 'bld-002', component_name: 'api-server', status_v2: { status: 'active' } },
    ],
    installUpdates: [
      { id: 'iacv-001', install_id: 'inst-001', workflow_id: 'wf-100', status: { status: 'success' } },
    ],
  },
  {
    branchId: 'branch-002',
    branchName: 'staging',
    active: true,
    activatedAt: '2026-06-15T08:00:00Z',
    latestRun: undefined,
    branchRun: undefined,
    builds: [],
    installUpdates: [],
  },
  {
    branchId: 'branch-003',
    branchName: 'legacy-deploy',
    active: false,
    activatedAt: '2026-05-01T12:00:00Z',
    latestRun: {
      id: 'wf-003',
      status: { status: 'failed', status_human_description: 'Build failed' },
      finished_at: '2026-06-20T09:15:00Z',
      app_branch_runs: [
        {
          id: 'abrn-003',
          vcs_connection_commit: {
            sha: 'def456abc789012345678901',
            message: 'fix: rollback config change',
            author_name: 'Bob Engineer',
          },
        },
      ],
    },
    branchRun: {
      id: 'abrn-003',
      vcs_connection_commit: {
        sha: 'def456abc789012345678901',
        message: 'fix: rollback config change',
        author_name: 'Bob Engineer',
      },
    },
    builds: [
      { id: 'bld-003', component_name: 'api-server', status_v2: { status: 'error' } },
    ],
    installUpdates: [],
  },
]

export const WithBranches = () => (
  <div className="max-w-2xl p-4">
    <InstallBranches
      branches={mockBranches}
      orgId="org123"
      appId="app123"
      installId="inst-001"
    />
  </div>
)

export const Empty = () => (
  <div className="max-w-2xl p-4">
    <InstallBranches branches={[]} orgId="org123" appId="app123" installId="inst-001" />
  </div>
)
