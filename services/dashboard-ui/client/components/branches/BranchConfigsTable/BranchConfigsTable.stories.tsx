import { BranchConfigsTable } from './BranchConfigsTable'
import type { TAppConfig } from '@/types'

export default {
  title: 'Branches/BranchConfigsTable',
}

const day = 86400000

const mockConfigs = [
  {
    id: 'cfg-3',
    version: 3,
    created_at: new Date(Date.now() - day).toISOString(),
    status_v2: { status: 'active' },
    cli_version: '0.19.2',
    component_ids: ['cmp-1', 'cmp-2'],
    action_ids: ['act-1'],
    runbook_ids: [],
    checksum: 'a1b2c3d4e5f67890',
    vcs_connection_commit: {
      sha: 'a1b2c3d4e5f6',
      message: 'Add a cache component to the deployment plan',
      author_name: 'Jane Doe',
    },
  },
  {
    id: 'cfg-2',
    version: 2,
    created_at: new Date(Date.now() - day * 4).toISOString(),
    status_v2: { status: 'outdated' },
    cli_version: '0.19.0',
    component_ids: ['cmp-1'],
    action_ids: [],
    runbook_ids: [],
  },
  {
    id: 'cfg-1',
    version: 1,
    created_at: new Date(Date.now() - day * 12).toISOString(),
    status_v2: { status: 'outdated' },
    cli_version: '0.18.4',
    component_ids: [],
    action_ids: [],
    runbook_ids: [],
  },
] as unknown as TAppConfig[]

export const Default = () => (
  <BranchConfigsTable configs={mockConfigs} appId="app-1" />
)

export const Empty = () => <BranchConfigsTable configs={[]} appId="app-1" />

export const Loading = () => (
  <BranchConfigsTable configs={[]} appId="app-1" isLoading />
)
