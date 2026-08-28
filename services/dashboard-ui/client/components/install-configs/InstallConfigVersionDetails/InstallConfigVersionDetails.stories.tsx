import { PanelStory } from '@/components/__stories__/helpers'
import { InstallConfigVersionDetails } from './InstallConfigVersionDetails'
import type { TConfigDiffNode, TInstallConfigVersion } from '@/types'

export default {
  title: 'InstallConfigs/InstallConfigVersionDetails',
}

const mockVersion = {
  id: 'icv-3',
  created_at: new Date(Date.now() - 3600000).toISOString(),
  created: false,
  file_path: 'installs/payments.toml',
  status: { status: 'active' },
  install_config_sync: {
    triggered_by: 'vcs-webhook',
    vcs_connection_commit: {
      sha: 'f00ba12345678',
      message: 'Bump the payments install to two replicas',
      author_name: 'Jane Doe',
    },
  },
} as unknown as TInstallConfigVersion

const mockDiff = {
  key: 'install',
  children: [
    {
      key: 'inputs',
      children: [
        { key: 'replicas', diff: { op: 'change', diff: '1 -> 2' } },
        { key: 'region', diff: { op: 'add', diff: 'us-west-2' } },
        { key: 'legacy_flag', diff: { op: 'remove', diff: 'true' } },
      ],
    },
  ],
} as unknown as TConfigDiffNode

export const Default = () => (
  <PanelStory>
    <InstallConfigVersionDetails version={mockVersion} diff={mockDiff} />
  </PanelStory>
)

export const DiffLoading = () => (
  <PanelStory>
    <InstallConfigVersionDetails version={mockVersion} isDiffLoading />
  </PanelStory>
)

export const NoDiff = () => (
  <PanelStory>
    <InstallConfigVersionDetails version={mockVersion} />
  </PanelStory>
)
