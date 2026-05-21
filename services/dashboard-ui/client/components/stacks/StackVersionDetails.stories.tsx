export default {
  title: 'Stacks/StackVersionDetails',
}

import { PanelStory } from '@/components/__stories__/helpers'
import { StackVersionDetails } from './StackVersionDetails'
import type { TInstallStack } from '@/types'

type TStackVersion = TInstallStack['versions'][number]

const mockVersion: TStackVersion = {
  id: 'sv-1',
  created_at: new Date(Date.now() - 3600000).toISOString(),
  updated_at: new Date(Date.now() - 1800000).toISOString(),
  quick_link_url: 'https://console.aws.amazon.com/cloudformation/home?#/stacks/create/review?templateURL=https://example.com/template.json',
  template_url: 'https://s3.amazonaws.com/nuon-stacks/template.json',
  app_config_id: 'config-1',
  aws_bucket_key: 'stacks/sv-1/template.json',
  aws_bucket_name: 'nuon-stacks-bucket',
  composite_status: {
    status: 'active',
    history: [
      { status: 'pending', created_at_ts: Date.now() - 7200000 },
      { status: 'active', created_at_ts: Date.now() - 3600000 },
    ],
  },
  runs: [
    {
      id: 'run-1',
      created_at: new Date(Date.now() - 3600000).toISOString(),
      data: { vpc_id: 'vpc-abc123', cluster_name: 'nuon-cluster' },
      data_contents: {},
    },
  ],
  contents: btoa(JSON.stringify({ template: 'cloudformation', version: '1.0' })),
} as TStackVersion

export const Default = () => (
  <PanelStory>
    <StackVersionDetails version={mockVersion} />
  </PanelStory>
)

export const Expired = () => (
  <PanelStory>
    <StackVersionDetails
      version={{ ...mockVersion, composite_status: { ...mockVersion.composite_status, status: 'expired' } }}
    />
  </PanelStory>
)
