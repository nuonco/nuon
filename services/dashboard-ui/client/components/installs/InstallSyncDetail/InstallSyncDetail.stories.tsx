export default {
  title: 'Installs/InstallSyncDetail',
}

import type { TAppInstallConfigSync } from '@/types'
import { InstallSyncDetail } from './InstallSyncDetail'

const baseSync = {
  id: 'syncabc123',
  triggered_by: 'git-push',
  created_at: '2026-08-24T14:30:00Z',
  status: {
    status: 'active',
    status_human_description: 'Sync completed successfully',
  },
  vcs_connection_commit: {
    sha: 'a1b2c3d4e5f6a7b8c9d0',
    message: 'Update replica count and add ingress config\n\nMore detail here',
    author_name: 'acme-engineer',
  },
  queue_id: 'queue123',
  queue_signal_id: 'signal123',
  workflow_id: 'wf123',
} as unknown as TAppInstallConfigSync

const withSteps = {
  ...baseSync,
  workflow: {
    id: 'wf123',
    steps: [
      {
        id: 'step1',
        name: 'Resolve app config',
        execution_type: 'default',
        group_idx: 0,
        status: {
          status: 'success',
          status_human_description: 'Resolved 3 components',
        },
      },
      {
        id: 'step2',
        name: 'Sync installs',
        execution_type: 'default',
        group_idx: 1,
        status: {
          status: 'active',
          metadata: { install_count: 4, region: 'us-east-1' },
        },
      },
    ],
  },
} as unknown as TAppInstallConfigSync

const withInstallSyncs = {
  ...baseSync,
  install_config_syncs: [
    {
      id: 'ics1',
      install_id: 'installaaa111',
      status: { status: 'active', status_human_description: 'Up to date' },
    },
    {
      id: 'ics2',
      install_id: 'installbbb222',
      status: { status: 'error', status_human_description: 'Plan failed' },
    },
  ],
} as unknown as TAppInstallConfigSync

const noCommit = {
  ...baseSync,
  vcs_connection_commit: undefined,
  status: { status: 'queued', status_human_description: 'Waiting to start' },
} as unknown as TAppInstallConfigSync

export const Default = () => (
  <div className="flex flex-col gap-4 max-w-4xl">
    <InstallSyncDetail
      sync={baseSync}
      orgId="orgabc"
      appId="appabc"
      syncId="syncabc123"
    />
  </div>
)

export const WithSteps = () => (
  <div className="flex flex-col gap-4 max-w-4xl">
    <InstallSyncDetail
      sync={withSteps}
      orgId="orgabc"
      appId="appabc"
      syncId="syncabc123"
    />
  </div>
)

export const WithInstallSyncs = () => (
  <div className="flex flex-col gap-4 max-w-4xl">
    <InstallSyncDetail
      sync={withInstallSyncs}
      orgId="orgabc"
      appId="appabc"
      syncId="syncabc123"
    />
  </div>
)

export const NoCommitInfo = () => (
  <div className="flex flex-col gap-4 max-w-4xl">
    <InstallSyncDetail
      sync={noCommit}
      orgId="orgabc"
      appId="appabc"
      syncId="syncabc123"
    />
  </div>
)
