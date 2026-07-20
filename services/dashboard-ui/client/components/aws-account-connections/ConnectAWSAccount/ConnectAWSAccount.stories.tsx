export default { title: 'AWS account connections/ConnectAWSAccount' }

import { ModalStory } from '@/components/__stories__/helpers'
import { ConnectAWSAccountModal } from './ConnectAWSAccount'

const noop = () => {}

export const Create = () => (
  <ModalStory>
    <ConnectAWSAccountModal
      isPending={false}
      onCreate={noop}
      onSetRole={noop}
      onVerify={noop}
    />
  </ModalStory>
)

export const ConfigureTrust = () => (
  <ModalStory>
    <ConnectAWSAccountModal
      connection={{
        id: 'awc123',
        created_at: '2026-07-16T00:00:00Z',
        updated_at: '2026-07-16T00:00:00Z',
        name: 'Demo smoke tests',
        account_id: '123456789012',
        default_region: 'us-west-2',
        verification_status: 'pending',
        external_id: 'example-external-id',
        trust_policy: { Version: '2012-10-17', Statement: [] },
      }}
      isPending={false}
      onCreate={noop}
      onSetRole={noop}
      onVerify={noop}
    />
  </ModalStory>
)
