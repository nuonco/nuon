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
  <div className="p-4 bg-cool-grey-950">
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
  <div className="p-4 bg-cool-grey-950">
    <DeploymentPlanGraph
      config={mockConfigIds}
      installsById={mockInstalls}
      orgId="org123"
    />
  </div>
)

export const SingleGroup = () => (
  <div className="p-4 bg-cool-grey-950">
    <DeploymentPlanGraph
      config={{ install_groups: [mockConfig.install_groups[0]] }}
      installsById={mockInstalls}
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
