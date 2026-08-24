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

const REGION_SUFFIXES = [
  'us-east-1', 'us-east-2', 'us-west-1', 'us-west-2', 'eu-west-1', 'eu-west-2',
  'eu-central-1', 'ap-south-1', 'ap-southeast-1', 'ap-southeast-2', 'ap-northeast-1',
  'ca-central-1', 'sa-east-1', 'af-south-1', 'me-south-1',
]

const RUNBOOK_NAMES = ['db-migrate', 'smoke-test', 'cache-warm', 'notify-slack']

function buildLargeRun(
  specs: { name: string; count: number; completed: number; status: string; runbooks?: number }[]
) {
  const installsById: Record<string, any> = {}
  const runs: any[] = specs.map((spec, gi) => {
    const installs = Array.from({ length: spec.count }, (_, i) => {
      const id = `inst-g${gi}-${i}`
      const region = REGION_SUFFIXES[i % REGION_SUFFIXES.length]
      installsById[id] = {
        id,
        name: `acme-${spec.name.toLowerCase().replace(/\s+/g, '-')}-${region}-${String(i + 1).padStart(2, '0')}`,
      }
      let status = 'pending'
      if (i < spec.completed) status = 'success'
      else if (i === spec.completed && spec.status === 'in-progress') status = 'in-progress'
      else if (i === spec.completed && spec.status === 'error') status = 'error'
      const runbooks =
        spec.runbooks && status !== 'pending'
          ? Array.from({ length: spec.runbooks }, (_, r) => ({
              runbook_id: `rb-g${gi}-${i}-${r}`,
              runbook_name: RUNBOOK_NAMES[r % RUNBOOK_NAMES.length],
              status:
                status === 'success'
                  ? 'success'
                  : r < spec.runbooks! - 1
                    ? 'success'
                    : status,
            }))
          : []
      return { install_id: id, status, phase: runbooks.length ? 'runbook' : undefined, runbooks }
    })
    return {
      id: `igr-${gi}`,
      install_group_id: `group-${gi}`,
      install_group_name: spec.name,
      install_group: { label_selector: { match_labels: { group: spec.name.toLowerCase().replace(/\s+/g, '-') } } },
      status: { status: spec.status },
      completed_installs: spec.completed,
      total_installs: spec.count,
      installs,
    }
  })
  return { runs, installsById }
}

const largeRun = buildLargeRun([
  { name: 'Canary', count: 4, completed: 4, status: 'success' },
  { name: 'Production US', count: 30, completed: 12, status: 'in-progress' },
  { name: 'Production EU', count: 18, completed: 0, status: 'pending' },
  { name: 'Staging', count: 9, completed: 5, status: 'error' },
])

export const ManyGroupsLargeInstallCounts = () => (
  <div className="p-4">
    <RunDeploymentGraph
      installGroupRuns={largeRun.runs}
      installsById={largeRun.installsById}
      orgId="org-demo"
    />
  </div>
)

const largeRunWithRunbooks = buildLargeRun([
  { name: 'Canary', count: 4, completed: 4, status: 'success', runbooks: 2 },
  { name: 'Production US', count: 30, completed: 12, status: 'in-progress', runbooks: 3 },
  { name: 'Production EU', count: 18, completed: 0, status: 'pending', runbooks: 2 },
  { name: 'Staging', count: 9, completed: 5, status: 'error', runbooks: 4 },
])

export const ManyGroupsLargeInstallCountsWithRunbooks = () => (
  <div className="p-4">
    <RunDeploymentGraph
      installGroupRuns={largeRunWithRunbooks.runs}
      installsById={largeRunWithRunbooks.installsById}
      orgId="org-demo"
    />
  </div>
)
