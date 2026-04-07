export default {
  title: 'Install Components/BuildSelect',
}

import { BuildSelect } from './BuildSelect'
import type { TBuild } from '@/types'

const noop = () => {}

const mockBuilds: TBuild[] = [
  {
    id: 'bld-abc123',
    status_v2: { status: 'active' },
    created_by: { email: 'dev@example.com' },
    created_at: '2024-01-15T10:30:00Z',
    vcs_connection_commit: { message: 'fix: update container config' },
    component_config_connection: {},
  } as TBuild,
  {
    id: 'bld-def456',
    status_v2: { status: 'active' },
    created_by: { email: 'dev@example.com' },
    created_at: '2024-01-14T08:00:00Z',
    vcs_connection_commit: { message: 'feat: add new endpoint' },
    component_config_connection: {},
  } as TBuild,
  {
    id: 'bld-ghi789',
    status_v2: { status: 'error' },
    created_by: { email: 'ci@example.com' },
    created_at: '2024-01-13T14:00:00Z',
    vcs_connection_commit: { message: 'chore: bump dependencies' },
    component_config_connection: {},
  } as TBuild,
]

export const Default = () => (
  <BuildSelect
    builds={mockBuilds}
    isLoading={false}
    isLoadingMore={false}
    hasMorePages={false}
    onSelectBuild={noop}
    onScroll={noop}
  />
)

export const WithSelection = () => (
  <BuildSelect
    builds={mockBuilds}
    selectedBuildId="bld-abc123"
    isLoading={false}
    isLoadingMore={false}
    hasMorePages={false}
    onSelectBuild={noop}
    onScroll={noop}
  />
)

export const WithCurrentDeployment = () => (
  <BuildSelect
    builds={mockBuilds}
    selectedBuildId="bld-abc123"
    currentBuildId="bld-abc123"
    currentDeployStatus="active"
    isLoading={false}
    isLoadingMore={false}
    hasMorePages={false}
    onSelectBuild={noop}
    onScroll={noop}
  />
)

export const Loading = () => (
  <BuildSelect
    builds={[]}
    isLoading={true}
    isLoadingMore={false}
    hasMorePages={false}
    onSelectBuild={noop}
    onScroll={noop}
  />
)

export const LoadingMore = () => (
  <BuildSelect
    builds={mockBuilds}
    isLoading={false}
    isLoadingMore={true}
    hasMorePages={true}
    onSelectBuild={noop}
    onScroll={noop}
  />
)

export const Empty = () => (
  <BuildSelect
    builds={[]}
    isLoading={false}
    isLoadingMore={false}
    hasMorePages={false}
    onSelectBuild={noop}
    onScroll={noop}
  />
)

export const WithError = () => (
  <BuildSelect
    builds={[]}
    isLoading={false}
    isLoadingMore={false}
    hasMorePages={false}
    error={{ error: 'Unable to load builds' }}
    onSelectBuild={noop}
    onScroll={noop}
  />
)
