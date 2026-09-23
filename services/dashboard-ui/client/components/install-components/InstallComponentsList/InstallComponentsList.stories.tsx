export default {
  title: 'Install Components/InstallComponentsList',
}

import { Button } from '@/components/common/Button'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { LatestDeployCard } from '@/components/install-components/LatestDeployCard'
import { HealthTimelineComponent } from '@/components/install-health/HealthTimeline'
import type { TComponentBuild, TDeploy, THealthTimelineDay } from '@/types'
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

const build = {
  id: 'bld-1',
  status_v2: { status: 'active' },
  created_at: '2026-09-22T13:41:00Z',
  vcs_connection_commit: {
    sha: '9c2f7a1b4d8e6350af19c4b7d2e058f36a1b9c4d',
    message: 'Pin acme-api chart to 2.4.0',
    author_name: 'Ada Lovelace',
  },
} as TComponentBuild

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

const details = (showHealth = true, latestDeploy: TDeploy | null = deploy) => (
  <>
    <div className="flex flex-col gap-2">
      <Text variant="subtext" weight="strong" theme="neutral">
        Latest deploy
      </Text>
      <LatestDeployCard
        flush
        deploy={latestDeploy ?? undefined}
        build={latestDeploy ? build : undefined}
        buildHref={latestDeploy ? '#' : undefined}
        href={latestDeploy ? '#' : undefined}
      />
    </div>
    {showHealth && latestDeploy ? (
      <div className="flex flex-col gap-2 border-t pt-4">
        <Text variant="subtext" weight="strong" theme="neutral">
          Health
        </Text>
        {health}
      </div>
    ) : null}
  </>
)

const actions = (
  <Button variant="secondary" size="sm">
    More
  </Button>
)

const component = (
  overrides: Partial<TInstallComponentListItem>
): TInstallComponentListItem => ({
  id: 'cmp-1',
  name: 'api',
  type: 'helm_chart',
  status: 'active',
  actions,
  latestDeploy: details(),
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
    latestDeploy: details(false),
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
    pagination={{ hasNext: true, offset: 0, limit: 10 }}
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
    components={components.map((entry) => ({
      ...entry,
      latestDeploy: details(false),
    }))}
  />
)

export const NeverDeployed = () => (
  <InstallComponentsList
    components={[
      component({
        latestDeploy: details(false, null),
      }),
    ]}
  />
)

export const WithDisabledComponent = () => (
  <InstallComponentsList
    components={[
      component({ enabled: true }),
      component({
        id: 'cmp-2',
        name: 'worker',
        type: 'job',
        enabled: false,
        status: 'inactive',
        latestDeploy: details(false),
      }),
      component({
        id: 'cmp-3',
        name: 'cache',
        type: 'terraform_module',
        latestDeploy: details(false),
      }),
    ]}
  />
)

export const AllComponentsDisabled = () => (
  <InstallComponentsList
    components={[
      component({ enabled: false, status: 'inactive' }),
      component({
        id: 'cmp-2',
        name: 'worker',
        type: 'job',
        enabled: false,
        status: 'inactive',
        latestDeploy: details(false, null),
      }),
    ]}
  />
)

export const Empty = () => <InstallComponentsList components={[]} />

export const Loading = () => <InstallComponentsList components={[]} loading />
