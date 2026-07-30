import { RunDeploymentGraph } from './RunDeploymentGraph'

export default {
  title: 'Branches/RunDeploymentGraph',
}

const mockRuns: any[] = [
  {
    id: 'igr-001',
    install_group_id: 'group-staging',
    install_group_name: 'Staging',
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
    <RunDeploymentGraph installGroupRuns={mockRuns} orgId="org123" />
  </div>
)

const failedRuns: any[] = [
  {
    id: 'igr-003',
    install_group_id: 'group-canary',
    install_group_name: 'Canary',
    status: { status: 'error' },
    completed_installs: 0,
    total_installs: 1,
    installs: [{ install_id: 'inst-006', status: 'error' }],
  },
]

export const Failed = () => (
  <div className="p-4">
    <RunDeploymentGraph installGroupRuns={failedRuns} orgId="org123" />
  </div>
)

export const Empty = () => (
  <div className="p-4">
    <RunDeploymentGraph installGroupRuns={[]} orgId="org123" />
  </div>
)
