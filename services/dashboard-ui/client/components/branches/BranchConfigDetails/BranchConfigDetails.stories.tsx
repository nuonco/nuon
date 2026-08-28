import { PanelStory } from '@/components/__stories__/helpers'
import { BranchConfigDetails } from './BranchConfigDetails'
import type { TAppConfig } from '@/types'

export default {
  title: 'Branches/BranchConfigDetails',
}

const mockConfig = {
  id: 'cfg-3',
  version: 3,
  created_at: new Date(Date.now() - 3600000).toISOString(),
  status_v2: { status: 'active' },
  cli_version: '0.19.2',
  checksum: 'a1b2c3d4e5f67890',
  component_ids: ['cmp-1', 'cmp-2'],
  action_ids: ['act-1'],
  runbook_ids: [],
  vcs_connection_commit: {
    sha: 'a1b2c3d4e5f6',
    message: 'Add a cache component to the deployment plan',
    author_name: 'Jane Doe',
  },
} as unknown as TAppConfig

const mockFullConfig = {
  ...mockConfig,
  stack: {
    type: 'aws-eks',
    name: 'acme-eks',
  },
  runner: {
    app_runner_type: 'aws-eks',
  },
  kubernetes_contexts: {
    contexts: [{ name: 'primary', namespace: 'payments' }],
  },
} as unknown as TAppConfig

export const Default = () => (
  <PanelStory>
    <BranchConfigDetails config={mockConfig} fullConfig={mockFullConfig} />
  </PanelStory>
)

export const ContentsLoading = () => (
  <PanelStory>
    <BranchConfigDetails config={mockConfig} isLoading />
  </PanelStory>
)

export const CliSynced = () => (
  <PanelStory>
    <BranchConfigDetails
      config={
        {
          ...mockConfig,
          vcs_connection_commit: undefined,
        } as unknown as TAppConfig
      }
      fullConfig={mockFullConfig}
    />
  </PanelStory>
)
