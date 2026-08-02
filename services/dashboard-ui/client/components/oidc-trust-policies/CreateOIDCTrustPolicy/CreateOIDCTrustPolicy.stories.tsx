export default {
  title: 'OIDCTrustPolicies/CreateOIDCTrustPolicy',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { CreateOIDCTrustPolicyModal } from './CreateOIDCTrustPolicy'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <CreateOIDCTrustPolicyModal
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <CreateOIDCTrustPolicyModal isPending={true} error={null} onSubmit={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <CreateOIDCTrustPolicyModal
      isPending={false}
      error={{
        error: 'claim_conditions must include a "sub" condition',
        description: '',
        user_error: true,
        status: 400,
      }}
      onSubmit={noop}
    />
  </ModalStory>
)
