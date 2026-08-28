import { PanelStory } from '@/components/__stories__/helpers'
import { InstallVersionDetails } from './InstallVersionDetails'
import type { TInstallAppConfigVersion } from '@/types'

export default {
  title: 'InstallVersions/InstallVersionDetails',
}

const mockVersion = {
  id: 'iacv-3',
  created_at: new Date(Date.now() - 3600000).toISOString(),
  status: { status: 'active' },
  old_app_config_id: 'cfg-2',
  new_app_config_id: 'cfg-3',
  workflow_id: 'wkf-3',
  app_branch_run_id: 'abr-3',
  metadata: { triggered_by: 'app-branch' },
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
} as unknown as TInstallAppConfigVersion

export const Default = () => (
  <PanelStory>
    <InstallVersionDetails
      version={mockVersion}
      orgId="org-1"
      installId="inst-1"
      appId="app-1"
    />
  </PanelStory>
)

export const SyncTriggered = () => (
  <PanelStory>
    <InstallVersionDetails
      version={
        {
          id: 'iacv-1',
          created_at: new Date(Date.now() - 86400000).toISOString(),
          status: { status: 'active' },
          new_app_config_id: 'cfg-1',
        } as unknown as TInstallAppConfigVersion
      }
      orgId="org-1"
      installId="inst-1"
      appId="app-1"
    />
  </PanelStory>
)
