import { InstallVersionsTimeline } from './InstallVersionsTimeline'
import type { TInstallAppConfigVersion } from '@/types'

export default {
  title: 'InstallVersions/InstallVersionsTimeline',
}

const day = 86400000

const mockVersions = [
  {
    id: 'iacv-3',
    created_at: new Date(Date.now() - day).toISOString(),
    status: { status: 'active' },
    old_app_config_id: 'cfg-2',
    new_app_config_id: 'cfg-3',
    workflow_id: 'wkf-3',
    app_branch_run_id: 'abr-3',
    app_branch_run: {
      pr_number: 42,
      workflow_id: 'wkf-3',
      app_branch: { id: 'branch-1', name: 'feat/add-cache' },
      vcs_connection_commit: {
        sha: 'a1b2c3d4e5f6',
        message: 'Add a cache component to the deployment plan',
        author_name: 'Jane Doe',
      },
    },
  },
  {
    id: 'iacv-2',
    created_at: new Date(Date.now() - day * 3).toISOString(),
    status: { status: 'in-progress' },
    workflow: { status: { status: 'error' } },
    old_app_config_id: 'cfg-1',
    new_app_config_id: 'cfg-2',
    metadata: { triggered_by: 'cli-sync' },
  },
  {
    id: 'iacv-1',
    created_at: new Date(Date.now() - day * 9).toISOString(),
    status: { status: 'active' },
    new_app_config_id: 'cfg-1',
  },
] as unknown as TInstallAppConfigVersion[]

export const Default = () => (
  <InstallVersionsTimeline
    versions={mockVersions}
    orgId="org-1"
    installId="inst-1"
    appId="app-1"
  />
)

export const Empty = () => <InstallVersionsTimeline versions={[]} />

export const Loading = () => <InstallVersionsTimeline versions={[]} isLoading />
