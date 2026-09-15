export default {
  title: 'Installs/AppSelect',
}

import { AppSelect } from './AppSelect'
import { appInstallBadge } from './app-install-readiness'
import type { TComponent } from '@/types'

const noop = () => {}

const mockApps = [
  {
    id: 'app-1',
    name: 'Production App',
    updated_at: '2024-01-15T00:00:00Z',
    runner_config: { app_runner_type: 'aws' },
    app_configs: [{ component_ids: ['cmp-1'] }],
  },
  {
    id: 'app-2',
    name: 'Staging App',
    updated_at: '2024-01-10T00:00:00Z',
    runner_config: { app_runner_type: 'azure' },
    app_configs: [{ component_ids: ['cmp-2'] }],
  },
  {
    id: 'app-3',
    name: 'Dev App',
    updated_at: '2024-01-05T00:00:00Z',
    runner_config: {},
  },
  {
    id: 'app-4',
    name: 'Sandbox only',
    updated_at: '2024-01-04T00:00:00Z',
    runner_config: { app_runner_type: 'aws' },
    app_configs: [{ component_ids: [] }],
  },
  {
    id: 'app-5',
    name: 'Unbuilt components',
    updated_at: '2024-01-03T00:00:00Z',
    runner_config: { app_runner_type: 'gcp' },
    app_configs: [{ component_ids: ['cmp-5'] }],
  },
] as any[]

const componentsByAppId: Record<string, TComponent[]> = {
  'app-1': [
    { id: 'cmp-1', latest_build: { status_v2: { status: 'active' } } } as TComponent,
  ],
  'app-2': [
    { id: 'cmp-2', latest_build: { status_v2: { status: 'active' } } } as TComponent,
  ],
  'app-5': [{ id: 'cmp-5' } as TComponent],
}

const badges = Object.fromEntries(
  mockApps.map((app) => [app.id, appInstallBadge(app, componentsByAppId[app.id])])
)

const baseProps = {
  isLoading: false,
  isLoadingMore: false,
  hasMorePages: false,
  error: null,
  searchQuery: '',
  onSearchChange: noop,
  onLoadMore: noop,
  onSelectApp: noop,
  onClose: noop,
}

export const Default = () => (
  <AppSelect apps={mockApps} badges={badges} {...baseProps} />
)

export const Loading = () => (
  <AppSelect apps={[]} {...baseProps} isLoading />
)

export const Empty = () => <AppSelect apps={[]} {...baseProps} />

export const WithSearch = () => (
  <AppSelect
    apps={mockApps.slice(0, 1)}
    badges={badges}
    {...baseProps}
    searchQuery="Production"
  />
)

export const WithError = () => (
  <AppSelect
    apps={[]}
    {...baseProps}
    error={{ error: 'Unable to load apps' }}
  />
)
