export default {
  title: 'Installs/AppSelect',
}

import { AppSelect } from './AppSelect'

const noop = () => {}

const mockApps = [
  {
    id: 'app-1',
    name: 'Production App',
    updated_at: '2024-01-15T00:00:00Z',
    runner_config: { app_runner_type: 'aws' },
  },
  {
    id: 'app-2',
    name: 'Staging App',
    updated_at: '2024-01-10T00:00:00Z',
    runner_config: { app_runner_type: 'azure' },
  },
  {
    id: 'app-3',
    name: 'Dev App',
    updated_at: '2024-01-05T00:00:00Z',
    runner_config: {},
  },
] as any[]

export const Default = () => (
  <AppSelect
    apps={mockApps}
    isLoading={false}
    isLoadingMore={false}
    hasMorePages={false}
    error={null}
    searchQuery=""
    onSearchChange={noop}
    onLoadMore={noop}
    onSelectApp={noop}
    onClose={noop}
  />
)

export const Loading = () => (
  <AppSelect
    apps={[]}
    isLoading={true}
    isLoadingMore={false}
    hasMorePages={true}
    error={null}
    searchQuery=""
    onSearchChange={noop}
    onLoadMore={noop}
    onSelectApp={noop}
    onClose={noop}
  />
)

export const Empty = () => (
  <AppSelect
    apps={[]}
    isLoading={false}
    isLoadingMore={false}
    hasMorePages={false}
    error={null}
    searchQuery=""
    onSearchChange={noop}
    onLoadMore={noop}
    onSelectApp={noop}
    onClose={noop}
  />
)

export const WithSearch = () => (
  <AppSelect
    apps={mockApps.slice(0, 1)}
    isLoading={false}
    isLoadingMore={false}
    hasMorePages={false}
    error={null}
    searchQuery="Production"
    onSearchChange={noop}
    onLoadMore={noop}
    onSelectApp={noop}
    onClose={noop}
  />
)

export const WithError = () => (
  <AppSelect
    apps={[]}
    isLoading={false}
    isLoadingMore={false}
    hasMorePages={false}
    error={{ error: 'Unable to load apps' }}
    searchQuery=""
    onSearchChange={noop}
    onLoadMore={noop}
    onSelectApp={noop}
    onClose={noop}
  />
)
