import { InstallConfigsTimeline } from './InstallConfigsTimeline'
import type { TInstallConfigVersion } from '@/types'

export default {
  title: 'InstallConfigs/InstallConfigsTimeline',
}

const day = 86400000

const mockVersions = [
  {
    id: 'icv-3',
    created_at: new Date(Date.now() - day).toISOString(),
    created: true,
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
  },
  {
    id: 'icv-2',
    created_at: new Date(Date.now() - day * 4).toISOString(),
    created: false,
    file_path: 'installs/payments.toml',
    status: { status: 'error' },
    install_config_sync: { triggered_by: 'manual' },
  },
] as unknown as TInstallConfigVersion[]

export const Default = () => (
  <InstallConfigsTimeline
    versions={mockVersions}
    orgId="org-1"
    installId="inst-1"
  />
)

export const Empty = () => <InstallConfigsTimeline versions={[]} />

export const Loading = () => <InstallConfigsTimeline versions={[]} isLoading />
