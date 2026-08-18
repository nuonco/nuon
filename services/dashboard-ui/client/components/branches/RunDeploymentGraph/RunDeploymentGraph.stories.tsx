import { RunDeploymentGraph } from './RunDeploymentGraph'

export default {
  title: 'Branches/RunDeploymentGraph',
}

const installsById: Record<string, any> = {
  'inst-001': { id: 'inst-001', name: 'acme-staging-eu' },
  'inst-002': { id: 'inst-002', name: 'acme-staging-us' },
  'inst-003': { id: 'inst-003', name: 'acme-prod-eu' },
  'inst-004': { id: 'inst-004', name: 'acme-prod-us' },
  'inst-005': { id: 'inst-005', name: 'acme-prod-ap' },
  'inst-006': { id: 'inst-006', name: 'acme-canary' },
  'inst-007': {
    id: 'inst-007',
    name: 'ws-workspace_01kzree4e4ejrsmaj9vbs4j0mg-sandbox-fd-aug-11',
  },
  'inst-008': { id: 'inst-008', name: 'acme-prod-payments' },
}

const mockRuns: any[] = [
  {
    id: 'igr-001',
    install_group_id: 'group-staging',
    install_group_name: 'Staging',
    install_group: { label_selector: { match_labels: { tier: 'staging' } } },
    status: { status: 'success' },
    completed_installs: 2,
    total_installs: 2,
    installs: [
      { install_id: 'inst-001', status: 'success' },
      { install_id: 'inst-002', status: 'success' },
    ],
  },
  {
    id: 'igr-002',
    install_group_id: 'group-prod',
    install_group_name: 'Production',
    install_group: { label_selector: { match_labels: { tier: 'prod', region: 'us-east-1' } } },
    status: { status: 'in-progress' },
    completed_installs: 1,
    total_installs: 3,
    installs: [
      { install_id: 'inst-003', status: 'success' },
      { install_id: 'inst-004', status: 'in-progress' },
      { install_id: 'inst-005', status: 'pending' },
    ],
  },
]

export const InProgress = () => (
  <div className="p-4">
    <RunDeploymentGraph installGroupRuns={mockRuns} installsById={installsById} orgId="org-demo" />
  </div>
)

export const WithoutInstallNames = () => (
  <div className="p-4">
    <RunDeploymentGraph installGroupRuns={mockRuns} orgId="org-demo" />
  </div>
)

const failedRuns: any[] = [
  {
    id: 'igr-003',
    install_group_id: 'group-canary',
    install_group_name: 'Canary',
    install_group: { label_selector: { match_labels: { canary: 'true' } } },
    status: { status: 'error' },
    completed_installs: 0,
    total_installs: 1,
    installs: [{ install_id: 'inst-006', status: 'error' }],
  },
]

export const Failed = () => (
  <div className="p-4">
    <RunDeploymentGraph installGroupRuns={failedRuns} installsById={installsById} orgId="org-demo" />
  </div>
)

const postDeployRunbookRuns: any[] = [
  {
    id: 'igr-004',
    install_group_id: 'group-prod',
    install_group_name: 'Production',
    install_group: { label_selector: { match_labels: { tier: 'prod' } } },
    status: { status: 'in-progress' },
    completed_installs: 1,
    total_installs: 2,
    installs: [
      {
        install_id: 'inst-007',
        status: 'success',
        phase: 'runbook',
        runbooks: [
          { runbook_id: 'rb1', runbook_name: 'db-migrate', status: 'success' },
          { runbook_id: 'rb2', runbook_name: 'smoke-test', status: 'success' },
        ],
      },
      {
        install_id: 'inst-008',
        status: 'error',
        phase: 'runbook',
        runbooks: [
          { runbook_id: 'rb1', runbook_name: 'db-migrate', status: 'success' },
          { runbook_id: 'rb2', runbook_name: 'smoke-test', status: 'error' },
        ],
      },
    ],
  },
]

export const WithPostDeployRunbooks = () => (
  <div className="p-4">
    <RunDeploymentGraph installGroupRuns={postDeployRunbookRuns} installsById={installsById} orgId="org-demo" />
  </div>
)

export const Empty = () => (
  <div className="p-4">
    <RunDeploymentGraph installGroupRuns={[]} orgId="org-demo" />
  </div>
)
