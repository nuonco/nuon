export default { title: 'CloudConnections/CreateCloudConnection' }

import { ModalStory } from '@/components/__stories__/helpers'
import type { TCloudConnection } from '@/types'
import { CreateCloudConnectionModal } from './CreateCloudConnection'

const noop = () => {}

const connection = {
  id: 'cc_01JEXAMPLE',
  name: 'Production AWS',
  platform: 'aws',
  target_id: '123456789012',
  principal: 'arn:aws:iam::123456789012:role/nuon-cloud-connection',
  preset: 'stacks',
  status: 'pending',
  status_message: '',
  setup: {
    issuer_url: 'https://api.example.com',
    subject: 'org:org_01JEXAMPLE:connection:cc_01JEXAMPLE',
    audience: 'sts.amazonaws.com',
    trust_policy: {},
    terraform:
      'resource "aws_iam_role" "nuon_cloud_connection" {\n  name = "acme-connection"\n}',
    cli: 'aws iam create-open-id-connect-provider --url https://api.example.com',
    cloudformation: 'NuonConnectionRole:\n  Type: AWS::IAM::Role',
    preset: 'stacks',
  },
  used_by: { installs: 0 },
} as TCloudConnection

export const Default = () => (
  <ModalStory>
    <CreateCloudConnectionModal
      connection={null}
      error={null}
      isPending={false}
      isVerifying={false}
      verifyError={null}
      onSubmit={noop}
      onVerify={noop}
      onDone={noop}
    />
  </ModalStory>
)

export const Setup = () => (
  <ModalStory>
    <CreateCloudConnectionModal
      connection={connection}
      error={null}
      isPending={false}
      isVerifying={false}
      verifyError={null}
      onSubmit={noop}
      onVerify={noop}
      onDone={noop}
    />
  </ModalStory>
)
