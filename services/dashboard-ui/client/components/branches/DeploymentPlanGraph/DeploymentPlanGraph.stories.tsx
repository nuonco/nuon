import { DeploymentPlanGraph } from './DeploymentPlanGraph'

export default {
  title: 'Branches/DeploymentPlanGraph',
}

const mockInstalls: Record<string, any> = {
  'inst-001': { id: 'inst-001', name: 'acme-prod', labels: { env: 'production', tier: 'primary' } },
  'inst-002': { id: 'inst-002', name: 'acme-staging', labels: { env: 'staging', tier: 'primary' } },
  'inst-003': { id: 'inst-003', name: 'widget-prod', labels: { env: 'production', tier: 'secondary' } },
  'inst-004': { id: 'inst-004', name: 'widget-staging', labels: { env: 'staging' } },
  'inst-005': { id: 'inst-005', name: 'demo-env', labels: {} },
}

const mockConfig: any = {
  install_groups: [
    {
      id: 'group-staging',
      name: 'Staging',
      order: 0,
      label_selector: { match_labels: { env: 'staging' } },
      max_parallel: 2,
      use_for_previews: true,
    },
    {
      id: 'group-prod-primary',
      name: 'Production primary',
      order: 1,
      label_selector: { match_labels: { env: 'production', tier: 'primary' } },
      max_parallel: 1,
    },
    {
      id: 'group-prod-secondary',
      name: 'Production secondary',
      order: 2,
      label_selector: { match_labels: { env: 'production', tier: 'secondary' } },
      max_parallel: 1,
    },
  ],
}

export const ThreeGroups = () => (
  <div className="p-4">
    <DeploymentPlanGraph
      config={mockConfig}
      installsById={mockInstalls}
      orgId="org123"
    />
  </div>
)

const mockConfigIds: any = {
  install_groups: [
    {
      id: 'group-1',
      name: 'Canary',
      order: 0,
      install_ids: ['inst-005'],
      max_parallel: 1,
    },
    {
      id: 'group-2',
      name: 'All production',
      order: 1,
      install_ids: ['inst-001', 'inst-003'],
      max_parallel: 2,
    },
  ],
}

export const TwoGroupsById = () => (
  <div className="p-4">
    <DeploymentPlanGraph
      config={mockConfigIds}
      installsById={mockInstalls}
      orgId="org123"
    />
  </div>
)

export const SingleGroup = () => (
  <div className="p-4">
    <DeploymentPlanGraph
      config={{ install_groups: [mockConfig.install_groups[0]] }}
      installsById={mockInstalls}
      orgId="org123"
    />
  </div>
)

const longNameInstalls: Record<string, any> = {
  'inst-long': {
    id: 'inst-long',
    name: 'ws-workspace_01kzree4e4ejrsmaj9vbs4j0mg-sandbox-forge-runner-primary',
    labels: { 'auto-deploy': 'true' },
  },
}

const longNameConfig: any = {
  install_groups: [
    {
      id: 'group-main',
      name: 'main',
      order: 0,
      label_selector: { match_labels: { 'auto-deploy': 'true' } },
      max_parallel: 1,
    },
    {
      id: 'group-canary',
      name: 'canary',
      order: 1,
      label_selector: { match_labels: { canary: 'true' } },
      max_parallel: 1,
    },
  ],
}

export const LongInstallName = () => (
  <div className="p-4">
    <DeploymentPlanGraph
      config={longNameConfig}
      installsById={longNameInstalls}
      orgId="org123"
    />
  </div>
)

export const Empty = () => (
  <div className="p-4">
    <DeploymentPlanGraph
      config={{ install_groups: [] }}
      installsById={{}}
      orgId="org123"
    />
  </div>
)

const REGION_SUFFIXES = [
  'us-east-1', 'us-east-2', 'us-west-1', 'us-west-2', 'eu-west-1', 'eu-west-2',
  'eu-central-1', 'ap-south-1', 'ap-southeast-1', 'ap-southeast-2', 'ap-northeast-1',
  'ca-central-1', 'sa-east-1', 'af-south-1', 'me-south-1',
]

function buildLargePlan(
  specs: { name: string; count: number; maxParallel?: number; useForPreviews?: boolean }[]
) {
  const installsById: Record<string, any> = {}
  const install_groups = specs.map((spec, gi) => {
    const install_ids: string[] = []
    for (let i = 0; i < spec.count; i++) {
      const id = `inst-g${gi}-${i}`
      const region = REGION_SUFFIXES[i % REGION_SUFFIXES.length]
      installsById[id] = {
        id,
        name: `acme-${spec.name.toLowerCase().replace(/\s+/g, '-')}-${region}-${String(i + 1).padStart(2, '0')}`,
        labels: {},
      }
      install_ids.push(id)
    }
    return {
      id: `group-${gi}`,
      name: spec.name,
      order: gi,
      install_ids,
      max_parallel: spec.maxParallel ?? 1,
      use_for_previews: spec.useForPreviews ?? false,
    }
  })
  return { config: { install_groups } as any, installsById }
}

const largePlan = buildLargePlan([
  { name: 'Canary', count: 4, useForPreviews: true },
  { name: 'Production US', count: 30, maxParallel: 4 },
  { name: 'Production EU', count: 18, maxParallel: 3 },
  { name: 'Staging', count: 9, maxParallel: 2 },
])

export const ManyGroupsLargeInstallCounts = () => (
  <div className="p-4">
    <DeploymentPlanGraph
      config={largePlan.config}
      installsById={largePlan.installsById}
      orgId="org123"
    />
  </div>
)

export const ManyGroupsLargeInstallCountsCompact = () => (
  <div className="p-4">
    <DeploymentPlanGraph
      config={largePlan.config}
      installsById={largePlan.installsById}
      orgId="org123"
      compact
    />
  </div>
)
