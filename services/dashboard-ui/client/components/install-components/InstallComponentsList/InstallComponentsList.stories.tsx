export default {
  title: 'Install Components/InstallComponentsList',
}

import { Button } from '@/components/common/Button'
import { SearchInput } from '@/components/common/SearchInput'
import { LatestDeployCard } from '@/components/install-components/LatestDeployCard'
import { HealthTimelineComponent } from '@/components/install-health/HealthTimeline'
import type { TDeploy, THealthTimelineDay } from '@/types'
import {
  InstallComponentsList,
  type TInstallComponentListItem,
} from './InstallComponentsList'

const deploy = {
  id: 'dpl-1',
  status_v2: { status: 'active' },
  created_at: '2026-09-22T14:00:00Z',
  updated_at: '2026-09-22T14:04:00Z',
  install_deploy_type: 'apply',
} as TDeploy

const daily: THealthTimelineDay[] = Array.from({ length: 30 }, (_, index) => ({
  date: `2026-08-${String(index + 1).padStart(2, '0')}`,
  health: index === 12 ? 'degraded' : 'healthy',
  unhealthy_seconds: 0,
  degraded_seconds: index === 12 ? 1800 : 0,
  unknown_seconds: 0,
  observed_seconds: 86400,
}))

const health = (
  <HealthTimelineComponent
    scope="component"
    days={30}
    daily={daily}
    uptimePercent={99.93}
    observedSeconds={2592000}
    currentHealth="healthy"
  />
)

const deployAction = (
  <Button variant="secondary" size="sm">
    Deploy
  </Button>
)

const component = (
  overrides: Partial<TInstallComponentListItem>
): TInstallComponentListItem => ({
  id: 'cmp-1',
  name: 'api',
  type: 'helm_chart',
  status: 'active',
  deployAction,
  latestDeploy: <LatestDeployCard deploy={deploy} href="#" />,
  health,
  ...overrides,
})

const components: TInstallComponentListItem[] = [
  component({}),
  component({ id: 'cmp-2', name: 'worker' }),
  component({
    id: 'cmp-3',
    name: 'cache',
    type: 'terraform_module',
    status: 'in-progress',
    health: undefined,
  }),
]

export const Default = () => (
  <InstallComponentsList
    components={components}
    actions={<Button variant="secondary">Component controls</Button>}
    search={
      <SearchInput
        placeholder="Search by name or ID..."
        value=""
        onChange={() => {}}
      />
    }
    filterActions={<Button variant="secondary">Filter (5)</Button>}
  />
)

export const NoResults = () => (
  <InstallComponentsList
    components={[]}
    filtered
    search={
      <SearchInput
        placeholder="Search by name or ID..."
        value="api"
        onChange={() => {}}
      />
    }
    filterActions={<Button variant="secondary">Filter (1)</Button>}
  />
)

export const WithoutHealth = () => (
  <InstallComponentsList
    components={components.map((entry) => ({ ...entry, health: undefined }))}
  />
)

export const NeverDeployed = () => (
  <InstallComponentsList
    components={[
      component({
        latestDeploy: <LatestDeployCard />,
        health: undefined,
      }),
    ]}
  />
)

export const Empty = () => <InstallComponentsList components={[]} />

export const Loading = () => <InstallComponentsList components={[]} loading />
