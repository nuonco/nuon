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
  capabilities: ['stacks', 'images'],
  status: 'pending',
  status_message: '',
  setup: {
    issuer_url: 'https://api.nuon.co',
    subject: 'org:org_01JEXAMPLE:connection:cc_01JEXAMPLE',
    audience: 'sts.amazonaws.com',
    trust_policy: {},
    terraform:
      'module "nuon_cloud_connection" {\n  source = "nuonco/ecr-access/aws"\n}',
    cli: 'aws iam create-open-id-connect-provider --url https://api.nuon.co',
    cloudformation: 'NuonConnectionRole:\n  Type: AWS::IAM::Role',
    capabilities: ['stacks', 'images'],
  },
  used_by: { installs: 0, components: 0 },
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
