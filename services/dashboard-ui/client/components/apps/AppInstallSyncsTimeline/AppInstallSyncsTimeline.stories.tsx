import { AppInstallSyncsTimeline } from './AppInstallSyncsTimeline'
import type { TAppInstallConfigSync } from '@/types'

export default {
  title: 'Apps/AppInstallSyncsTimeline',
}

const day = 86400000

const mockSyncs = [
  {
    id: 'ais-3',
    created_at: new Date(Date.now() - day).toISOString(),
    triggered_by: 'vcs-webhook',
    status: {
      status: 'awaiting_approval',
      status_human_description: 'Waiting on install creation approval',
    },
    vcs_connection_commit: { sha: 'abc1234567890' },
    install_creation_approval: {
      id: 'approval-1',
      status: 'pending',
      proposed_installs: [
        { name: 'acme-west', file_path: 'installs/acme-west.toml', config: {} },
        { name: 'acme-east', file_path: 'installs/acme-east.toml', config: {} },
      ],
    },
    install_config_syncs: [
      { id: 'ics-1', install_id: 'inst-1', status: { status: 'active' } },
    ],
  },
  {
    id: 'ais-2',
    created_at: new Date(Date.now() - day * 5).toISOString(),
    triggered_by: 'manual',
    status: { status: 'active' },
    vcs_connection_commit: { sha: 'def4567890123' },
    install_config_syncs: [
      { id: 'ics-2', install_id: 'inst-1', status: { status: 'active' } },
      { id: 'ics-3', install_id: 'inst-2', status: { status: 'error' } },
    ],
    queue_id: 'queue-1',
    queue_signal_id: 'signal-1',
  },
] as unknown as TAppInstallConfigSync[]

export const Default = () => (
  <AppInstallSyncsTimeline syncs={mockSyncs} orgId="org-1" appId="app-1" />
)

export const Empty = () => <AppInstallSyncsTimeline syncs={[]} />

export const Loading = () => <AppInstallSyncsTimeline syncs={[]} isLoading />
